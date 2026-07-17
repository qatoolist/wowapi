// Pgx query tracing (FBL-06 T3, decision D-08 / ADR-W00-E02-S003-008): a thin,
// hand-rolled pgx.QueryTracer over the framework's own observability port
// (tracing.Tracer, aliased as observability.Tracer) — NOT otelpgx, which
// would bind OTel vendor types into this kernel package. One span per query,
// parented under whatever span is active in the
// query's context; sampling therefore inherits the parent span's decision
// (the otel adapter's ParentBased sampler), with no independent choice here.

package database

import (
	"context"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qatoolist/wowapi/v2/kernel/tracing"
)

// WithQueryTracer makes every query on the pool (Query/QueryRow/Exec) open a
// span named after the query's leading SQL keyword ("db.SELECT", "db.INSERT",
// …; "db.query" for anything unrecognized — bounded cardinality by
// construction). The span carries a db.statement attr (the SQL text as issued,
// whitespace-trimmed and truncated — parameterized queries contain $n
// placeholders, so bound values never appear) and, on success, a
// db.rows_affected attr; a failed query is marked errored via RecordError.
//
// Opt in at the composition root, mirroring outbox.WithRelayTracer /
// jobs.WithRunnerTracer:
//
//	pool, err := database.NewPool(ctx, dsn, cfg.DB, database.WithQueryTracer(tr))
//
// A nil or NoOpTracer leaves the pool config untouched, preserving the
// documented zero-cost disabled-tracing path.
//
// Security: bound parameter values (TraceQueryStartData.Args) are never
// attached to the span. SQL assembled by string concatenation would surface
// its embedded literals in db.statement — parameterize instead, exactly as the
// rest of this package already requires.
func WithQueryTracer(tr tracing.Tracer) Option {
	return func(pc *pgxpool.Config) {
		if tr == nil || tr == tracing.NoOpTracer {
			return
		}
		pc.ConnConfig.Tracer = queryTracer{tr: tr}
	}
}

// queryTracer implements pgx.QueryTracer over the observability.Tracer port.
// It is stateless: the per-query Span travels in the context pgx hands back to
// TraceQueryEnd, so concurrent queries across pooled connections never share
// mutable state.
type queryTracer struct{ tr tracing.Tracer }

// querySpanKey carries the in-flight query's Span from TraceQueryStart to the
// matching TraceQueryEnd (pgx passes the returned context through).
type querySpanKey struct{}

func (q queryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	ctx, span := q.tr.StartSpan(ctx, "db."+sqlVerb(data.SQL))
	span.SetAttr("db.statement", statementSummary(data.SQL))
	return context.WithValue(ctx, querySpanKey{}, span)
}

func (q queryTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	span, ok := ctx.Value(querySpanKey{}).(tracing.Span)
	if !ok {
		return
	}
	if data.Err != nil {
		span.RecordError(data.Err)
	} else {
		span.SetAttr("db.rows_affected", strconv.FormatInt(data.CommandTag.RowsAffected(), 10))
	}
	span.End()
}

// Compile-time assurance the tracer satisfies pgx's interface.
var _ pgx.QueryTracer = queryTracer{}

// sqlVerbs is the closed set of leading keywords used for span naming; the
// closed set (not the raw first token) is what bounds span-name cardinality.
var sqlVerbs = map[string]struct{}{
	"SELECT": {}, "INSERT": {}, "UPDATE": {}, "DELETE": {}, "WITH": {},
	"BEGIN": {}, "COMMIT": {}, "ROLLBACK": {}, "SAVEPOINT": {}, "RELEASE": {},
	"CREATE": {}, "ALTER": {}, "DROP": {}, "TRUNCATE": {}, "GRANT": {}, "REVOKE": {},
	"CALL": {}, "DO": {}, "SET": {}, "SHOW": {}, "COPY": {}, "EXPLAIN": {},
	"VACUUM": {}, "ANALYZE": {}, "LISTEN": {}, "NOTIFY": {}, "REFRESH": {},
}

// sqlVerb returns the query's leading keyword uppercased when it is a known
// SQL verb, else "query".
func sqlVerb(sql string) string {
	s := strings.TrimSpace(sql)
	end := len(s)
	for i := range len(s) {
		if c := s[i]; c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '(' || c == ';' {
			end = i
			break
		}
	}
	verb := strings.ToUpper(s[:end])
	if _, ok := sqlVerbs[verb]; ok {
		return verb
	}
	return "query"
}

// maxStatementAttrLen bounds the db.statement attr size; long SQL is truncated
// on a rune boundary with a trailing ellipsis.
const maxStatementAttrLen = 512

func statementSummary(sql string) string {
	s := strings.TrimSpace(sql)
	if len(s) <= maxStatementAttrLen {
		return s
	}
	cut := maxStatementAttrLen
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut] + "…"
}
