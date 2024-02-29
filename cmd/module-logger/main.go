package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
)

var logFilePath = "/var/log/modules.log"

func main() {
	var userFlag = flag.String("user", "", "username")
	var packageFlag = flag.String("package", "", "package name")
	var versionFlag = flag.String("version", "", "package version")
	flag.Parse()

	if *userFlag == "" || *packageFlag == "" || *versionFlag == "" {
		fmt.Println("Usage: module-logger --user <username> --package <package> --version <version>")
		os.Exit(1)
	}

	fileHandle, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("error opening log file for appending: %v\n", err)
		os.Exit(1)
	}
	logger := slog.New(slog.NewTextHandler(fileHandle, nil))

	logger.Info("loaded module", "user", *userFlag, "package", *packageFlag, "version", *versionFlag)
}
