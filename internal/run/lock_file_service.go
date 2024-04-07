package run

import (
	"context"
	"fmt"

	"github.com/tofutf/tofutf/internal/rbac"
)

func lockFileCacheKey(runID string) string {
	return fmt.Sprintf("%s.terraform.lock.hcl", runID)
}

// GetLockFile returns the lock file for the run.
func (s *Service) GetLockFile(ctx context.Context, runID string) ([]byte, error) {
	subject, err := s.CanAccess(ctx, rbac.GetLockFileAction, runID)
	if err != nil {
		return nil, err
	}

	if plan, err := s.cache.Get(lockFileCacheKey(runID)); err == nil {
		return plan, nil
	}
	// cache miss; retrieve from db
	file, err := s.db.GetLockFile(ctx, runID)
	if err != nil {
		s.logger.Error("retrieving lock file", "id", runID, "subject", subject, "err", err)
		return nil, err
	}

	// cache lock file before returning
	if err := s.cache.Set(lockFileCacheKey(runID), file); err != nil {
		s.logger.Error("caching lock file", "err", err)
	}
	return file, nil
}

// UploadLockFile persists the lock file for a run.
func (s *Service) UploadLockFile(ctx context.Context, runID string, file []byte) error {
	subject, err := s.CanAccess(ctx, rbac.UploadLockFileAction, runID)
	if err != nil {
		return err
	}

	if err := s.db.SetLockFile(ctx, runID, file); err != nil {
		s.logger.Error("uploading lock file", "id", runID, "subject", subject, "err", err)
		return err
	}
	s.logger.Info("uploaded lock file", "id", runID)

	// cache lock file before returning
	if err := s.cache.Set(lockFileCacheKey(runID), file); err != nil {
		s.logger.Error("caching lock file", "err", err)
	}
	return nil
}
