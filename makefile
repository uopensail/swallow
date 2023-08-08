.PHONY: build clean run

GITHASH := $(shell git rev-parse --short HEAD)
GOLDFLAGS += -X app.__GITHASH__=$(GITHASH)
GOFLAGS = -ldflags "$(GOLDFLAGS)"

all: build-dev
third: 
	rm -rf ./lib 
	mkdir -p ./lib
	cd cpp && g++ -fpic -shared -o libswallow.so swallow.cpp -l rocksdb -std=c++17 && cp libswallow.so ../lib
build: third
	go build -o build/swallow $(GOFLAGS)
build-dev: build
	cp -rf conf/dev build/conf
build-prod: build
	cp -rf conf/prod build/conf
run: build-dev
	./build/swallow -config="./build/conf"
clean:
	rm -rf ./build