package sql

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhook_Sync_Create(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	repo := cloud.NewTestRepo()
	want := otf.NewTestWebhook(repo, github.Defaults())

	createFunc := func(context.Context, otf.WebhookCreatorOptions) (*otf.Webhook, error) {
		return want, nil
	}

	got, err := db.SyncWebhook(ctx, otf.SyncWebhookOptions{
		CreateWebhookFunc: createFunc,
		Identifier:        want.Identifier,
		Cloud:             want.CloudName(),
	})
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestWebhook_Sync_Update(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	webhook := createTestWebhook(t, db)

	updateFunc := func(context.Context, otf.WebhookUpdaterOptions) (string, error) {
		return "updated-vcs-id", nil
	}
	opts := otf.SyncWebhookOptions{
		UpdateWebhookFunc: updateFunc,
		Identifier:        webhook.Identifier,
		Cloud:             webhook.CloudName(),
	}

	got, err := db.SyncWebhook(ctx, opts)
	require.NoError(t, err)
	assert.Equal(t, "updated-vcs-id", got.VCSID)
}

func TestWebhook_Sync_NoChange(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	want := createTestWebhook(t, db)

	updateFunc := func(context.Context, otf.WebhookUpdaterOptions) (string, error) {
		return want.VCSID, nil
	}
	opts := otf.SyncWebhookOptions{
		UpdateWebhookFunc: updateFunc,
		Identifier:        want.Identifier,
		Cloud:             want.CloudName(),
	}

	got, err := db.SyncWebhook(ctx, opts)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestWebhook_Delete(t *testing.T) {
	db := newTestDB(t)
	hook := createTestWebhook(t, db)

	err := db.DeleteWebhook(context.Background(), hook.WebhookID)
	require.NoError(t, err)
}