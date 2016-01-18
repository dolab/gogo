all: gobuild gotest

godev:
	cd gogo && go run main.go

gobuild: goclean goinstall

gorebuild: goclean goreinstall

goclean:
	rm -rf bin
	rm -rf pkg

goinstall:
	cd gogo && go get -t -v ./...

goreinstall:
	cd gogo && go get -t -a -v ./...

gotest:
	go test github.com/skeleton/app/controllers
	go test github.com/skeleton/app/middlewares
	# go test github.com/skeleton/app/models

gopackage:
	mkdir -p bin && go build -a -o bin/skeleton src/github.com/skeleton/gogo/main.go

travis: gobuild gotest
