package storage

import "context"

type Transactional interface {
	BeginTX(ctx context.Context) (context.Context, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

func TryBeginTX(ctx context.Context, storage any) (context.Context, error) {
	txStorage, ok := storage.(Transactional)
	if !ok {
		return ctx, nil
	}
	return txStorage.BeginTX(ctx)
}

func TryCommit(ctx context.Context, storage any) error {
	txStorage, ok := storage.(Transactional)
	if !ok {
		return nil
	}
	return txStorage.Commit(ctx)
}

func TryRollback(ctx context.Context, storage any) error {
	txStorage, ok := storage.(Transactional)
	if !ok {
		return nil
	}
	return txStorage.Rollback(ctx)
}
