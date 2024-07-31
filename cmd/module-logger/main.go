package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"
)

var (
	DebounceTimoutSeconds int
	LogFilePath           string
)

type ModuleActivation struct {
	Username       string    `json:"username"`
	PackageName    string    `json:"package_name"`
	PackageVersion string    `json:"package_version"`
	Timestamp      time.Time `json:"timestamp"`
	Expiration     time.Time `json:"expiration"`
}

func NewModuleActivation(username string, packageName string, packageVersion string) *ModuleActivation {
	return &ModuleActivation{Username: username, PackageName: packageName, PackageVersion: packageVersion, Timestamp: time.Now(), Expiration: time.Now().Add(time.Duration(DebounceTimoutSeconds) * time.Second)}
}

func (ma *ModuleActivation) WithExpirationTimeout(seconds int) *ModuleActivation {
	ma.Expiration = ma.Timestamp.Add(time.Duration(seconds) * time.Second)
	return ma
}

type ModuleCache struct {
	Path                   string
	DebounceTimeoutSeconds int
	Activations            []ModuleActivation
}

func NewModuleCache(path string) *ModuleCache {
	return &ModuleCache{
		Path:        path,
		Activations: []ModuleActivation{},
	}
}

func (mc *ModuleCache) WithDebounceTimeout(seconds int) *ModuleCache {
	mc.DebounceTimeoutSeconds = seconds
	return mc
}

func (mc *ModuleCache) WithCacheFilePath(path string) *ModuleCache {
	mc.Path = path
	return mc
}

func (mc *ModuleCache) Load() (*ModuleCache, error) {
	cacheFile, err := os.Open(mc.Path)
	if err != nil {
		// not found means cache should just load empty
		// other errors? ehhhh
		return mc, nil
	}
	defer cacheFile.Close()

	bytes, err := io.ReadAll(cacheFile)
	if err != nil {
		return nil, fmt.Errorf("cache file exists but is unreadable %s: %v", mc.Path, err)
	}

	var activations []ModuleActivation
	err = json.Unmarshal(bytes, &activations)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache")
	}

	mc.Activations = append(mc.Activations, activations...)
	return mc, nil
}

func (mc *ModuleCache) Save() error {
	jsonData, err := json.Marshal(mc.Activations)
	if err != nil {
		return fmt.Errorf("failed to marshal json data from cache: %v", err)
	}

	err = os.WriteFile(mc.Path, jsonData, 0600)
	if err != nil {
		return fmt.Errorf("unable to open cache for writing: %s: %v", mc.Path, err)
	}

	return nil
}

func (mc *ModuleCache) ReadyToWrite(ma *ModuleActivation) bool {
	for _, mca := range mc.Activations {
		if mca.Username == ma.Username && mca.PackageName == ma.PackageName && mca.PackageVersion == ma.PackageVersion {
			if ma.Timestamp.Before(mca.Expiration) {
				return false
			}
		}
	}
	return true
}

func (mc *ModuleCache) Add(ma *ModuleActivation) {
	mc.Activations = append(mc.Activations, *ma)
}

func (mc *ModuleCache) Clean() {
	var unexpiredActivations []ModuleActivation
	for _, ma := range mc.Activations {
		if time.Now().Before(ma.Expiration) {
			unexpiredActivations = append(unexpiredActivations, ma)
		}
	}
	mc.Activations = unexpiredActivations
}

func Log(path string, ma *ModuleActivation) error {
	fileHandle, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("error opening log file for appending: %v\n", err)
	}
	logger := slog.New(slog.NewJSONHandler(fileHandle, nil))

	logger.Info("loaded module", "user", ma.Username, "package", ma.PackageName, "version", ma.PackageVersion)

	return nil
}

func PrintErrorAndQuit(err error) {
	fmt.Printf("module-logger error: %v\n", err)
	os.Exit(1)
}

func main() {
	var err error
	var userFlag = flag.String("user", "", "username")
	var packageFlag = flag.String("package", "", "package name")
	var versionFlag = flag.String("version", "", "package version")
	var debounceTimeoutFlag = flag.Int("debounceSeconds", 300, "timeout in seconds to not register duplicate activations")
	var cacheFilePathFlag = flag.String("cacheFilePath", "/gpfs/t2/module-logger/cache.json", "path for the module logger cache")
	var logFilePathFlag = flag.String("logFilePath", "/gpfs/t2/module-logger/modules.log", "path for the module logger log file")
	flag.Parse()

	if *userFlag == "" || *packageFlag == "" || *versionFlag == "" {
		fmt.Println("Usage: module-logger --user <username> --package <package> --version <version>")
		os.Exit(1)
	}

	DebounceTimoutSeconds = *debounceTimeoutFlag

	mc, err := NewModuleCache(*cacheFilePathFlag).Load()
	if err != nil {
		PrintErrorAndQuit(err)
	}

	ma := NewModuleActivation(*userFlag, *packageFlag, *versionFlag)

	if mc.ReadyToWrite(ma) {
		err = Log(*logFilePathFlag, ma)
		if err != nil {
			PrintErrorAndQuit(err)
		}
		mc.Add(ma)
	}

	mc.Clean()

	err = mc.Save()
	if err != nil {
		PrintErrorAndQuit(err)
	}
}
