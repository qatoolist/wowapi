package httpx_test

// route_contract_test.go — W01-E03-S002 (FBL-08 / MATRIX CS-08): boot-time
// request-contract enforcement for mutating routes at the RouteMeta seam.
//
// The framework's validation library (kernel/validation + BindAndValidate) is
// opt-in per handler: a POST handler that never calls BindAndValidate gets
// ZERO validation and nothing detects it. These tests pin both halves of the
// fix: an undeclared-contract mutating route always fails registration through
// the existing Router.Err() accumulation path, while declared and explicitly
// body-less mutations remain valid.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/kernel/validation"
)

func contractNoop(http.ResponseWriter, *http.Request) {}

// createWidgetRequest is the fixture request contract for the enforcement
// and adversarial tests below.
type createWidgetRequest struct {
	Name string `json:"name" validate:"required"`
}

// TestRouterRequireRequestContractsRejectsUndeclaredMutatingRoute is
// AC-W01-E03-S002-02: a POST route that declares
// neither a Request contract nor the NoRequestBody waiver fails registration
// with an error naming the route and the missing contract, surfaced through
// the existing Router.Err() accumulation path.
func TestRouterRequireRequestContractsRejectsUndeclaredMutatingRoute(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch} {
		r := httpx.NewRouter()
		r.Handle(method, "/things", httpx.RouteMeta{Permission: "things.write"}, contractNoop)
		err := r.Err()
		if err == nil {
			t.Fatalf("%s route without a declared contract must fail registration", method)
		}
		for _, want := range []string{method, "/things", "request contract"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("%s: error must mention %q:\n%v", method, want, err)
			}
		}
	}
}

// TestRouterRequireRequestContractsAllowsDeclaredContract: declaring the
// contract satisfies the check — the enforcement rejects only the
// undeclared case, not mutating routes as such.
func TestRouterRequireRequestContractsAllowsDeclaredContract(t *testing.T) {
	r := httpx.NewRouter()
	r.Handle(http.MethodPost, "/widgets", httpx.RouteMeta{
		Permission: "widgets.create",
		Request:    createWidgetRequest{},
	}, contractNoop)
	if err := r.Err(); err != nil {
		t.Fatalf("POST route with a declared Request contract must register cleanly, got: %v", err)
	}
}

// TestRouterRequireRequestContractsWaiverExemptsBodylessMutation is
// AC-W01-E03-S002-04: NoRequestBody waives the requirement for a genuinely
// body-less mutation, and the waived route still boots with enforcement on.
func TestRouterRequireRequestContractsWaiverExemptsBodylessMutation(t *testing.T) {
	r := httpx.NewRouter()
	r.Handle(http.MethodPost, "/things/{id}/archive", httpx.RouteMeta{
		Permission:    "things.archive",
		NoRequestBody: true,
	}, contractNoop)
	if err := r.Err(); err != nil {
		t.Fatalf("NoRequestBody waiver must exempt a body-less mutation from the contract check, got: %v", err)
	}
}

// TestRouterRejectsContractWithWaiverContradiction: Request and NoRequestBody
// are mutually exclusive — declaring both is a registration error regardless
// of the enforcement flag, mirroring the Public/Permission contradiction rule.
func TestRouterRejectsContractWithWaiverContradiction(t *testing.T) {
	r := httpx.NewRouter() // enforcement deliberately OFF: the contradiction is invalid everywhere
	r.Handle(http.MethodPost, "/widgets", httpx.RouteMeta{
		Permission:    "widgets.create",
		Request:       createWidgetRequest{},
		NoRequestBody: true,
	}, contractNoop)
	if r.Err() == nil {
		t.Fatal("Request + NoRequestBody is contradictory and must fail registration even with enforcement off")
	}
}

// TestRouterRequireRequestContractsIgnoresNonMutatingMethods: the check is
// scoped to POST/PUT/PATCH — reads and body-less-by-spec verbs (GET, DELETE)
// never need a contract even with enforcement on.
func TestRouterRequireRequestContractsIgnoresNonMutatingMethods(t *testing.T) {
	r := httpx.NewRouter()
	r.Handle(http.MethodGet, "/things", httpx.RouteMeta{Permission: "things.list"}, contractNoop)
	r.Handle(http.MethodDelete, "/things/{id}", httpx.RouteMeta{Permission: "things.delete"}, contractNoop)
	if err := r.Err(); err != nil {
		t.Fatalf("GET/DELETE must not require a request contract, got: %v", err)
	}
}

// TestValidatedHandlerRejectsInvalidDTOWith400FieldErrors is the story's
// headline adversarial proof (AC-W01-E03-S002-03): a route built through the
// adaptor with a declared RouteMeta.Request contract answers an invalid DTO
// with HTTP 400 carrying field errors (the existing KindValidation problem-
// details shape), and the business handler never runs.
func TestValidatedHandlerRejectsInvalidDTOWith400FieldErrors(t *testing.T) {
	v := validation.New()
	r := httpx.NewRouter()

	handlerRan := false
	r.Handle(http.MethodPost, "/widgets", httpx.RouteMeta{
		Permission: "widgets.create",
		Request:    createWidgetRequest{},
	}, httpx.ValidatedHandler(v, 1<<20, func(w http.ResponseWriter, _ *http.Request, in createWidgetRequest) {
		handlerRan = true
		httpx.WriteJSON(w, http.StatusCreated, httpx.OK(in))
	}))
	if err := r.Err(); err != nil {
		t.Fatalf("declaring route must register cleanly with enforcement on: %v", err)
	}

	// Adversarial: violates the `validate:"required"` tag on name.
	req := httptest.NewRequest(http.MethodPost, "/widgets", strings.NewReader(`{"name":""}`))
	rec := httptest.NewRecorder()
	r.Routes()[0].Handler(rec, req)

	if handlerRan {
		t.Fatal("business handler must not run for an invalid DTO")
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	var p struct {
		Code   string `json:"code"`
		Errors []struct {
			Field string `json:"field"`
			Code  string `json:"code"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("400 body is not valid JSON: %v\n%s", err, rec.Body.String())
	}
	if len(p.Errors) == 0 {
		t.Fatalf("400 body must carry field errors, got: %s", rec.Body.String())
	}
	if p.Errors[0].Field != "name" {
		t.Errorf("field error path = %q, want %q (body: %s)", p.Errors[0].Field, "name", rec.Body.String())
	}
}

// TestValidatedHandlerPassesValidDTOToBusinessLogic: the happy path — a valid
// body binds, validates, and reaches the business handler as a typed value.
func TestValidatedHandlerPassesValidDTOToBusinessLogic(t *testing.T) {
	v := validation.New()
	var got createWidgetRequest
	h := httpx.ValidatedHandler(v, 1<<20, func(w http.ResponseWriter, _ *http.Request, in createWidgetRequest) {
		got = in
		httpx.WriteJSON(w, http.StatusCreated, httpx.OK(in))
	})

	req := httptest.NewRequest(http.MethodPost, "/widgets", strings.NewReader(`{"name":"ok"}`))
	rec := httptest.NewRecorder()
	h(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (body: %s)", rec.Code, rec.Body.String())
	}
	if got.Name != "ok" {
		t.Errorf("bound value = %+v, want Name=ok", got)
	}
}
