package html

import (
	"io"
	"log"
	"log/slog"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/jaschaephraim/lrserver"
)

var (
	// livereload relies on fsnotify, and fsnotify does not support watching
	// directories recursively:
	//
	// https://github.com/fsnotify/fsnotify/issues/18
	//
	// Therefore we need to reference each directory individually...
	staticPath    = filepath.Join(localPath, "static")
	templatesPath = filepath.Join(staticPath, "templates")
	watchPaths    = []string{
		templatesPath,
		filepath.Join(templatesPath, "partials"),
		filepath.Join(templatesPath, "content"),
		filepath.Join(staticPath, "css"),
		filepath.Join(staticPath, "js"),
		filepath.Join(staticPath, "images"),
	}
)

func startLiveReloadServer(logger *slog.Logger) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	for _, path := range watchPaths {
		if err := watcher.Add(path); err != nil {
			return err
		}
	}
	srv := lrserver.New(lrserver.DefaultName, lrserver.DefaultPort)

	// suppress noisy printing to stdout/stderr
	srv.SetStatusLog(log.New(io.Discard, "", 0))
	srv.SetErrorLog(log.New(io.Discard, "", 0))

	go srv.ListenAndServe() //nolint:errcheck
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				srv.Reload(event.Name)
			case err := <-watcher.Errors:
				logger.Error("livereload watcher error", "err", err)
			}
		}
	}()
	return nil
}
