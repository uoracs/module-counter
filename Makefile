.PHONY: build install clean run
all: build

build:
	@go build -o bin/module-logger cmd/module-logger/main.go

install:
	@cp bin/module-logger /usr/local/sbin/module-logger
	@chown root:root /usr/local/sbin/module-logger
	@chmod 4755 /usr/local/sbin/module-logger

run: build
	@go run cmd/module-logger/main.go

clean:
	@rm -f bin/module-logger /usr/local/bin/module-logger
