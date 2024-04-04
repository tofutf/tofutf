-- name: InsertGPGKey :exec
INSERT INTO registry_gpg_keys (
    id,
    organization_name,
    ascii_armor,
    key_id,
    created_at,
    updated_at
) VALUES (
    pggen.arg('id'),
    pggen.arg('organization_name'),
    pggen.arg('ascii_armor'),
    pggen.arg('key_id'),
    pggen.arg('created_at'),
    pggen.arg('updated_at')
);

-- name: UpdateGPGKey :exec
UPDATE registry_gpg_keys
SET organization_name = pggen.arg('new_organization_name'), 
    updated_at = pggen.arg('updated_at')
WHERE key_id = pggen.arg('key_id') AND 
    organization_name = pggen.arg('organization_name');

-- name: DeleteGPGKey :exec
DELETE FROM registry_gpg_keys
WHERE key_id = pggen.arg('key_id') AND 
    organization_name = pggen.arg('organization_name');

-- name: ListGPGKeys :many
SELECT *
FROM registry_gpg_keys
WHERE organization_name = ANY(pggen.arg('organization_names')::TEXT[]);

-- name: GetGPGKey :one
SELECT *
FROM registry_gpg_keys
WHERE key_id = pggen.arg('key_id') AND 
    organization_name = pggen.arg('organization_name');