package api

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	types "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/jsonapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tofutf/tofutf/internal"
)

func TestClient_UnmarshalResponse(t *testing.T) {
	want := types.WorkspaceList{
		Items: []*types.Workspace{
			{ID: "ws-1", Outputs: []*types.WorkspaceOutputs{}},
			{ID: "ws-2", Outputs: []*types.WorkspaceOutputs{}},
		},
		Pagination: &types.Pagination{},
	}
	var b bytes.Buffer
	bw := io.Writer(&b)
	err := jsonapi.MarshalPayload(bw, want)
	require.NoError(t, err)

	var got types.WorkspaceList
	err = unmarshalResponse(bytes.NewReader(b.Bytes()), &got)
	require.NoError(t, err)

	assert.Equal(t, want, got)
}

func TestClient_checkResponseCode(t *testing.T) {
	tests := []struct {
		name     string
		response *http.Response
		want     error
	}{
		{"200 OK", &http.Response{StatusCode: 200}, nil},
		{"204 No Content", &http.Response{StatusCode: 204}, nil},
		{"401 Not Authorized", &http.Response{StatusCode: 401}, internal.ErrUnauthorized},
		{"404 Not Found", &http.Response{StatusCode: 404}, internal.ErrResourceNotFound},
		{
			"500 Error",
			&http.Response{
				Status: "500 Internal Server Error",
				Body:   newBody(`{"errors":[{"status":"500","title":"Internal Server Error","detail":"cannot marshal unknown type: *types.AgentToken"}]}`),
			},
			errors.New("Internal Server Error: cannot marshal unknown type: *types.AgentToken"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, checkResponseCode(tt.response))
		})
	}
}

type bodyReader struct {
	*strings.Reader
}

func newBody(body string) *bodyReader {
	return &bodyReader{Reader: strings.NewReader(body)}
}

func (r *bodyReader) Close() error { return nil }
