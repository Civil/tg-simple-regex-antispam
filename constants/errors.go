package constants

import (
	"errors"
)

var ErrRequiresStateDir = errors.New(
	"banDB requires `state_dir` configuration parameter",
)

var ErrStateDirNotString = errors.New(
	"state_dir is not a string",
)

var ErrKeyCollision = errors.New(
	"key collision found",
)

var (
	ErrStateDirEmpty = errors.New("state_dir cannot be empty")
	ErrNIsZero       = errors.New("n cannot be equal to 0")
)
