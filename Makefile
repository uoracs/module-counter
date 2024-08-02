.PHONY: build install clean run test
all: build

build:
	@go build -o bin/module-logger main.go

test: 
	@go test ./...

install:
	@cp bin/module-logger /usr/local/sbin/module-logger
	@chown root:root /usr/local/sbin/module-logger
	@chmod 4755 /usr/local/sbin/module-logger

run: build
	@go run main.go

clean:
	@rm -f bin/module-logger /usr/local/bin/module-logger
