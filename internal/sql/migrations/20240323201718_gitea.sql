-- +goose Up

INSERT INTO vcs_kinds (name) VALUES
	('gitea');

-- +goose Down

DELETE FROM vcs_kinds WHERE name = 'gitea';