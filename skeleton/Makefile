all: gobuild gotest gopackage

dev: gobuild godev

godev:
	go run app/main.go

gobuild: goclean goinstall

gorebuild: goclean goreinstall

goclean:
	go clean -r ./...

goinstall:
	go get github.com/dolab/httptesting
	go get github.com/dolab/httpmitm
	go get github.com/dolab/logger
	go get github.com/golib/assert

goreinstall:
	go get -u github.com/dolab/httptesting
	go get -u github.com/dolab/httpmitm
	go get -u github.com/dolab/logger
	go get -u github.com/golib/assert

gotest:
	go test github.com/skeleton/app/controllers
	go test github.com/skeleton/app/middlewares
	go test github.com/skeleton/app/models

gopackage:
	mkdir -p bin && go build -a -o bin/skeleton app/main.go

travis: gobuild gotest
