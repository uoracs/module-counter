.PHONY: build install clean run container handler
all: build

build:
    @go build -o bin/module-logger cmd/module-logger/main.go

run: build
    @go run cmd/module-logger/main.go

install: build
    @cp bin/module-logger /usr/local/bin/module-logger

container:
    @docker build -t module-logger .

handler:
    @go build -o handler cmd/module-logger/main.go

clean:
    @rm -f bin/module-logger /usr/local/bin/module-logger
