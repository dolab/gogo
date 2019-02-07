package commands

import (
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"

	gen "github.com/dolab/gogo/internal/protoc-gen"

	"github.com/golib/cli"
)

// Proto
var (
	Proto *_Proto

	protoDirs = map[ProtoType][]string{
		ProtoTypeProto:    {"app", "protos"},
		ProtoTypeAPI:      {"app", "controllers"},
		ProtoTypeProtobuf: {"gogo", "pbs"},
		ProtoTypeService:  {"gogo", "services"},
		ProtoTypeClient:   {"gogo", "clients"},
	}
)

type _Proto struct{}

func (*_Proto) Command() cli.Command {
	return cli.Command{
		Name:    "protoc",
		Aliases: []string{"pg"},
		Usage:   "generate rpc components by invoking protoc compiler along with protoc-gen-gogo plugin.",
		Flags:   Proto.Flags(),
		Action:  Proto.Action(),
		Subcommands: cli.Commands{
			{
				Name:    "protobuf",
				Aliases: []string{"pb"},
				Usage:   "generate protobuf component.",
				Flags:   []cli.Flag{},
				Action:  Proto.NewProtobuf(),
			},
			{
				Name:    "service",
				Aliases: []string{"svc"},
				Usage:   "generate service component.",
				Flags:   []cli.Flag{},
				Action:  Proto.NewService(),
			},
			{
				Name:    "client",
				Aliases: []string{"c"},
				Usage:   "generate client stubs component of service.",
				Flags:   []cli.Flag{},
				Action:  Proto.NewClient(),
			},
		},
	}
}

func (*_Proto) Flags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   "protoc",
			EnvVar: "GOGO_PROTOC",
		},
		cli.BoolFlag{
			Name:   "all",
			EnvVar: "GOGO_PROTOC_ALL",
		},
		cli.BoolFlag{
			Name:   "skip-testing",
			EnvVar: "GOGO_SKIP_TESTING",
		},
	}
}

func (*_Proto) Action() cli.ActionFunc {
	return func(ctx *cli.Context) error {
		name := path.Clean(ctx.Args().First())

		// protobuf
		err := Proto.newProto(ProtoTypeProtobuf, name)
		if err != nil {
			return err
		}

		// service
		err = Proto.newProto(ProtoTypeService, name)
		if err != nil {
			return err
		}

		// client
		err = Proto.newProto(ProtoTypeClient, name)
		if err != nil {
			return err
		}

		// api
		err = Proto.newProto(ProtoTypeAPI, name)
		if err != nil {
			return err
		}

		return nil
	}
}

func (*_Proto) NewProtobuf() cli.ActionFunc {
	return func(ctx *cli.Context) error {
		name := path.Clean(ctx.Args().First())

		return Proto.newProto(ProtoTypeProtobuf, name)
	}
}

func (*_Proto) NewService() cli.ActionFunc {
	return func(ctx *cli.Context) error {
		name := path.Clean(ctx.Args().First())

		return Proto.newProto(ProtoTypeService, name)
	}
}

func (*_Proto) NewClient() cli.ActionFunc {
	return func(ctx *cli.Context) error {
		name := path.Clean(ctx.Args().First())

		return Proto.newProto(ProtoTypeClient, name)
	}
}

func (*_Proto) NewAPI() cli.ActionFunc {
	return func(ctx *cli.Context) error {
		name := path.Clean(ctx.Args().First())

		return Proto.newProto(ProtoTypeAPI, name)
	}
}

func (*_Proto) newProto(proto ProtoType, name string, args ...string) error {
	if !proto.Valid() {
		return ErrProtoType
	}

	// adjust name for proto file
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == ".." {
		log.Warnf("Please input file name of proto fo generating")
		return nil
	}
	if !strings.HasSuffix(name, ".proto") {
		name += ".proto"
	}

	// detect protoc compiler
	protoc, err := exec.LookPath("protoc")
	if err != nil {
		log.Warnf("Cannot find protoc compiler: %v", err)

		switch runtime.GOOS {
		case "darwin":
			log.Info("\t==> You can install it by running *brew install protoc*,")
		default:
			log.Info("\t==> You can install it by following instruction from https://github.com/protocolbuffers/protobuf/releases,")
		}
		log.Info("\t   or specify your installation by passing GOGO_PROTOC env.")

		return err
	}

	// detect protoc-gen-gogo plugin
	if _, err := exec.LookPath("protoc-gen-gogo"); err != nil {
		log.Warnf("Cannot find protoc-gen-gogo plugin: %v", err)
		log.Info("\t==> You can install it by running *go get -u github.com/dolab/gogo/cmd/protoc-gen-gogo*")

		return err
	}

	// detect root of command running
	root, err := os.Getwd()
	if err != nil {
		log.Error(err.Error())

		return err
	}

	// load app metadata
	protoApp, err := LoadAppData(root)
	if err != nil {
		log.Errorf("LoadAppData(%s): %v", root, err)

		return err
	}

	// NOTE: we use protobuf go out as handler input param and output return
	protoArgs := []string{
		// protoc --proto_path=?
		// specify root path of protos for protoc compiler
		"--proto_path", ProtoTypeProto.Root(root),
		// protoc --go_out=?
		// specify go_out options for protoc compiler
		"--go_out", ProtoTypeProtobuf.Root(root),
	}

	switch proto {
	case ProtoTypeProtobuf:
		// ignore
	case ProtoTypeService:
		// protoc --gogo_out=?
		// specify gogo_out options of service for protoc-gen-gogo plugin
		protoSvcRoot := ProtoTypeService.Root(root)

		protoArgs = append(protoArgs, "--gogo_out")
		protoArgs = append(protoArgs, "package_name=services,import_prefix="+protoApp.ImportPrefix()+",service=source_only:"+protoSvcRoot)

	case ProtoTypeClient:
		// protoc --gogo_out=?
		// specify gogo_out options of client for protoc-gen-gogo plugin
		protoClientRoot := ProtoTypeClient.Root(root)

		protoArgs = append(protoArgs, "--gogo_out")
		protoArgs = append(protoArgs, "package_name=clients,import_prefix="+protoApp.ImportPrefix()+",client=source_only:"+protoClientRoot)

	case ProtoTypeAPI:
		// protoc --gogo_out=?
		// specify gogo_out options of client for protoc-gen-gogo plugin
		protoAPIRoot := ProtoTypeAPI.Root(root)

		protoArgs = append(protoArgs, "--gogo_out")
		protoArgs = append(protoArgs, "package_name=controllers,import_prefix="+protoApp.ImportPrefix()+",api=source_only:"+protoAPIRoot)
	}

	// all protos
	names := gen.AllProtoImports(path.Join(ProtoTypeProto.Root(root), name))
	names = append(names, name)

	protoArgs = append(protoArgs, names...)

	output, err := exec.Command(protoc, protoArgs...).CombinedOutput()
	if err != nil {
		log.Warnf("Running\n\tprotoc %s\n\t ==> %s\n\t %v", strings.Join(protoArgs, " "), output, err)

		return err
	}

	log.Infof("Generate from %s: OK!", name)
	return nil
}
