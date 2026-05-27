package repository

import "errors"

var (
	ErrNotFound        = errors.New("record not found")
	ErrConflict        = errors.New("resource conflict")
	ErrForbidden       = errors.New("forbidden")
	ErrDependencyCycle = errors.New("dependency cycle detected")
)
