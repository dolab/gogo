package commands

import "errors"

var (
	ErrNoneEmptyDirectory = errors.New("Can't initialize a new gogo application within an none empty directory, please choose an empty directory.")
)
