package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestRunArgsIsValid(t *testing.T) {
	tdt := []struct {
		in   runArgs
		want bool
	}{
		{
			runArgs{
				user:           "testuser1",
				packageName:    "testpackage1",
				packageVersion: "testversion1",
				moduleFilePath: "testfilepath1",
				expireSeconds:  9999,
				cacheFilePath:  "/test/cache/1",
				logFilePath:    "/test/module/path/1",
			},
			true,
		},
		{
			runArgs{
				user:           "testuser2",
				packageName:    "testpackage2",
				packageVersion: "testversion2",
				moduleFilePath: "testfilepath2",
			},
			true,
		},
		{
			runArgs{
				user:           "testinvaliduser2",
				packageName:    "testpackage2",
				packageVersion: "testversion2",
			},
			false,
		},
	}
	for _, tt := range tdt {
		t.Run(fmt.Sprintf("%v", tt.in), func(t *testing.T) {
			got := tt.in.IsValid()
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModuleCacheLoad(t *testing.T) {
	var c1 *ModuleCache
	var c2 *ModuleCache

	tempCachePath, err := os.CreateTemp("", "testcache")
	if err != nil {
		t.Errorf("failed to create temp cache: %v", err)
	}

	// shouldnt get any filesystem errors
	t.Run("create cache object", func(t *testing.T) {
		c1 = NewModuleCache(tempCachePath.Name())
		t.Logf("tempCachePath: %v", tempCachePath.Name())
		c1, err = c1.Load()
		if err != nil {
			t.Errorf("failed to load temp cache: %v", err)
		}
	})

	// empty cache should have 0 activations
	t.Run("no activations on empty cache", func(t *testing.T) {
		numActivations := len(c1.Activations)
		if numActivations != 0 {
			t.Errorf("temp cache activations should be 0, got %v", numActivations)
		}
	})

	// create a dummy activation and save it
	ma := NewModuleActivation("testuser1", "testpackage1", "testversion1", "testmodulefilepath1", 100)
	t.Run("add module activation and check length", func(t *testing.T) {
		c1.Add(ma)
		numActivations := len(c1.Activations)
		if numActivations != 1 {
			t.Errorf("temp cache activations should be 1, got %v", numActivations)
		}
	})

	// save the activation data to disk
	t.Run("save activations to disk", func(t *testing.T) {
		err := c1.Save()
		if err != nil {
			t.Errorf("failed saving temp cache: %v", err)
		}
	})

	// instantiate new cache object at same path
	t.Run("create new cache from same cache file", func(t *testing.T) {
		c2 = NewModuleCache(tempCachePath.Name())
		c2, err = c2.Load()
		if err != nil {
			t.Errorf("new temp cache object should read existing data: %v", err)
		}
	})

	// new cache object should have 1 module activation
	t.Run("one activation on existing cache", func(t *testing.T) {
		numActivations := len(c2.Activations)
		if numActivations != 1 {
			t.Errorf("new temp cache activations should be 1, got %v", numActivations)
		}
	})
}

func TestModuleCacheReadyToWrite(t *testing.T) {
	var c *ModuleCache

	tempCachePath, err := os.CreateTemp("", "testcache")
	if err != nil {
		t.Errorf("failed to create temp cache: %v", err)
	}

	c = NewModuleCache(tempCachePath.Name())
	ma1 := NewModuleActivation("testuser1", "testpackage1", "testversion1", "testmodulefilepath1", 100)
	ma2 := NewModuleActivation("test2", "package2", "version2", "modulefilepath2", -1)
	c.Add(ma1)
	c.Add(ma2)

	tdt := []struct {
		name string
		want bool
		in   *ModuleActivation
	}{
		{
			"unique activation should be ready",
			true,
			NewModuleActivation("super", "unique", "yeah", "cool", 5),
		},
		{
			"existing activation inside timeout should not be ready",
			false,
			ma1,
		},
		{
			"existing activation after timeout should be ready",
			true,
			ma2,
		},
	}
	for _, tt := range tdt {
		t.Run(tt.name, func(t *testing.T) {
			got := c.ReadyToWrite(tt.in)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModuleCacheClean(t *testing.T) {
	var c *ModuleCache

	tempCachePath, err := os.CreateTemp("", "testcache")
	if err != nil {
		t.Errorf("failed to create temp cache: %v", err)
	}

	c = NewModuleCache(tempCachePath.Name())
	ma1 := NewModuleActivation("testuser1", "testpackage1", "testversion1", "testmodulefilepath1", 100)
	ma2 := NewModuleActivation("test2", "package2", "version2", "modulefilepath2", -1)
	c.Add(ma1)
	c.Add(ma2)

	t.Run("temp cache has 2 activations before clean", func(t *testing.T) {
		numActivations := len(c.Activations)
		if numActivations != 2 {
			t.Errorf("want 2, got %v", numActivations)
		}
	})

	c.Clean()

	t.Run("temp cache has 1 activation after clean", func(t *testing.T) {
		numActivations := len(c.Activations)
		if numActivations != 1 {
			t.Errorf("want 1, got %v", numActivations)
		}
	})
}

type LogEntry struct {
	User string `json:"user"`
	Package string `json:"package"`
	Version string `json:"version"`
	Path string `json:"path"`
}

func TestLog(t *testing.T) {
	tdt := []struct {
		name string
		want LogEntry
		in   *ModuleActivation
	}{
		{
			"unique activation should be ready",
			LogEntry{"super", "unique", "yeah", "cool"},
			NewModuleActivation("super", "unique", "yeah", "cool", 5),
		},
	}
	for _, tt := range tdt {
		t.Run(tt.name, func(t *testing.T) {
			var b bytes.Buffer
			w := bufio.NewWriter(&b)

			Log(w, tt.in)

			w.Flush()

			var le LogEntry
			err := json.Unmarshal(b.Bytes(), &le)
			if err != nil {
				t.Errorf("failed to unmarshal log entry: %v", err)
			}

			if tt.want.User != le.User {
				t.Errorf("user field doesn't match. got %s want %s", le.User, tt.want.User)
			}
			if tt.want.Package != le.Package {
				t.Errorf("package field doesn't match. got %s want %s", le.Package, tt.want.Package)
			}
			if tt.want.Version != le.Version {
				t.Errorf("version field doesn't match. got %s want %s", le.Version, tt.want.Version)
			}
			if tt.want.Path != le.Path {
				t.Errorf("path field doesn't match. got %s want %s", le.Path, tt.want.Path)
			}
		})
	}
}
