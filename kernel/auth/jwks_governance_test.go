package auth_test

import (
	"net/http"
	"testing"

	"github.com/qatoolist/wowapi/v2/kernel/auth"
	"github.com/qatoolist/wowapi/v2/kernel/config"
)

func TestNewJWKSKeySource_ProdCustomClientRequiresTrustedIssuers(t *testing.T) {
	_, err := auth.NewJWKSKeySource(auth.JWKSConfig{
		Issuer:         "https://idp.example.com",
		JWKSURI:        "https://idp.example.com/jwks",
		Client:         &http.Client{Timeout: 5},
		TrustedIssuers: nil,
		Env:            config.EnvProd,
	})
	if err == nil {
		t.Fatal("expected prod profile + custom JWKS client + no trusted issuers to fail")
	}
}

func TestNewJWKSKeySource_ProdCustomClientWithTrustedIssuersAllowed(t *testing.T) {
	_, err := auth.NewJWKSKeySource(auth.JWKSConfig{
		Issuer:         "https://idp.example.com",
		JWKSURI:        "https://idp.example.com/jwks",
		Client:         &http.Client{Timeout: 5},
		TrustedIssuers: []string{"https://idp.example.com"},
		Env:            config.EnvProd,
	})
	if err != nil {
		t.Fatalf("expected prod profile + custom JWKS client + trusted issuers to succeed: %v", err)
	}
}

func TestNewJWKSKeySource_NonProdCustomClientWithoutTrustedIssuersAllowed(t *testing.T) {
	_, err := auth.NewJWKSKeySource(auth.JWKSConfig{
		Issuer:         "https://idp.example.com",
		JWKSURI:        "https://idp.example.com/jwks",
		Client:         &http.Client{Timeout: 5},
		TrustedIssuers: nil,
		Env:            config.EnvStage,
	})
	if err != nil {
		t.Fatalf("expected non-prod profile to allow custom JWKS client without trusted issuers: %v", err)
	}
}

func TestNewJWKSKeySource_DefaultClientInProdAllowedWithoutTrustedIssuers(t *testing.T) {
	_, err := auth.NewJWKSKeySource(auth.JWKSConfig{
		Issuer:         "https://idp.example.com",
		JWKSURI:        "https://idp.example.com/jwks",
		TrustedIssuers: nil,
		Env:            config.EnvProd,
	})
	if err != nil {
		t.Fatalf("expected default JWKS client in prod to succeed without trusted issuers: %v", err)
	}
}
