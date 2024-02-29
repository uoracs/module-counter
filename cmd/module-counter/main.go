package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
)

var jsonPath = "test-packages.json"

// example data
// {
//   "users": [
//     {
//       "name": "lcrown",
//       "packages": [
//         { "name": "r", "version": "4.3.2", "count": 2 },
//         { "name": "httpie", "version": "10.1", "count": 10 }
//       ]
//     },
//     {
//       "name": "marka",
//       "packages": [
//         { "name": "r", "version": "4.3.2", "count": 198 },
//         { "name": "httpie", "version": "10.1", "count": 1 },
//         { "name": "matlab", "version": "12.4", "count": 2 }
//       ]
//     }
//   ]
// }

type PackageData struct {
	Users []UserPackages `json:"users"`
}

type UserPackages struct {
	Name     string `json:"name"`
	Packages []struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Count   int    `json:"count"`
	} `json:"packages"`
}

func errorAndExit(err error) {
	fmt.Printf("Error: %v\n", err)
	os.Exit(1)
}

func readUserPackagesFromFile(jsonPath string) ([]UserPackages, error) {
	// read json file
	var jsonData PackageData
	bytes, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("error reading json file: %v", err)
	}
	err = json.Unmarshal(bytes, &jsonData)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling json: %v", err)
	}
	return jsonData.Users, nil
}

func writeUserPackagesToFile(userPackages []UserPackages, jsonPath string) error {
	// write json file
	data := PackageData{Users: userPackages}
	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshalling json: %v", err)
	}
	err = os.WriteFile(jsonPath, bytes, 0600)
	if err != nil {
		return fmt.Errorf("error writing json file: %v", err)
	}
	return nil
}

func incrementPackage(userPackages []UserPackages, username string, packageName string, packageVersion string) []UserPackages {
	for i, user := range userPackages {
		if user.Name == username {
			// found the user
			for j, pkg := range user.Packages {
				if pkg.Name == packageName && pkg.Version == packageVersion {
					// found the package
					user.Packages[j].Count++
					slog.Debug("Incremented count", "user", username, "package", packageName, "version", packageVersion)
					return userPackages
				}
			}
			// package not found, add it
			userPackages[i].Packages = append(user.Packages, struct {
				Name    string `json:"name"`
				Version string `json:"version"`
				Count   int    `json:"count"`
			}{Name: packageName, Version: packageVersion, Count: 1})
			slog.Debug("Added package", "user", username, "package", packageName, "version", packageVersion)
			return userPackages
		}
	}
	// user not found, add them
	userPackages = append(userPackages, UserPackages{
		Name: username,
		Packages: []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
			Count   int    `json:"count"`
		}{
			{Name: packageName, Version: packageVersion, Count: 1},
		},
	})
	slog.Debug("Added user and package", "user", username, "package", packageName, "version", packageVersion)
	return userPackages
}

func main() {
	var userFlag = flag.String("user", "", "username")
	var packageFlag = flag.String("package", "", "package name")
	var versionFlag = flag.String("version", "", "package version")
	var debugFlag = flag.Bool("debug", false, "debug mode")
	flag.Parse()

	if *userFlag == "" || *packageFlag == "" || *versionFlag == "" {
		fmt.Println("Usage: module-counter --user <username> --package <package> --version <version>")
		os.Exit(1)
	}

	if *debugFlag {
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
		slog.SetDefault(logger)
	}

	// get json data
	userPackages, err := readUserPackagesFromFile(jsonPath)
	if err != nil {
		errorAndExit(err)
	}

	userPackages = incrementPackage(userPackages, *userFlag, *packageFlag, *versionFlag)

	// write json data
	err = writeUserPackagesToFile(userPackages, jsonPath)
	if err != nil {
		errorAndExit(err)
	}
}
