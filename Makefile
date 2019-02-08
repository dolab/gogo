all: goclean gotest gorace

goclean:
	go clean -r ./...
	go mod tidy

gotest:
	go test -v -timeout 10s -race github.com/dolab/gogo/pkgs/errors
	go test -v -timeout 10s -race github.com/dolab/gogo/pkgs/gid
	go test -v -timeout 10s -race github.com/dolab/gogo/pkgs/named
	go test -v -timeout 10s -race github.com/dolab/gogo/pkgs/protocol/json
	go test -v -timeout 10s -race github.com/dolab/gogo/pkgs/protocol/jsonpb
	go test -v -timeout 10s -race github.com/dolab/gogo/pkgs/protocol/protobuf
	go test -v -timeout 10s -race github.com/dolab/gogo/internal/params
	go test -v -timeout 10s -race github.com/dolab/gogo/internal/protoc-gen/message
	go test -v -timeout 10s -race github.com/dolab/gogo/internal/render
	go test -v -timeout 10s
	go test -v -timeout 10s -race

gorace: goclean
	go test -timeout 10s -race

gobench: goclean
	go test -run=none -bench=. ./internal/render
	go test -run=none -bench=.

travis: goclean gotest gorace
