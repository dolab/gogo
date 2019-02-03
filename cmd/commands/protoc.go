package commands

import (
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"

	"github.com/golib/cli"
)

// Proto
var (
	Proto *_Proto

	protoDirs = map[ProtoType][]string{
		ProtoTypeProto:    {"app", "protos"},
		ProtoTypeProtobuf: {"gogo", "pbs"},
		ProtoTypeService:  {"gogo", "services"},
		ProtoTypeClient:   {"gogo", "clients"},
	}
)

type _Proto struct{}

func (_ *_Proto) Command() cli.Command {
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

func (_ *_Proto) Flags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:   "all",
			EnvVar: "GOGO_RPC_ALL",
		},
		cli.BoolFlag{
			Name:   "skip-testing",
			EnvVar: "GOGO_SKIP_TESTING",
		},
	}
}

func (_ *_Proto) Action() cli.ActionFunc {
	return func(ctx *cli.Context) error {
		log.Infof(">>> %#v", ctx.Args())

		return nil
	}
}

func (_ *_Proto) NewProtobuf() cli.ActionFunc {
	return func(ctx *cli.Context) error {
		name := path.Clean(ctx.Args().First())

		return Proto.newProto(ProtoTypeProtobuf, name)
	}
}

func (_ *_Proto) NewService() cli.ActionFunc {
	return func(ctx *cli.Context) error {
		name := path.Clean(ctx.Args().First())

		return Proto.newProto(ProtoTypeService, name)
	}
}

func (_ *_Proto) NewClient() cli.ActionFunc {
	return func(ctx *cli.Context) error {
		name := path.Clean(ctx.Args().First())

		return Proto.newProto(ProtoTypeClient, name)
	}
}

func (_ *_Proto) newProto(proto ProtoType, name string, args ...string) error {
	if !proto.Valid() {
		return ErrProtoType
	}

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
			log.Info("==> You can specify your installation by passing GOGO_PROTOC or install it by running *brew install protoc*")
		default:
			log.Info("==> You can install it following https://github.com/protocolbuffers/protobuf/releases")
		}

		return err
	}

	// detect protoc-gen-gogo plugin
	if _, err := exec.LookPath("protoc-gen-gogo"); err != nil {
		log.Warnf("Cannot find protoc-gen-gogo plugin: %v", err)
		log.Info("==> You can install it by running *go get -u github.com/dolab/gogo/cmd/protoc-gen-gogo*")

		return err
	}

	root, err := os.Getwd()
	if err != nil {
		log.Error(err.Error())

		return err
	}

	protoRoot := proto.Root(root)
	pbsRoot := protoRoot
	switch proto {
	case ProtoTypeProtobuf:
		// ignore
	default:
		pbsRoot = ProtoTypeProtobuf.Root(root)
	}

	protoArgs := []string{
		// specify root path of protos for protoc compiler
		"--proto_path", ProtoTypeProto.Root(root),
		// specify go_out options for protoc compiler
		"--go_out", pbsRoot,
	}

	switch proto {
	case ProtoTypeProtobuf:
		// ignore
	case ProtoTypeService:
		// specify gogo_out options of service for protoc-gen-gogo plugin
		protoArgs = append(protoArgs, "--gogo_out")
		protoArgs = append(protoArgs, "package_name=services,service=source_only:"+protoRoot)

	case ProtoTypeClient:
		// specify gogo_out options of client for protoc-gen-gogo plugin
		protoArgs = append(protoArgs, "--gogo_out")
		protoArgs = append(protoArgs, "package_name=clients,client=source_only:"+protoRoot)
	}

	protoArgs = append(protoArgs, name)

	output, err := exec.Command(protoc, protoArgs...).CombinedOutput()
	if err != nil {
		log.Errorf("Run protoc %s: %v", strings.Join(protoArgs, " "), err)
		log.Info(string(output))
		return err
	}

	log.Infof("Generate from %s: OK!", name)
	return nil
}
