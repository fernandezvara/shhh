package ops

import (
	"errors"
)

// defaults
const (
	passwdOK = "password is ok" // this is the string that will be checked on executions
)

// errors
var (
	ErrAlreadyExist     = errors.New("file already exists")
	ErrNotExist         = errors.New("file do not exists")
	ErrPasswordNotMatch = errors.New("invalid password")
)

type Entry struct {
	ID    []byte `storm:"id"`
	Value []byte
}
