package otf

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStream(t *testing.T) {
	store := testChunkStore{
		store: map[string][]byte{
			"test-123": []byte("\x02cat sat on the map\x03"),
		},
	}

	buf := new(bytes.Buffer)

	err := Stream(context.Background(), "test-123", &store, buf, time.Millisecond, 1)
	require.NoError(t, err)
}

func TestGetChunk(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		opts    GetChunkOptions
		want    string
		wantErr bool
	}{
		{
			name: "get all data",
			data: "1234567890",
			want: "1234567890",
		},
		{
			name: "get chunk",
			data: "1234567890",
			opts: GetChunkOptions{Offset: 3, Limit: 4},
			want: "4567",
		},
		{
			name: "get chunk with limit beyond size of data",
			data: "1234567890",
			opts: GetChunkOptions{Offset: 3, Limit: 99},
			want: "4567890",
		},
		{
			name:    "get chunk with offset beyond size of data",
			data:    "1234567890",
			opts:    GetChunkOptions{Offset: 99, Limit: 4},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetChunk([]byte(tt.data), tt.opts)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, string(got))
			}
		})
	}
}