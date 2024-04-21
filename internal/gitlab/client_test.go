package gitlab

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/testutils"
	"github.com/tofutf/tofutf/internal/vcs"
)

func TestClient_GetUser(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("GET /api/v4/user", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"username":"bobby"}`)
	})

	got, err := client.GetCurrentUser(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "bobby", got)
}

func TestClient_GetRepository(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("GET /api/v4/projects/{path_with_namespace}", func(w http.ResponseWriter, r *http.Request) {
		pathWithNamespace := r.PathValue("path_with_namespace")
		require.Equal(t, pathWithNamespace, "acme/terraform")

		fmt.Fprint(w, `{"path_with_namespace":"acme/terraform","default_branch":"master"}`)
	})

	got, err := client.GetRepository(context.Background(), "acme/terraform")
	require.NoError(t, err)

	assert.Equal(t, "acme/terraform", got.Path)
	assert.Equal(t, "master", got.DefaultBranch)
}

func TestClient_ListRepositories(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("GET /api/v4/projects", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{"path_with_namespace":"acme/terraform"}]`)
	})

	got, err := client.ListRepositories(context.Background(), vcs.ListRepositoriesOptions{})
	require.NoError(t, err)

	assert.Equal(t, []string{"acme/terraform"}, got)
}

func TestClient_GetRepoTarball(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("GET /api/v4/projects/{path_with_namespace}/repository/archive.tar.gz", func(w http.ResponseWriter, r *http.Request) {
		pathWithNamespace := r.PathValue("path_with_namespace")
		require.Equal(t, pathWithNamespace, "acme/terraform")

		w.Write(testutils.ReadFile(t, "../testdata/gitlab.tar.gz")) //nolint:errcheck
	})

	got, ref, err := client.GetRepoTarball(context.Background(), vcs.GetRepoTarballOptions{
		Repo: "acme/terraform",
	})
	require.NoError(t, err)
	assert.Equal(t, "0335fb07bb0244b7a169ee89d15c7703e4aaf7de", ref)

	dst := t.TempDir()
	err = internal.Unpack(bytes.NewReader(got), dst)
	require.NoError(t, err)
	assert.FileExists(t, path.Join(dst, "afile"))
	assert.FileExists(t, path.Join(dst, "bfile"))
}

func TestClient_GetRepoTarballSubGroup(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("GET /api/v4/projects/{path_with_subgroup_and_namespace}/repository/archive.tar.gz", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		w.Write(testutils.ReadFile(t, "../testdata/gitlab.tar.gz")) //nolint:errcheck
	})

	got, ref, err := client.GetRepoTarball(context.Background(), vcs.GetRepoTarballOptions{
		Repo: "acme/subgroup/terraform",
	})
	require.NoError(t, err)
	assert.Equal(t, "0335fb07bb0244b7a169ee89d15c7703e4aaf7de", ref)

	dst := t.TempDir()
	err = internal.Unpack(bytes.NewReader(got), dst)
	require.NoError(t, err)
	assert.FileExists(t, path.Join(dst, "afile"))
	assert.FileExists(t, path.Join(dst, "bfile"))
}

func TestClient_CreateWebhook(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("POST /api/v4/projects/{path_with_namespace}/hooks", func(w http.ResponseWriter, r *http.Request) {
		pathWithNamespace := r.PathValue("path_with_namespace")
		require.Equal(t, pathWithNamespace, "acme/terraform")

		fmt.Fprint(w, `{"id":1}`)
	})

	got, err := client.CreateWebhook(context.Background(), vcs.CreateWebhookOptions{
		Repo: "acme/terraform",
	})
	require.NoError(t, err)
	assert.Equal(t, "1", got)
}

func TestClient_UpdateWebhook(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("PUT /api/v4/projects/{path_with_namespace}/hooks/1", func(w http.ResponseWriter, r *http.Request) {
		pathWithNamespace := r.PathValue("path_with_namespace")
		require.Equal(t, pathWithNamespace, "acme/terraform")

		fmt.Fprint(w, `{"id":1}`)
	})

	err := client.UpdateWebhook(context.Background(), "1", vcs.UpdateWebhookOptions{
		Repo: "acme/terraform",
	})
	require.NoError(t, err)
}

func TestClient_GetWebhook(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("GET /api/v4/projects/{path_with_namespace}/hooks/1", func(w http.ResponseWriter, r *http.Request) {
		pathWithNamespace := r.PathValue("path_with_namespace")
		require.Equal(t, pathWithNamespace, "acme/terraform")

		fmt.Fprint(w, `{"id":1}`)
	})

	_, err := client.GetWebhook(context.Background(), vcs.GetWebhookOptions{
		ID:   "1",
		Repo: "acme/terraform",
	})
	require.NoError(t, err)
}

func TestClient_DeleteWebhook(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("DELETE /api/v4/projects/{path_with_namespace}/hooks/1", func(w http.ResponseWriter, r *http.Request) {
		pathWithNamespace := r.PathValue("path_with_namespace")
		require.Equal(t, pathWithNamespace, "acme/terraform")

		fmt.Fprint(w, `{"id":1}`)
	})

	err := client.DeleteWebhook(context.Background(), vcs.DeleteWebhookOptions{
		ID:   "1",
		Repo: "acme/terraform",
	})
	require.NoError(t, err)
}

func TestClient_ListPullRequestFiles(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("GET /api/v4/projects/{path_with_namespace}/merge_requests/1/diffs", func(w http.ResponseWriter, r *http.Request) {
		pathWithNamespace := r.PathValue("path_with_namespace")
		require.Equal(t, pathWithNamespace, "acme/terraform")

		fmt.Fprint(w, `[{"old_path":"main.tf","new_path":"main.tf"},{"old_path":"dev.tf","new_path":"prod.tf"}]`)
	})

	got, err := client.ListPullRequestFiles(context.Background(), "acme/terraform", 1)
	require.NoError(t, err)
	assert.Equal(t, []string{"dev.tf", "main.tf", "prod.tf"}, got)
}

func TestClient_GetCommit(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("GET /api/v4/projects/{path_with_namespace}/repository/commits/abc123", func(w http.ResponseWriter, r *http.Request) {
		pathWithNamespace := r.PathValue("path_with_namespace")
		require.Equal(t, pathWithNamespace, "acme/terraform")

		fmt.Fprint(w, `{"id":"abc123","web_url":"https://gitlab.com/commits/abc123"}`)
	})

	got, err := client.GetCommit(context.Background(), "acme/terraform", "abc123")
	require.NoError(t, err)
	want := vcs.Commit{
		SHA: "abc123",
		URL: "https://gitlab.com/commits/abc123",
	}
	assert.Equal(t, want, got)
}
