package commands

import "path"

// Gogo component types of generated file
const (
	_comType ComponentType = iota
	ComTypeController
	ComTypeFilter
	ComTypeModel
	comType_
)

// A ComponentType defines component of gogo generate file type
type ComponentType int

func (ct ComponentType) Valid() bool {
	return ct > _comType && ct < comType_
}

func (ct ComponentType) Root(pwd string) string {
	pwd = ensureAppRoot(pwd)

	dirs, ok := comDirs[ct]
	if !ok {
		return pwd
	}

	return path.Clean(path.Join(pwd, path.Join(dirs...)))
}

func (ct ComponentType) String() string {
	switch ct {
	case ComTypeController:
		return "controller"

	case ComTypeFilter:
		return "filter"

	case ComTypeModel:
		return "model"

	}

	return "Unknown compoment"
}

// Gogo protoc types of generated file
const (
	_protoType ProtoType = iota
	ProtoTypeProto
	ProtoTypeProtobuf
	ProtoTypeService
	ProtoTypeClient
	protoType_
)

// A ProtoType defines protoc generate file type
type ProtoType int

func (pt ProtoType) Valid() bool {
	return pt > _protoType && pt < protoType_
}

func (pt ProtoType) Root(pwd string) string {
	pwd = ensureAppRoot(pwd)

	dirs, ok := protoDirs[pt]
	if !ok {
		return pwd
	}

	return path.Clean(path.Join(pwd, path.Join(dirs...)))
}

func (pt ProtoType) String() string {
	switch pt {
	case ProtoTypeProtobuf:
		return "protobuf"
	case ProtoTypeService:
		return "gogo rpc service"
	case ProtoTypeClient:
		return "gogo rpc client"
	}

	return "Unknown proto"
}
