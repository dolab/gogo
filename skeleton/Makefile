all: gobuild gotest

godev:
	cd gogo && go run main.go

gobuild: goclean goinstall

gorebuild: goclean goreinstall

goclean:
	go clean -r ./...

goinstall:
	go get github.com/dolab/httptesting
	go get github.com/dolab/logger
	go get github.com/golib/assert

goreinstall:
	go get -u github.com/dolab/httptesting
	go get -u github.com/dolab/logger
	go get -u github.com/golib/assert

gotest:
	go test github.com/skeleton/app/controllers
	go test github.com/skeleton/app/middlewares
	go test github.com/skeleton/app/models

gopackage:
	mkdir -p bin && go build -a -o bin/skeleton src/github.com/skeleton/gogo/main.go

travis: gobuild gotest
