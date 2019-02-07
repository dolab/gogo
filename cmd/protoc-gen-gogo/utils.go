package main

import (
	"fmt"
	"strings"

	"github.com/dolab/gogo/pkgs/named"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

func fullServiceName(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) string {
	name := named.ToCamelCase(service.GetName())
	if pkg := toPackageName(file); pkg != "" {
		name = pkg + "." + name
	}
	return name
}

func fileDescSliceContains(slice []*descriptor.FileDescriptorProto, f *descriptor.FileDescriptorProto) bool {
	for _, sf := range slice {
		if f == sf {
			return true
		}
	}
	return false
}

// pathPrefix returns the base path for all methods handled by a particular
// service. It includes a trailing slash. (for example "/gogo/example.Haberdasher/").
func pathPrefix(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) string {
	return fmt.Sprintf("/gogo/%s/", fullServiceName(file, service))
}

func unexportName(s string) string {
	return strings.ToLower(s[:1]) + s[1:]
}

func toPackageName(file *descriptor.FileDescriptorProto) string {
	return file.GetPackage()
}

func toServiceName(service *descriptor.ServiceDescriptorProto) string {
	return named.ToCamelCase(service.GetName())
}

func toServiceStruct(service *descriptor.ServiceDescriptorProto) string {
	return unexportName(toServiceName(service)) + "Service"
}

func toClientStruct(service *descriptor.ServiceDescriptorProto) string {
	return unexportName(toServiceName(service)) + "Client"
}

func toAPIStruct(service *descriptor.ServiceDescriptorProto) string {
	return unexportName(toServiceName(service)) + "API"
}

func toMethodName(method *descriptor.MethodDescriptorProto) string {
	return named.ToCamelCase(method.GetName())
}

func toTypeName(typo string) string {
	return named.ToCamelCase(typo)
}
