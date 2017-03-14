all: clean build
test: build test_redis

build:
	go build

clean:
	test -f pipe && rm pipe || echo 'cleaned' 

test_redis:
	go test -d redis
