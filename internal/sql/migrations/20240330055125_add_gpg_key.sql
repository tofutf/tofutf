-- +goose Up
-- +goose StatementBegin

CREATE TABLE registry_gpg_keys (
    id text PRIMARY KEY,
    organization_name text REFERENCES organizations(name) NOT NULL,
    ascii_armor text NOT NULL,
    key_id text NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    UNIQUE(organization_name, key_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE registry_gpg_keys;

-- +goose StatementEnd
