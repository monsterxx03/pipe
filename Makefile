all: clean build

build:
	go build

clean:
	test -f pipe && rm pipe || echo 'cleaned' 
