package templates

var (
	makefileTemplate = `all: gobuild gotest gopackage

dev: gobuild godev

test: gobuild gotest

godev:
	go run app/main.go

gobuild: goclean goinstall

gorebuild: goclean goreinstall

goclean:
	go clean ./...
	go mod tidy

goinstall:
	go get -v github.com/dolab/gogo@v3.1.0
	go get -v github.com/dolab/httpmitm@master
	go get -v github.com/dolab/httptesting@master
	go get -v github.com/golib/assert@master

goreinstall:
	go get -v -u github.com/dolab/gogo@master
	go get -v -u github.com/dolab/httpmitm@master
	go get -v -u github.com/dolab/httptesting@master
	go get -v -u github.com/golib/assert@master

gotest:
	go test {{.Namespace}}/{{.Application}}/app/controllers
	go test {{.Namespace}}/{{.Application}}/app/middlewares
	go test {{.Namespace}}/{{.Application}}/app/models

gopackage:
	mkdir -p bin && go build -a -o bin/{{.Application}} app/main.go

travis: gobuild gotest
`
)
