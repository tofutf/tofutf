package pubsub

import "github.com/tofutf/tofutf/internal/sql"

type fakeListener struct{}

func (f *fakeListener) RegisterFunc(table string, ff sql.ForwardFunc) {
}
