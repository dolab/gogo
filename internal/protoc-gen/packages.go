package gen

import (
	"fmt"
	"path"
	"strings"

	"github.com/dolab/gogo/pkgs/named"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

// PackageName returns the package name to use in the generated file.
// The result explicitly reports whether the name came from an option go_package
// statement. If explicit is false, the name was derived from the protocol
// buffer's package statement or the input file name.
func PackageName(file *descriptor.FileDescriptorProto) (name string, explicit bool) {
	// Does the file have a "go_package" option?
	if _, pkg, ok := GoPackageOption(file); ok {
		return pkg, true
	}

	// Does the file have a package clause?
	if pkg := file.GetPackage(); pkg != "" {
		return pkg, false
	}

	// Use the file base name.
	return named.Basename(file.GetName()), false
}

// DeducePackageName figures out the go package name to use for generated code.
// It will try to use the explicit go_package setting in a file (if set, must be
// consistent in all files). If no files have go_package set, then use the
// protobuf package name (must be consistent in all files)
func DeducePackageName(files []*descriptor.FileDescriptorProto) (string, error) {
	var packageName string
	for _, file := range files {
		name, explicit := PackageName(file)
		if explicit {
			name = named.ToIdentifier(name)
			if packageName != "" && packageName != name {
				// Make sure they're all set consistently.
				return "", fmt.Errorf("Protos have conflicting go_package options, must be the same: %q and %q", packageName, name)
			}
			packageName = name
		}
	}
	if packageName != "" {
		return packageName, nil
	}

	// If there is no explicit setting, then check the implicit package name
	// (derived from the protobuf package name) of the files and make sure it's
	// consistent.
	for _, file := range files {
		name, _ := PackageName(file)
		name = named.ToIdentifier(name)
		if packageName != "" && packageName != name {
			return "", fmt.Errorf("Protos have conflicting package names, must be the same or overridden with go_package: %q and %q", packageName, name)
		}
		packageName = name
	}

	// All the files have the same name, so we're good.
	return packageName, nil
}

// GoPackageOption interprets the file's go_package option.
// If there is no go_package, it returns ("", "", false).
// If there's a simple name, it returns ("", pkg, true).
// If the option implies an import path, it returns (importPath, pkg, true).
func GoPackageOption(file *descriptor.FileDescriptorProto) (importPath, pkg string, ok bool) {
	pkg = file.GetOptions().GetGoPackage()
	if pkg == "" {
		return
	}

	ok = true

	// The presence of a slash implies there's an import path.
	slash := strings.LastIndex(pkg, "/")
	if slash < 0 {
		return
	}

	importPath, pkg = pkg, pkg[slash+1:]

	// A semicolon-delimited suffix overrides the package name.
	sc := strings.IndexByte(importPath, ';')
	if sc < 0 {
		return
	}

	importPath, pkg = importPath[:sc], importPath[sc+1:]
	return
}

// ParseGoPackageOption interprets the file's go_package option.
// Allowed formats:
// 	option go_package = "foo";
// 	option go_package = "github.com/example/foo";
// 	option go_package = "github.com/example/foo;bar";
func ParseGoPackageOption(s string) (importPath, packageName string) {
	semicolonPos := strings.Index(s, ";")
	if semicolonPos > -1 {
		importPath = s[:semicolonPos]
		packageName = s[semicolonPos+1:]
		return
	}

	if strings.Contains(s, "/") {
		importPath = s
		_, packageName = path.Split(s)
		return
	}

	importPath = ""
	packageName = s
	return
}
