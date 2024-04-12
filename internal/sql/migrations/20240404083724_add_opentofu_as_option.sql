-- +goose Up
-- +goose StatementBegin

CREATE TYPE ENGINE_KINDS AS ENUM('terraform', 'opentofu');

CREATE TYPE ENGINE AS (kind engine_kinds, version text);

ALTER TABLE workspaces
ADD COLUMN engine engine;

UPDATE workspaces
SET engine = ('terraform', workspaces.terraform_version);

ALTER TABLE workspaces 
DROP COLUMN terraform_version;

ALTER TABLE workspaces
ALTER COLUMN engine SET NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE workspaces
ADD COLUMN terraform_version TEXT;

UPDATE workspaces
SET terraform_version = (workspaces.engine).version
WHERE (engine).kind = 'terraform';

-- falling back to the most reasonable option. We shouldn't ever really need 
-- to do the down migration so this should be okay.
UPDATE workspaces
SET terraform_version = '1.7.0'
WHERE (engine).kind = 'opentofu';

ALTER TABLE workspaces
DROP COLUMN engine;

ALTER TABLE workspaces
ALTER COLUMN terraform_version SET NOT NULL;

DROP TYPE ENGINE;

DROP TYPE ENGINE_KINDS;

-- +goose StatementEnd
