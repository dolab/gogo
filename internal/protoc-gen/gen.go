package gen

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/golang/protobuf/proto"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

// Generator defines protoc plugin interface
type Generator interface {
	Generate(in *plugin.CodeGeneratorRequest) *plugin.CodeGeneratorResponse
}

// Run runs a Generator with protoc input and write data generated to os.Stdout for protoc
func Run(g Generator) {
	in := unmarshal(os.Stdin)
	out := g.Generate(in)
	marshal(os.Stdout, out)
}

func unmarshal(r io.Reader) *plugin.CodeGeneratorRequest {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		Errorf("read os.Stdin: %v", err)
	}

	req := new(plugin.CodeGeneratorRequest)
	if err = proto.Unmarshal(data, req); err != nil {
		Errorf("unmarshal proto message: %v", err)
	}

	if len(req.FileToGenerate) == 0 {
		Failf("no files to generate")
	}

	return req
}

func marshal(w io.Writer, resp *plugin.CodeGeneratorResponse) {
	data, err := proto.Marshal(resp)
	if err != nil {
		Errorf("marshal proto message: %v", err)
	}
	_, err = w.Write(data)
	if err != nil {
		Errorf("flush data: %v", err)
	}
}
