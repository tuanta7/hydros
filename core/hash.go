package core

import (
	"context"

	"golang.org/x/crypto/bcrypt"
)

type Hasher interface {
	Compare(ctx context.Context, hash, data []byte) error
	Hash(ctx context.Context, data []byte) ([]byte, error)
}

type BCrypt struct {
	cost int
}

func NewBCryptHasher(cost int) *BCrypt {
	return &BCrypt{
		cost: cost,
	}
}

const DefaultBCryptWorkFactor = 12

func (b *BCrypt) Hash(ctx context.Context, data []byte) ([]byte, error) {
	wf := b.cost
	if wf == 0 {
		wf = DefaultBCryptWorkFactor
	}

	s, err := bcrypt.GenerateFromPassword(data, wf)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (b *BCrypt) Compare(ctx context.Context, hash, data []byte) error {
	if err := bcrypt.CompareHashAndPassword(hash, data); err != nil {
		return err
	}
	return nil
}
