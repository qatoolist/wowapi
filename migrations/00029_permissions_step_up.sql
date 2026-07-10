-- GAP-004: step-up/MFA seedability. The authz evaluator and HTTP gate already
-- enforce Permission.StepUp (kernel/authz/evaluator.go, kernel/httpx/authz_gate.go)
-- and seeds.PermissionSeed now carries a step_up field, but the permissions
-- catalog itself has no column to persist it (00006_authz.sql). Without this,
-- a product cannot declare `step_up: true` in seed YAML and have it survive a
-- reseed — it would have to register the permission out-of-band, exactly the
-- workaround this gap closes. Defaults to false so every existing catalog row
-- is unaffected until a seed opts in.

-- +goose Up

ALTER TABLE permissions ADD COLUMN step_up boolean NOT NULL DEFAULT false;

-- +goose Down

ALTER TABLE permissions DROP COLUMN step_up;
