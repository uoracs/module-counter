.PHONY: build install clean run container handler
all: build

build:
    @go build -o bin/module-counter cmd/module-counter/main.go

run: build
    @go run cmd/module-counter/main.go

install: build
    @cp bin/module-counter /usr/local/bin/module-counter

container:
    @docker build -t module-counter .

handler:
    @go build -o handler cmd/module-counter/main.go

clean:
    @rm -f bin/module-counter /usr/local/bin/module-counter
