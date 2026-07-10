-- Scoped privileged framework services (GAP-006). The kernel/privileged package
-- exposes narrowly-scoped, audited services (relationship-edge grant/revoke,
-- tenant-scope rule-version activation) that modules reach through
-- module.Context — replacing the per-product SECURITY DEFINER bridges a product
-- otherwise had to author (SEC-24 / SEC-13). Those services run in a TENANT-BOUND
-- app_platform transaction: platform privilege for the protected writes, but with
-- app.tenant_id set so every table's RLS still isolates the tenant.
--
-- To run as app_platform (instead of an owner-privileged SECURITY DEFINER
-- function) the role needs a few grants it did not previously hold. Each is the
-- MINIMUM the services require and each stays tenant-isolated by the existing
-- app_tenant_id() RLS policies (which are role-unqualified, so they already bind
-- app_platform):
--
--   * acting_capacities — SELECT only: the relationship-grant path checks the
--     subject capacity exists and is active in the bound tenant (absorbing the
--     bridge's identity_committee_seat_capacity_invalid check). Read-only; the
--     module still cannot write the identity graph.
--   * audit_logs — SELECT, INSERT: the services write an audit row in the same
--     tx as the privileged write. INSERT+SELECT only preserves the append-only
--     invariant (00017) for app_platform exactly as for app_rt — no UPDATE/DELETE.
--   * audit_chain — SELECT, INSERT, UPDATE: the per-tenant hash-chain head is
--     advanced with each audit row (00018). 00027 already granted app_platform
--     SELECT here (cross-tenant anchor read); this adds the INSERT/UPDATE the
--     tenant-bound audit write needs. The chain's tenant-isolation policy still
--     scopes these writes to the bound tenant.
--
-- Deliberately NOT changed: the `relationships` and `rule_versions` grants
-- (00005 / 00008) are untouched — app_platform already holds exactly the writes
-- the services need there, and app_rt keeps its SELECT-only / propose-only
-- posture. No new privilege reaches the module runtime role.

-- +goose Up

GRANT SELECT ON acting_capacities TO app_platform;
GRANT SELECT, INSERT ON audit_logs TO app_platform;
GRANT INSERT, UPDATE ON audit_chain TO app_platform; -- SELECT already granted in 00027

-- +goose Down

REVOKE INSERT, UPDATE ON audit_chain FROM app_platform;
REVOKE SELECT, INSERT ON audit_logs FROM app_platform;
REVOKE SELECT ON acting_capacities FROM app_platform;
