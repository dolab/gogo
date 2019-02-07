all: goclean gotest gorace

goclean:
	go clean -r ./...
	go mod tidy

gotest:
	go test -v -race github.com/dolab/gogo/pkgs/errors
	go test -v -race github.com/dolab/gogo/pkgs/gid
	go test -v -race github.com/dolab/gogo/pkgs/named
	go test -v -race github.com/dolab/gogo/pkgs/protocol/json
	go test -v -race github.com/dolab/gogo/pkgs/protocol/jsonpb
	go test -v -race github.com/dolab/gogo/pkgs/protocol/protobuf
	go test -v -race github.com/dolab/gogo/internal/params
	go test -v -race github.com/dolab/gogo/internal/protoc-gen/message
	go test -v -race github.com/dolab/gogo/internal/render
	go test -v
	go test -v -race

gorace: goclean
	go test -race

gobench: goclean
	go test -run=^$ -bench=.

travis: goclean gotest gorace
