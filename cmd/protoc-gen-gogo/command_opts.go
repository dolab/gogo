package main

import (
	"fmt"
	"strings"
)

type commandOpts struct {
	packageName    string            // Generate package name
	importPrefix   string            // String to prefix to imported package file names.
	importMap      map[string]string // Mapping from .proto file name to import path.
	relativePaths  bool              // Paths value, used to control file output directory
	onlyService    bool              // Generate service only and output to xxx.service.go
	onlyClient     bool              // Generate client only and output to xxx.client.go
	onlyAPI        bool              // Generate api only and output to xxx.api.go
	onlyAPITesting bool              // Generate api testing only and output to xxx.api_test.go
}

// parseCommandOpts breaks the comma-separated list of key=value pairs
// in the parameter (a member of the request protobuf) into a key/value map.
// It then sets command line parameter mappings defined by those entries.
func parseCommandOpts(parameter string) (*commandOpts, error) {
	kvs := make(map[string]string)
	for _, p := range strings.Split(parameter, ",") {
		if p == "" {
			continue
		}

		i := strings.Index(p, "=")
		if i < 0 {
			return nil, fmt.Errorf("invalid parameter %q: expected format of parameter to be k=v", p)
		}

		k := p[0:i]
		v := p[i+1:]
		if v == "" {
			return nil, fmt.Errorf("invalid parameter %q: expected format of parameter to be k=v", k)
		}

		kvs[k] = v
	}

	args := &commandOpts{
		importMap: make(map[string]string),
	}
	for k, v := range kvs {
		switch {
		case k == "package_name":
			if v == "" {
				return nil, fmt.Errorf("package name does not support %q", v)
			}
			args.packageName = v
		case k == "import_prefix":
			args.importPrefix = v
		// Support import map 'M' prefix per https://github.com/golang/protobuf/blob/6fb5325/protoc-gen-go/generator/generator.go#L497.
		case len(k) > 0 && k[0] == 'M':
			args.importMap[k[1:]] = v // 1 is the length of 'M'.
		case len(k) > 0 && strings.HasPrefix(k, "go_import_mapping@"):
			args.importMap[k[18:]] = v // 18 is the length of 'go_import_mapping@'.
		case k == "paths":
			if v != "source_relative" {
				return nil, fmt.Errorf("paths does not support %q", v)
			}
			args.relativePaths = true
		case k == "service":
			if v != "source_only" {
				return nil, fmt.Errorf("service does not support %q", v)
			}
			args.onlyService = true
		case k == "client":
			if v != "source_only" {
				return nil, fmt.Errorf("client does not support %q", v)
			}
			args.onlyClient = true
		case k == "api":
			if v != "source_only" {
				return nil, fmt.Errorf("api does not support %q", v)
			}
			args.onlyAPI = true
		case k == "api_testing":
			if v != "source_only" {
				return nil, fmt.Errorf("api testing does not support %q", v)
			}
			args.onlyAPITesting = true
		default:
			return nil, fmt.Errorf("unknown parameter %q", k)
		}
	}

	if args.onlyService && args.onlyClient {
		return nil, fmt.Errorf("can not set service=source_only along with client=source_only simultaneous")
	}

	return args, nil
}
