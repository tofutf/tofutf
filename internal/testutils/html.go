package testutils

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	otfhtml "github.com/tofutf/tofutf/internal/http/html"
	"golang.org/x/net/html"
)

func AttrMap(node *html.Node) map[string]string {
	m := make(map[string]string, len(node.Attr))
	for _, attr := range node.Attr {
		m[attr.Key] = attr.Val
	}
	return m
}

func AssertRedirect(t *testing.T, w *httptest.ResponseRecorder, path string) {
	if assert.Equal(t, 302, w.Code, w.Body.String()) {
		redirect, _ := w.Result().Location()
		assert.Equal(t, path, redirect.Path)
	}
}

func NewRenderer(t *testing.T) otfhtml.Renderer {
	renderer, err := otfhtml.NewRenderer(false)
	require.NoError(t, err)
	return renderer
}
