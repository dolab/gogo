package message

import (
	"github.com/dolab/gogo/pkgs/errors"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

// A Registry is parsed protoc descriptors
type Registry struct {
	allFiles    []*descriptor.FileDescriptorProto
	nameToFiles map[string]*descriptor.FileDescriptorProto

	// Mapping of fully-qualified names to their definitions
	protoToMessages map[string]*MessageDefinition
}

func New(files []*descriptor.FileDescriptorProto) *Registry {
	r := &Registry{
		allFiles:        files,
		nameToFiles:     make(map[string]*descriptor.FileDescriptorProto),
		protoToMessages: make(map[string]*MessageDefinition),
	}

	// First, index the file descriptors by name. We need this so
	// NewMessageFromFile can correctly scan imports.
	for _, file := range files {
		r.nameToFiles[file.GetName()] = file
	}

	// Next, index all the message definitions by their fully-qualified proto
	// names.
	for _, file := range files {
		defs := NewMessageFromFile(file, r.nameToFiles)

		for name, def := range defs {
			r.protoToMessages[name] = def
		}
	}
	return r
}

func (r *Registry) FileComments(file *descriptor.FileDescriptorProto) (DefinitionComments, error) {
	return commentsAtPath([]int32{packagePath}, file), nil
}

func (r *Registry) ServiceComments(file *descriptor.FileDescriptorProto, svc *descriptor.ServiceDescriptorProto) (DefinitionComments, error) {
	for i, s := range file.Service {
		if s == svc {
			path := []int32{servicePath, int32(i)}
			return commentsAtPath(path, file), nil
		}
	}

	return DefinitionComments{}, errors.New(file.GetName(), "service not found in file", nil)
}

func (r *Registry) MethodComments(file *descriptor.FileDescriptorProto, svc *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) (DefinitionComments, error) {
	for i, s := range file.Service {
		if s == svc {
			path := []int32{servicePath, int32(i)}
			for j, m := range s.Method {
				if m == method {
					path = append(path, serviceMethodPath, int32(j))
					return commentsAtPath(path, file), nil
				}
			}
		}
	}

	return DefinitionComments{}, errors.New(file.GetName(), "service not found in file", nil)
}

func (r *Registry) MethodInputDefinition(method *descriptor.MethodDescriptorProto) *MessageDefinition {
	return r.protoToMessages[method.GetInputType()]
}

func (r *Registry) MethodOutputDefinition(method *descriptor.MethodDescriptorProto) *MessageDefinition {
	return r.protoToMessages[method.GetOutputType()]
}

func (r *Registry) MessageDefinition(name string) *MessageDefinition {
	return r.protoToMessages[name]
}
