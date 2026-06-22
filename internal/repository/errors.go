package repository

import (
	"errors"

	"github.com/erewhile/iam/internal/ent/db"
)

var (
	ErrNotFound      = errors.New("resource not found")
	ErrAlreadyExists = errors.New("resource already exists")
)

func wrapErr(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case db.IsNotFound(err):
		return ErrNotFound
	case db.IsConstraintError(err):
		return ErrAlreadyExists
	default:
		return err
	}
}
