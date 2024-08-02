package main

import (
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

// func TestModuleCacheReadyToWrite(t *testing.T) {
// 	var c *ModuleCache
//
// 	tempCachePath, err := os.CreateTemp("", "testcache")
// 	if err != nil {
// 		t.Errorf("failed to create temp cache: %v", err)
// 	}
//
// 	c = NewModuleCache(tempCachePath.Name())
// 	ma := NewModuleActivation("testuser2", "testpackage2", "testversion2", "testmodulefilepath2", 100)
// }
