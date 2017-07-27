all: clean build
test: build go_test

build:
	go build

clean:
	test -f pipe && rm pipe || echo 'cleaned' 

go_test:
	go test -v `go list  ./... | grep -v vendor`
