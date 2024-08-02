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

type ModuleActivation struct {
	Username       string    `json:"username"`
	PackageName    string    `json:"package_name"`
	PackageVersion string    `json:"package_version"`
	ModuleFilePath string    `json:"module_file_path"`
	Timestamp      time.Time `json:"timestamp"`
	Expiration     time.Time `json:"expiration"`
}

func NewModuleActivation(username string, packageName string, packageVersion string, moduleFilePath string, expireSeconds int) *ModuleActivation {
	return &ModuleActivation{
		Username:       username,
		PackageName:    packageName,
		PackageVersion: packageVersion,
		ModuleFilePath: moduleFilePath,
		Timestamp:      time.Now(),
		Expiration:     time.Now().Add(time.Duration(expireSeconds) * time.Second),
	}
}

type ModuleCache struct {
	Path        string
	Activations []ModuleActivation
}

func NewModuleCache(path string) *ModuleCache {
	return &ModuleCache{
		Path:        path,
		Activations: []ModuleActivation{},
	}
}

func (mc *ModuleCache) Load() (*ModuleCache, error) {
	cacheFile, err := os.Open(mc.Path)
	if err != nil {
		// not found means cache should just load empty
		// other errors? ehhhh
		return mc, nil
	}
	defer cacheFile.Close()

	fs, err := cacheFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file statistics: %v", err)
	}

	if fs.Size() == 0 {
		return mc, nil
	}

	bytes, err := io.ReadAll(cacheFile)
	if err != nil {
		return nil, fmt.Errorf("cache file exists but is unreadable %s: %v", mc.Path, err)
	}

	var activations []ModuleActivation
	err = json.Unmarshal(bytes, &activations)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache: %v", err)
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

func GetLogFileHandle(path string) (io.Writer, error) {
	fileHandle, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("error opening log file for appending: %v", err)
	}
	return fileHandle, nil
}

func Log(w io.Writer, ma *ModuleActivation) {
	logger := slog.New(slog.NewJSONHandler(w, nil))
	logger.Info("loaded module", "user", ma.Username, "package", ma.PackageName, "version", ma.PackageVersion, "path", ma.ModuleFilePath)
}

func PrintErrorAndExit(err error) {
	fmt.Printf("module-logger error: %v\n", err)
	os.Exit(1)
}

func PrintUsageAndExit() {
	fmt.Println("Usage: module-logger --user <username> --package <package> --version <version> --modulefilepath <path>")
	os.Exit(1)
}

type runArgs struct {
	user           string
	packageName    string
	packageVersion string
	moduleFilePath string
	expireSeconds  int
	cacheFilePath  string
	logFilePath    string
}

func (a runArgs) IsValid() bool {
	if a.user != "" && a.packageName != "" && a.packageVersion != "" {
		return true
	}
	return false
}

func Run(args runArgs) {
	var err error

	if !args.IsValid() {
		PrintUsageAndExit()
	}

	mc, err := NewModuleCache(args.cacheFilePath).Load()
	if err != nil {
		PrintErrorAndExit(err)
	}

	ma := NewModuleActivation(args.user, args.packageName, args.packageVersion, args.moduleFilePath, args.expireSeconds)

	if mc.ReadyToWrite(ma) {
		h, err := GetLogFileHandle(args.logFilePath)
		if err != nil {
			PrintErrorAndExit(err)
		}
		Log(h, ma)
		mc.Add(ma)
	}

	mc.Clean()

	err = mc.Save()
	if err != nil {
		PrintErrorAndExit(err)
	}
}

func main() {
	var userFlag = flag.String("user", "", "username")
	var packageFlag = flag.String("package", "", "package name")
	var versionFlag = flag.String("version", "", "package version")
	var moduleFilePathFlag = flag.String("modulefilepath", "", "path to the module file")
	var expireSecondsFlag = flag.Int("expireSeconds", 300, "timeout in seconds to not register duplicate activations")
	var cacheFilePathFlag = flag.String("cacheFilePath", "/gpfs/t2/module-logger/cache.json", "path for the module logger cache")
	var logFilePathFlag = flag.String("logFilePath", "/gpfs/t2/module-logger/modules.log", "path for the module logger log file")
	flag.Parse()

	args := runArgs{
		user:           *userFlag,
		packageName:    *packageFlag,
		packageVersion: *versionFlag,
		moduleFilePath: *moduleFilePathFlag,
		expireSeconds:  *expireSecondsFlag,
		cacheFilePath:  *cacheFilePathFlag,
		logFilePath:    *logFilePathFlag,
	}

	Run(args)
}
