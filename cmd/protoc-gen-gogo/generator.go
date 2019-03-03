package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"path"
	"strconv"
	"strings"

	gen "github.com/dolab/gogo/internal/protoc-gen"
	"github.com/dolab/gogo/internal/protoc-gen/message"
	"github.com/dolab/gogo/pkgs/named"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"golang.org/x/tools/go/ast/astutil"
)

const (
	serviceNamespaceSuffix = "ServiceNamespace"
)

type generator struct {
	filesHandled int

	// Package naming:
	packageName       string // Name of the package that we're generating
	packagePath       string // Path of the package that we're outputing
	fileToPackageName map[*descriptor.FileDescriptorProto]string

	// List of files that were inputs to the generator. We need to hold this in
	// the struct so we can write a header for the file that lists its inputs.
	genFiles []*descriptor.FileDescriptorProto

	registry *message.Registry

	// Map to record whether we've built each package
	aliases    map[string]string
	aliasNames map[string]bool

	importPrefix string            // String to prefix to imported package file names.
	importMap    map[string]string // Mapping from .proto file name to import path.

	// Package output:
	sourceRelativePaths  bool // instruction on where to write output files
	sourceOnlyService    bool // is generate service only?
	sourceOnlyClient     bool // is generate client only?
	sourceOnlyAPI        bool // is generate api only?
	sourceOnlyAPITesting bool // is generate api testing only?

	// Output buffer that holds the bytes we want to write out for a single file.
	// Gets reset after working on a file.
	buf *bytes.Buffer
}

func newGenerator() *generator {
	g := &generator{
		aliases:           make(map[string]string),
		aliasNames:        make(map[string]bool),
		importMap:         make(map[string]string),
		fileToPackageName: make(map[*descriptor.FileDescriptorProto]string),
		buf:               bytes.NewBuffer(nil),
	}

	return g
}

func (g *generator) Generate(in *plugin.CodeGeneratorRequest) *plugin.CodeGeneratorResponse {
	// Parse options from command line
	opts, err := parseCommandOpts(in.GetParameter())
	if err != nil {
		gen.Failf("could not parse options passed to --gogo_out=[k=v]: %v", err.Error())
	}

	g.importPrefix = opts.importPrefix
	g.importMap = opts.importMap
	g.sourceRelativePaths = opts.relativePath
	g.sourceOnlyService = opts.onlyService
	g.sourceOnlyClient = opts.onlyClient
	g.sourceOnlyAPI = opts.onlyAPI
	g.sourceOnlyAPITesting = opts.onlyAPITesting

	// Collect information on types.
	g.registry = message.New(in.ProtoFile)
	g.genFiles = gen.AllDescriptorFiles(in)

	// Register names of packages that we import.
	g.aliasPackageName("context")
	g.aliasPackageName("fmt")
	g.aliasPackageName("http")
	g.aliasPackageName("url")
	g.aliasPackageName("sync")
	g.aliasPackageName("gogo")
	g.aliasPackageName("protocol")
	g.aliasPackageName("clients")
	g.aliasPackageName("errors")
	g.aliasPackageName("pbs")
	g.aliasPackageName("testing")
	g.aliasPackageName("assert")

	// Time to figure out package names of objects defined in protobuf.
	// First, we'll figure out the name for the package we're generating.
	packageName, err := gen.DeducePackageName(g.genFiles)
	if err != nil {
		gen.Failf("cannot resolve package name for generation: %v", err.Error())
	}
	if name := named.ToIdentifier(opts.packageName); name != "" {
		g.packageName = name
	} else {
		g.packageName = packageName
	}

	// required by generateAPI
	g.packagePath = opts.packagePath

	// Next, we need to pick names for all the files that are dependencies.
	for _, file := range in.ProtoFile {
		if fileDescSliceContains(g.genFiles, file) {
			// This is a file we are generating. It gets the shared package name.
			g.fileToPackageName[file] = g.packageName
		} else {
			// This is a dependency. Use its package name.
			name := file.GetPackage()
			if name == "" {
				name = named.Basename(file.GetName())
			}

			name = named.ToIdentifier(name)

			g.fileToPackageName[file] = g.aliasPackageName(name)
		}
	}

	// Showtime! Generate the response.
	resp := new(plugin.CodeGeneratorResponse)
	for _, file := range g.genFiles {
		switch {
		case g.sourceOnlyAPI:
			genFiles := g.generateAPI(file)
			if len(genFiles) > 0 {
				resp.File = append(resp.File, genFiles...)
			}
		default:
			genFile := g.generate(file)
			if genFile != nil {
				resp.File = append(resp.File, genFile)
			}
		}
	}
	return resp
}

func (g *generator) generate(file *descriptor.FileDescriptorProto) *plugin.CodeGeneratorResponse_File {
	if len(file.Service) == 0 || g.sourceOnlyAPI {
		return nil
	}

	tmpfile := new(plugin.CodeGeneratorResponse_File)

	g.genFileHeader(file)

	switch {
	case g.sourceOnlyService:
		tmpfile.Name = proto.String(g.goFilename(file, ".service.go"))

		g.genServiceImport(file)

		// for each service, generate service
		for i, service := range file.Service {
			g.genService(file, service, i)
		}

		g.genFileDescriptor(file)
	case g.sourceOnlyClient:
		tmpfile.Name = proto.String(g.goFilename(file, ".client.go"))

		g.genClientImport(file)

		// for each service, generate client
		for i, service := range file.Service {
			g.genClient(file, service, i)
		}
	default:
		tmpfile.Name = proto.String(g.goFilename(file, ".rpc.go"))

		g.genRPCImport(file)

		// for each service, generate service and client stubs
		for i, service := range file.Service {
			g.genService(file, service, i)
			g.genClient(file, service, i)
		}

		g.genFileDescriptor(file)
	}

	tmpfile.Content = proto.String(g.goFormat())

	g.filesHandled++
	return tmpfile
}

func (g *generator) generateAPI(file *descriptor.FileDescriptorProto) []*plugin.CodeGeneratorResponse_File {
	if len(file.Service) == 0 || !g.sourceOnlyAPI {
		return nil
	}

	// gen xxx.api.go
	codes, err := ioutil.ReadFile(path.Join(g.packagePath, g.goFilename(file, ".api.go")))
	if err != nil {
		g.genAPIFileHeader(file)
		g.genAPIImport(file)

		// for each service, generate api stubs
		for i, service := range file.Service {
			g.genAPI(file, service, i)
		}
	} else {
		g.P(string(codes))

		// for each service, generate api stubs
		for _, service := range file.Service {
			g.genAPIImplement(file, service)
		}
	}

	apiFile := new(plugin.CodeGeneratorResponse_File)
	apiFile.Name = proto.String(g.goFilename(file, ".api.go"))
	apiFile.Content = proto.String(g.goFormat(true))

	// gen xxx.api_test.go
	codes, err = ioutil.ReadFile(path.Join(g.packagePath, g.goFilename(file, ".api_test.go")))
	if err != nil {
		g.genAPIFileHeader(file)
		g.genAPITestingImport(file)

		// for each service, gnerate api testing stubs
		for i, service := range file.Service {
			g.genAPITesting(file, service, i)
		}
	} else {
		g.P(string(codes))

		// for each service, gnerate api testing stubs
		for _, service := range file.Service {
			g.genAPITestingImplement(file, service)
		}
	}

	testingFile := new(plugin.CodeGeneratorResponse_File)
	testingFile.Name = proto.String(g.goFilename(file, ".api_test.go"))
	testingFile.Content = proto.String(g.goFormat(true))

	g.filesHandled++
	return []*plugin.CodeGeneratorResponse_File{apiFile, testingFile}
}

// Big header comments to makes it easier to visually parse a generated file.
func (g *generator) genComment(sectionTitle string) {
	g.P()
	g.P(`// `, strings.Repeat("=", len(sectionTitle)))
	g.P(`// `, sectionTitle)
	g.P(`// `, strings.Repeat("=", len(sectionTitle)))
	g.P()
}

func (g *generator) genFileHeader(file *descriptor.FileDescriptorProto) {
	g.P("// DO NOT EDIT! Code generated by protoc-gen-gogo(", gen.Version, ").")
	g.P("// \tsource: ", file.GetName())
	g.P()

	// global comments
	if g.filesHandled == 0 {
		g.P("/*")
		g.P("Package ", g.packageName, " is a generated gogo rpc stub package.")
		g.P("This code was generated with github.com/dolab/gogo/cmd/protoc-gen-gogo(", gen.Version, ").")
		g.P()

		comment, err := g.registry.FileComments(file)
		if err == nil && comment.Leading != "" {
			for _, line := range strings.Split(comment.Leading, "\n") {
				line = strings.TrimPrefix(line, " ")
				// ensure we don't escape from the block comment
				line = strings.Replace(line, "*/", "*-/", -1)
				g.P(line)
			}
			g.P()
		}

		g.P("It is generated from these files:")

		for _, f := range g.genFiles {
			g.P("\t", f.GetName())
		}
		g.P("*/")
	}

	g.P(`package `, g.packageName)
	g.P()
}

func (g *generator) genAPIFileHeader(file *descriptor.FileDescriptorProto) {
	g.P("// Code generated by protoc-gen-gogo(", gen.Version, ").")
	g.P("// \tsource: ", file.GetName())
	g.P()
	g.P(`package `, g.packageName)
	g.P()
}

func (g *generator) genRPCImport(file *descriptor.FileDescriptorProto) {
	if len(file.Service) == 0 {
		return
	}

	g.P(`import (`)
	g.P(`	"context"`)
	g.P(`	"fmt"`)
	g.P(`	"net/http"`)
	g.P(`	"sync"`)
	g.P()
	g.P(`	"github.com/dolab/gogo"`)
	g.P(`	"github.com/dolab/gogo/pkgs/protocol"`)
	g.P(`	"github.com/dolab/gogo/pkgs/named"`)
	g.P(`	"`, g.importPrefix, `/gogo/errors"`)
	g.P(`	"`, g.importPrefix, `/gogo/pbs"`)

	// for dependences
	g.genDependenceImport(file)

	g.P(`)`)
	g.P()
}

func (g *generator) genServiceImport(file *descriptor.FileDescriptorProto) {
	if len(file.Service) == 0 {
		return
	}

	g.P(`import (`)
	g.P(`	"context"`)
	g.P(`	"fmt"`)
	g.P(`	"net/http"`)
	g.P(`	"sync"`)
	g.P()
	g.P(`	"github.com/dolab/gogo"`)
	g.P(`	"github.com/dolab/gogo/pkgs/protocol"`)
	g.P(`	"`, g.importPrefix, `/gogo/pbs"`)

	// for dependences
	g.genDependenceImport(file)

	g.P(`)`)
	g.P()
}

func (g *generator) genClientImport(file *descriptor.FileDescriptorProto) {
	if len(file.Service) == 0 {
		return
	}

	g.P(`import (`)
	g.P(`	"context"`)
	g.P(`	"net/http"`)
	g.P(`	"sync"`)
	g.P()
	g.P(`	"github.com/dolab/gogo/pkgs/protocol"`)
	g.P(`	"github.com/dolab/gogo/pkgs/named"`)
	g.P(`	"`, g.importPrefix, `/gogo/pbs"`)

	// for dependences
	g.genDependenceImport(file)

	g.P(`)`)
	g.P()
}

func (g *generator) genAPIImport(file *descriptor.FileDescriptorProto) {
	if len(file.Service) == 0 {
		return
	}

	g.P(`import (`)
	g.P(`	"context"`)
	g.P()
	g.P(`	"github.com/dolab/gogo/pkgs/protocol"`)
	g.P(`	"`, g.importPrefix, `/gogo/pbs"`)
	g.P(`)`)
	g.P()
}

func (g *generator) genAPITestingImport(file *descriptor.FileDescriptorProto) {
	if len(file.Service) == 0 {
		return
	}

	g.P(`import (`)
	g.P(`	"context"`)
	g.P(`	"net/http"`)
	g.P(`	"testing"`)
	g.P()
	g.P(`	"github.com/dolab/gogo/pkgs/protocol"`)
	g.P(`	"github.com/golib/assert"`)
	g.P(`	"`, g.importPrefix, `/gogo/clients"`)
	g.P(`	"`, g.importPrefix, `/gogo/pbs"`)
	g.P(`)`)
	g.P()
}

func (g *generator) genDependenceImport(file *descriptor.FileDescriptorProto) {
	// It's legal to import a message and use it as an input or output for a
	// method. Make sure to import the package of any such message. First, dedupe
	// them.
	deps := make(map[string]string) // Map of package name to quoted import path.
	for _, s := range file.Service {
		for _, m := range s.Method {
			messages := []*message.MessageDefinition{
				g.registry.MethodInputDefinition(m),
				g.registry.MethodOutputDefinition(m),
			}
			for _, message := range messages {
				substitution, ok := g.importMap[message.File.GetName()]
				if !ok {
					continue
				}

				pkg := g.goPackageName(message.File)
				importPath := g.importPrefix + substitution
				deps[pkg] = strconv.Quote(importPath)
			}
		}
	}

	if len(deps) > 0 {
		g.P()
	}
	for pkg, importPath := range deps {
		g.P(`	`, pkg, ` `, importPath)
	}
}

func (g *generator) genMethodSignature(method *descriptor.MethodDescriptorProto) string {
	name := toMethodName(method)
	input := g.goTypeName(method.GetInputType())
	output := g.goTypeName(method.GetOutputType())

	return fmt.Sprintf(`%s(ctx %s.Context, in *%s.%s) (out *%s.%s, err error)`,
		name,
		g.aliases["context"],
		g.aliases["pbs"], input,
		g.aliases["pbs"], output,
	)
}

func (g *generator) genService(
	file *descriptor.FileDescriptorProto,
	service *descriptor.ServiceDescriptorProto,
	index int) {
	svcName := toServiceName(service)

	// Const.
	g.genServiceNamespace(file, service)

	// Interface
	g.genComment(svcName + ` Interface`)
	g.genServiceInterface(file, service)

	// Service
	g.genComment(svcName + ` Service Implements`)
	g.genServiceImplement(file, service)
}

func (g *generator) genServiceNamespace(
	file *descriptor.FileDescriptorProto,
	service *descriptor.ServiceDescriptorProto) {
	svcName := toServiceName(service)
	nsName := svcName + serviceNamespaceSuffix

	g.P(`// `, nsName, ` is used for all URL paths on a gogo `, svcName, ` service.`)
	g.P(`// It can be used in an HTTP mux to route gogo requests along with non-gogo requests on other routes.`)
	g.P(`//`)
	g.P(`// Requests are always: POST `, nsName, `/<Method>`)
	g.P(`const (`)
	g.P(`	`, nsName, ` = `, strconv.Quote(pathPrefix(file, service)))
	g.P(`)`)
	g.P()
}

func (g *generator) genServiceInterface(
	file *descriptor.FileDescriptorProto,
	service *descriptor.ServiceDescriptorProto) {
	svcName := toServiceName(service)

	comments, err := g.registry.ServiceComments(file, service)
	if err == nil {
		g.writeComments(comments)
	} else {
		g.P(`// A `, svcName, ` service interface`)
	}

	g.P(`type `, svcName, ` interface {`)
	for _, method := range service.Method {
		comments, err := g.registry.MethodComments(file, service, method)
		if err == nil {
			g.writeComments(comments)
		}

		g.P(`	`, g.genMethodSignature(method))
		g.P()
	}
	g.P(`}`)
}

func (g *generator) genServiceImplement(
	file *descriptor.FileDescriptorProto,
	service *descriptor.ServiceDescriptorProto) {
	svcName := toServiceName(service)
	svcStruct := toServiceStruct(service)
	nsName := svcName + serviceNamespaceSuffix

	// Server implementation.
	g.P(`type `, svcStruct, ` struct {`)
	g.P(`	`, svcName)
	g.P()
	g.P(`	mux       `, g.aliases["sync"], `.RWMutex`)
	g.P(`	namespace string`)
	g.P(`	names     []string`)
	// g.P(`	hooks     *`, g.aliases["gogo"], `.ServerHooks`)
	g.P(`}`)
	g.P()

	// Constructor for service implementation
	g.P(`// New`, svcName, `Service returns `, g.aliases["gogo"], `.RPCServicer of `, svcName)
	// g.P(`func New`, svcName, `Service(svc `, svcName, `, hooks *`, g.aliases["gogo"], `.ServerHooks) `, g.aliases["gogo"], `RPCServicer {`)
	g.P(`func New`, svcName, `Service(svc `, svcName, `) `, g.aliases["gogo"], `.RPCServicer {`)
	g.P(`	return &`, svcStruct, `{`)
	g.P(`		`, svcName, `: svc,`)
	g.P(`		namespace: `, nsName, `,`)
	g.P(`	}`)
	g.P(`}`)
	g.P()

	// Methods
	for _, method := range service.Method {
		g.genServiceImplementMethod(file, service, method)
	}

	// Accessors
	g.genServiceAccessors(file, service)

	// Serve Errors
	g.genServiceServeError(file, service)
}

func (g *generator) genServiceImplementMethod(
	file *descriptor.FileDescriptorProto,
	service *descriptor.ServiceDescriptorProto,
	method *descriptor.MethodDescriptorProto) {
	pkgName := toPackageName(file)
	svcName := toServiceName(service)
	svcStruct := toServiceStruct(service)
	name := named.ToCamelCase(method.GetName())

	g.P(`func (s *`, svcStruct, `) Serve`, name, `(ctx *`, g.aliases["gogo"], `.Context) {`)
	g.P(`	rtx := ctx.Request.Context()`)
	g.P(`	rtx = `, g.aliases["protocol"], `.WithPackage(rtx, "`, pkgName, `")`)
	g.P(`	rtx = `, g.aliases["protocol"], `.WithService(rtx, "`, svcName, `")`)
	g.P(`	rtx = `, g.aliases["protocol"], `.WithMethod(rtx, "`, name, `")`)
	g.P(`	rtx = `, g.aliases["protocol"], `.WithResponseWriter(rtx, ctx.Response)`)
	g.P()
	g.P(`	// resolve codec of protocol`)
	g.P(`	codec, err := `, g.aliases["protocol"], `.NewFromHTTPRequest(ctx.Request)`)
	g.P(`	if err != nil {`)
	g.P(`		err = `, g.aliases["protocol"], `.ErrInvalidProtocol.WithError(err)`)
	g.P()
	g.P(`		s.serveError(rtx, ctx.Response, err)`)
	g.P(`		return`)
	g.P(`	}`)
	g.P()
	g.P(`	// parse input`)
	g.P(`	input := new(`, g.aliases["pbs"], `.`, g.goTypeName(method.GetInputType()), `)`)
	g.P()
	g.P(`	if err := codec.Unmarshal(ctx.Request.Body, input); err != nil {`)
	g.P(`		err = `, g.aliases["protocol"], `.ErrMalformedRequestMessage.WithError(err)`)
	g.P()
	g.P(`		s.serveError(rtx, ctx.Response, err)`)
	g.P(`		return`)
	g.P(`	}`)
	g.P()
	g.P(`	// call service method`)
	g.P(`	var output *`, g.aliases["pbs"], `.`, g.goTypeName(method.GetOutputType()))
	g.P(`	func() {`)
	g.P(`		defer func() {`)
	g.P(`			// ensure request body closed`)
	g.P(`			ctx.Request.Body.Close()`)
	g.P()
	g.P(`			// In case of a panic, serve a 500 error of panic.`)
	g.P(`			if r := recover(); r != nil {`)
	g.P(`				s.serveError(rtx, ctx.Response, `, g.aliases["protocol"], `.ErrInternalError.WithError(`, g.aliases["fmt"], `.Errorf("PANIC: %v", r)))`)
	g.P(`			}`)
	g.P(`		}()`)
	g.P()
	g.P(`		output, err = s.`, svcName, `.`, name, `(rtx, input)`)
	g.P(`	}()`)
	g.P(`	if err != nil {`)
	g.P(`		s.serveError(rtx, ctx.Response, err)`)
	g.P(`		return`)
	g.P(`	}`)
	g.P(`	if output == nil {`)
	g.P(`		s.serveError(rtx, ctx.Response, `, g.aliases["protocol"], `.ErrInvalidResponseMessage)`)
	g.P(`		return`)
	g.P(`	}`)
	g.P()
	g.P(`	// marshal response data`)
	g.P(`	buf, err := codec.Marshal(output)`)
	g.P(`	if err != nil {`)
	g.P(`		err = `, g.aliases["protocol"], `.ErrInvalidMarshaler.WithError(err)`)
	g.P()
	g.P(`		s.serveError(rtx, ctx.Response, err)`)
	g.P(`		return`)
	g.P(`	}`)
	g.P()
	g.P(`	// response to client`)
	g.P(`	rtx = `, g.aliases["protocol"], `.WithStatusCode(rtx, `, g.aliases["http"], `.StatusOK)`)
	g.P()
	g.P(`	ctx.SetStatus(`, g.aliases["http"], `.StatusOK)`)
	g.P(`	ctx.SetHeader("Content-Type", codec.ContentType())`)
	g.P(`	ctx.Response.FlushHeader()`)
	g.P()
	g.P(`	if _, err := ctx.Response.Write(buf); err != nil {`)
	g.P(`		err = `, g.aliases["protocol"], `.ErrWritePOSTResponse.WithError(err)`)
	g.P()
	g.P(`		ctx.Logger.Error(err.Error())`)
	g.P(`	}`)
	g.P(`}`)
	g.P()
}

func (g *generator) genServiceAccessors(
	file *descriptor.FileDescriptorProto,
	service *descriptor.ServiceDescriptorProto) {
	svcStruct := toServiceStruct(service)

	index := 0
	for i, s := range file.Service {
		if s.GetName() == service.GetName() {
			index = i
		}
	}

	g.P(`// ProtocGenGogoVersion returns version of protoc-gen-gogo generated this stub package`)
	g.P(`func (s *`, svcStruct, `) ProtocGenGogoVersion() string {`)
	g.P(`	return `, strconv.Quote(gen.Version))
	g.P(`}`)
	g.P()
	g.P(`// ServiceNamespace returns namespace of service`)
	g.P(`func (s *`, svcStruct, `) ServiceNamespace() string {`)
	g.P(`	return s.namespace`)
	g.P(`}`)
	g.P()
	g.P(`// ServiceNames returns all names registered of service`)
	g.P(`func (s *`, svcStruct, `) ServiceNames() []string {`)
	g.P(`	return s.names`)
	g.P(`}`)
	g.P()
	g.P(`// ServiceRegistry returns all names to handlers of service required by `, g.aliases["gogo"], `.MountRPC interface`)
	g.P(`func (s *`, svcStruct, `) ServiceRegistry(prefix string) map[string]`, g.aliases["gogo"], `.Middleware {`)
	g.P(`	s.mux.Lock()`)
	g.P(`	defer s.mux.Unlock()`)
	g.P()
	g.P(`	prefix += s.ServiceNamespace()`)
	g.P()
	g.P(`	s.names = []string{`)

	// Names
	for _, method := range service.Method {
		name := named.ToCamelCase(method.GetName())
		g.P(`		prefix + "`, name, `",`)
	}

	g.P(`	}`)
	g.P()
	g.P(`	return map[string]`, g.aliases["gogo"], `.Middleware{`)

	// Services
	for _, method := range service.Method {
		name := named.ToCamelCase(method.GetName())
		g.P(`		prefix + "`, name, `": s.Serve`, name, `,`)
	}

	g.P(`	}`)
	g.P(`}`)
	g.P()
	g.P(`func (s *`, svcStruct, `) ServiceDescriptor() ([]byte, int) {`)
	g.P(`	return `, g.getFileDescriptorName(file), `, `, strconv.Itoa(index))
	g.P(`}`)
	g.P()
}

func (g *generator) genServiceServeError(
	file *descriptor.FileDescriptorProto,
	service *descriptor.ServiceDescriptorProto) {
	svcStruct := toServiceStruct(service)

	g.P(`// serveError writes an HTTP response with a valid `, g.aliases["protocol"], `.RequestFailure error.`)
	g.P(`// If err is not a `, g.aliases["protocol"], `.Error, it will get wrapped.`)
	g.P(`func (s *`, svcStruct, `) serveError(ctx `, g.aliases["context"], `.Context, resp `, g.aliases["http"], `.ResponseWriter, err error) {`)
	g.P(`	// non request failure error are wrapped as Internal (default)`)
	g.P(`	perr, ok := err.(`, g.aliases["protocol"], `.RequestFailure)`)
	g.P(`	if !ok {`)
	g.P(`		perr = `, g.aliases["protocol"], `.ErrInternalError`)
	g.P(`	}`)
	g.P()
	g.P(`	statusCode := perr.StatusCode()`)
	g.P(`	ctx = `, g.aliases["protocol"], `.WithStatusCode(ctx, statusCode)`)
	g.P()
	g.P(`	resp.Header().Set("Content-Type", "application/json") // Error responses are always JSON (instead of protobuf)`)
	g.P(`	resp.WriteHeader(statusCode)                          // HTTP response status code`)
	g.P()
	g.P(`	respBody := perr.Error()`)
	g.P(`	if _, werr := resp.Write([]byte(respBody)); werr != nil {`)
	g.P(`		// We have two options here. We could log the error, or just silently ignore the error.`)
	g.P(`		//`)
	g.P(`		// Logging is unacceptable because we don't have a user-controlled`)
	g.P(`		// logger; writing out to stderr without permission is too rude.`)
	g.P(`		//`)
	g.P(`		// Silently ignoring the error is our least-bad option. It's highly`)
	g.P(`		// likely that the connection is broken and the original 'err' says`)
	g.P(`		// so anyway.`)
	g.P(`		_ = werr`)
	g.P(`	}`)
	g.P(`}`)
}

func (g *generator) genClient(
	file *descriptor.FileDescriptorProto,
	service *descriptor.ServiceDescriptorProto,
	index int) {
	svcName := toServiceName(service)

	// Const.
	g.genServiceNamespace(file, service)

	// Service Interface
	g.genComment(svcName + ` Interface`)
	g.genServiceInterface(file, service)

	// Client Implements
	g.genComment(svcName + ` Client`)
	g.genClientImplement(file, service)
}

func (g *generator) genClientImplement(
	file *descriptor.FileDescriptorProto,
	service *descriptor.ServiceDescriptorProto) {
	svcName := toServiceName(service)
	nsName := svcName + serviceNamespaceSuffix
	clientStruct := toClientStruct(service)
	newClientFunc := "New" + svcName + "Client"
	count := strconv.Itoa(len(service.Method))

	g.P(`type `, clientStruct, ` struct {`)
	g.P(`	mux    sync.RWMutex`)
	g.P(`	client `, g.aliases["protocol"], `.HTTPClient`)
	g.P(`	proto  *`, g.aliases["protocol"], `.Protocol`)
	g.P(`	routes [`, count, `]string`)
	g.P(`}`)
	g.P()
	g.P(`// `, newClientFunc, ` creates a client that stubs the `, svcName, ` interface.`)
	g.P(`// It communicates using protocol supplied and can be configured with a custom `, g.aliases["protocol"], `.HTTPClient.`)
	g.P(`func `, newClientFunc, `(addr string, ptype `, g.aliases["protocol"], `.ProtocolType, clients ...`, g.aliases["protocol"], `.HTTPClient) `, svcName, ` {`)
	g.P(`	// resolve protocol`)
	g.P(`	proto, err := `, g.aliases["protocol"], `.New(ptype)`)
	g.P(`	if err != nil {`)
	g.P(`		panic(err)`)
	g.P(`	}`)
	g.P()
	g.P(`	// define routes`)
	g.P(`	prefix := named.ToURL(addr) + `, nsName)
	g.P(`	routes := [`, count, `]string{`)

	// Routes
	for _, method := range service.Method {
		name := named.ToCamelCase(method.GetName())
		g.P(`		prefix + "`, name, `",`)
	}

	g.P(`	}`)
	g.P(``)
	g.P(`	if len(clients) == 0 {`)
	g.P(`		clients = []`, g.aliases["protocol"], `.HTTPClient{`)
	g.P(`			&`, g.aliases["http"], `.Client{},`)
	g.P(`		}`)
	g.P(`	}`)
	g.P()
	g.P(`	return &`, clientStruct, `{`)
	g.P(`		client: clients[0],`)
	g.P(`		proto:  proto,`)
	g.P(`		routes: routes,`)
	g.P(`	}`)
	g.P(`}`)
	g.P()

	// Methods
	for i, method := range service.Method {
		g.genClientImplementMethod(file, service, method, i)
	}
}

func (g *generator) genClientImplementMethod(
	file *descriptor.FileDescriptorProto,
	service *descriptor.ServiceDescriptorProto,
	method *descriptor.MethodDescriptorProto,
	index int) {
	pkgName := toPackageName(file)
	svcName := toServiceName(service)
	clientStruct := toClientStruct(service)

	// Metadata
	name := toMethodName(method)
	output := g.goTypeName(method.GetOutputType())

	// Comments
	comments, err := g.registry.MethodComments(file, service, method)
	if err == nil {
		g.writeComments(comments)
	}

	// Method
	g.P(`func (c *`, clientStruct, `) `, g.genMethodSignature(method), ` {`)
	g.P(`	ctx = `, g.aliases["protocol"], `.WithPackage(ctx, "`, pkgName, `")`)
	g.P(`	ctx = `, g.aliases["protocol"], `.WithService(ctx, "`, svcName, `")`)
	g.P(`	ctx = `, g.aliases["protocol"], `.WithMethod(ctx, "`, name, `")`)
	g.P()
	g.P(`	out = new(`, g.aliases["pbs"], `.`, output, `)`)
	g.P()
	g.P(`	err = c.proto.NewRequest(c.client).Do(ctx, c.routes[`, strconv.Itoa(index), `], in, out)`)
	g.P(`	if err != nil {`)
	g.P(`		out = nil`)
	g.P(`	}`)
	g.P(`	return`)
	g.P(`}`)
	g.P()
}

func (g *generator) genAPI(
	file *descriptor.FileDescriptorProto,
	service *descriptor.ServiceDescriptorProto,
	index int) {
	svcName := toServiceName(service)
	apiStruct := toAPIStruct(service)

	g.P(`// A `, svcName, ` API`)
	g.P(`var (`)
	g.P(`	`, svcName, ` *`, apiStruct, ``)
	g.P(`)`)
	g.P()
	g.P(`type `, apiStruct, ` struct{}`)
	g.P()

	// API Implements
	g.genAPIImplement(file, service)
}

func (g *generator) genAPIImplement(
	file *descriptor.FileDescriptorProto,
	service *descriptor.ServiceDescriptorProto) {
	// APIs
	for _, method := range service.Method {
		g.genAPIImplementMethod(file, service, method)
	}
}

func (g *generator) genAPIImplementMethod(
	file *descriptor.FileDescriptorProto,
	service *descriptor.ServiceDescriptorProto,
	method *descriptor.MethodDescriptorProto) {
	apiStruct := toAPIStruct(service)

	// Comments
	comments, err := g.registry.MethodComments(file, service, method)
	if err == nil {
		g.writeComments(comments)
	}

	// Method
	g.P(`func (*`, apiStruct, `) `, g.genMethodSignature(method), ` {`)
	g.P(`	err = `, g.aliases["protocol"], `.ErrNotImplemented`)
	g.P(`	return`)
	g.P(`}`)
	g.P()
}

func (g *generator) genAPITesting(
	file *descriptor.FileDescriptorProto,
	service *descriptor.ServiceDescriptorProto,
	index int) {
	// API Testings
	g.genAPITestingImplement(file, service)
}

func (g *generator) genAPITestingImplement(
	file *descriptor.FileDescriptorProto,
	service *descriptor.ServiceDescriptorProto) {
	// API Testings
	for _, method := range service.Method {
		g.genAPITestingImplementMethod(file, service, method)
	}
}

func (g *generator) genAPITestingImplementMethod(
	file *descriptor.FileDescriptorProto,
	service *descriptor.ServiceDescriptorProto,
	method *descriptor.MethodDescriptorProto) {
	svcName := toServiceName(service)

	name := toMethodName(method)
	input := g.goTypeName(method.GetInputType())

	g.P(`func Test_`, svcName, `_`, name, `(t *`, g.aliases["testing"], `.T) {`)
	g.P(`	it := assert.New(t)`)
	g.P(`	client := clients.New`, svcName, `Client(gogotesting.Url("/v1"), `, g.aliases["protocol"], `.ProtocolTypeJSON)`)
	g.P()
	g.P(`	// it should work`)
	g.P(`	in := new(`, g.aliases["pbs"], `.`, input, `)`)
	g.P()
	g.P(`	out, err := client.`, name, `(context.Background(), in)`)
	g.P(`	if it.NotNil(err) {`)
	g.P(`		it.Nil(out)`)
	g.P()
	g.P(`		perr, ok := err.(`, g.aliases["protocol"], `.RequestFailure)`)
	g.P(`		if it.True(ok) {`)
	g.P(`			it.Equal(`, g.aliases["http"], `.StatusNotImplemented, perr.StatusCode())`)
	g.P(`		}`)
	g.P(`	}`)
	g.P(`}`)
	g.P()
}

// getFileDescriptorName is the variable name used in generated code to refer
// to the compressed bytes of this descriptor. It is not exported, so it is only
// valid inside the generated package.
//
// protoc-gen-go writes its own version of this file, but so does
// protoc-gen-gogo - with a different name! Twirp aims to be compatible with
// both; the simplest way forward is to write the file descriptor again as
// another variable that we control.
func (g *generator) getFileDescriptorName(file *descriptor.FileDescriptorProto) string {
	// Copied straight out of protoc-gen-go, which trims out comments.
	pb := proto.Clone(file).(*descriptor.FileDescriptorProto)
	pb.SourceCodeInfo = nil

	b, err := proto.Marshal(pb)
	if err != nil {
		gen.Failf(err.Error())
	}

	// gen hash of descriptor content
	hasher := md5.New()
	hasher.Write(b)

	return fmt.Sprintf("gogoFileDescriptor_%s_%x", named.Basename(file.GetName()), hasher.Sum(nil))
}

func (g *generator) genFileDescriptor(file *descriptor.FileDescriptorProto) {
	// Copied straight out of protoc-gen-go, which trims out comments.
	pb := proto.Clone(file).(*descriptor.FileDescriptorProto)
	pb.SourceCodeInfo = nil

	b, err := proto.Marshal(pb)
	if err != nil {
		gen.Failf(err.Error())
	}

	// gen gzip of descriptor
	var buf bytes.Buffer
	w, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	w.Write(b)
	w.Close()
	b = buf.Bytes()

	g.P()
	g.P("var ", g.getFileDescriptorName(file), " = []byte{")
	g.P("	// ", fmt.Sprintf("%d", len(b)), " bytes of a gzipped FileDescriptorProto")
	for len(b) > 0 {
		n := 16
		if n > len(b) {
			n = len(b)
		}

		s := ""
		for _, c := range b[:n] {
			s += fmt.Sprintf("0x%02x,", c)
		}
		g.P(`	`, s)

		b = b[n:]
	}
	g.P("}")
	g.P()
}

// goFilename returns the output name for the generated Go file.
func (g *generator) goFilename(file *descriptor.FileDescriptorProto, suffix string) string {
	name := *file.Name
	if ext := path.Ext(name); ext == ".proto" || ext == ".protodevel" {
		name = name[:len(name)-len(ext)]
	}

	name += suffix

	if g.sourceRelativePaths {
		return name
	}

	// Does the file have go_package option? If it does, it may override the
	// filename.
	if importPath, _, ok := gen.GoPackageOption(file); ok && importPath != "" {
		// Replace the existing dirname with the declared import path.
		_, name = path.Split(name)

		name = path.Join(importPath, name)

		return name
	}

	return name
}

func (g *generator) goPackageName(file *descriptor.FileDescriptorProto) string {
	return g.fileToPackageName[file]
}

// Given a protobuf name for a Message, return the Go name we will use for that
// type, including its package prefix.
func (g *generator) goTypeName(protoName string) string {
	def := g.registry.MessageDefinition(protoName)
	if def == nil {
		gen.Failf("could not find message for %q", protoName)
	}

	var prefix string
	if pkg := g.goPackageName(def.File); pkg != g.packageName {
		prefix = pkg + "."
	}

	var name string
	for _, parent := range def.Lineage() {
		name += parent.Descriptor.GetName() + "_"
	}
	name += def.Descriptor.GetName()

	return prefix + toTypeName(name)
}

func (g *generator) goFormat(dupMiddlewares ...bool) string {
	defer g.buf.Reset()

	// Reformat generated code.
	fset := token.NewFileSet()
	raw := g.buf.Bytes()

	fast, err := parser.ParseFile(fset, "", raw, parser.ParseComments)
	if err != nil {
		// Print out the bad code with line numbers.
		// This should never happen in practice, but it can while changing generated code,
		// so consider this a debugging aid.
		var src bytes.Buffer
		s := bufio.NewScanner(bytes.NewReader(raw))
		for line := 1; s.Scan(); line++ {
			fmt.Fprintf(&src, "%5d\t%s\n", line, s.Bytes())
		}

		gen.Failf("invalid source code was generated: %v\n%s\n", err, src.String())
	}

	// trim duplicated methods
	if len(dupMiddlewares) > 0 && dupMiddlewares[0] {
		commentMap := ast.NewCommentMap(fset, fast, fast.Comments)

		methods := map[string]bool{}
		astutil.Apply(fast, func(cur *astutil.Cursor) bool {
			if decl, ok := cur.Node().(*ast.FuncDecl); ok {
				if methods[decl.Name.Name] {
					cur.Delete()

					return false
				}

				methods[decl.Name.Name] = true
			}

			return true
		}, func(cur *astutil.Cursor) bool {
			return true
		})

		fast.Comments = commentMap.Filter(fast).Comments()
	}

	out := bytes.NewBuffer(nil)
	err = (&printer.Config{
		Mode:     printer.TabIndent | printer.UseSpaces,
		Tabwidth: 8,
	}).Fprint(out, fset, fast)
	if err != nil {
		gen.Failf("cannot reformat source code: %v", err)
	}

	return out.String()
}

// P forwards to g.gen.P by appending newline.
func (g *generator) P(args ...string) {
	for _, v := range args {
		g.buf.WriteString(v)
	}

	g.buf.WriteByte('\n')
}

func (g *generator) writeComments(comments message.DefinitionComments) bool {
	text := strings.TrimSuffix(comments.Leading, "\n")
	if len(strings.TrimSpace(text)) == 0 {
		return false
	}

	split := strings.Split(text, "\n")
	for _, line := range split {
		g.P("// ", strings.TrimPrefix(line, " "))
	}

	return len(split) > 0
}

func (g *generator) aliasPackageName(name string) (alias string) {
	alias = name

	i := 1
	for g.aliasNames[alias] {
		alias = name + strconv.Itoa(i)
		i++
	}

	g.aliasNames[alias] = true
	g.aliases[name] = alias

	return alias
}
