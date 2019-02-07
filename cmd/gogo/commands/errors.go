package commands

import "errors"

var (
	ErrNoneEmptyDirectory = errors.New("Can't initialize a new gogo application within an none empty directory, please choose an empty directory.")
	ErrComponentType      = errors.New("Invalid component type, available types are ComTypeController, ComTypeModel and ComTypeFilter")
	ErrProtoType          = errors.New("Invalid proto type, available types are ComTypeController, ComTypeModel and ComTypeFilter")
	ErrInvalidRoot        = errors.New("Invalid component code generation root path, it must run within /root/to/myapp or /root/to/myapp/gogo")
)
