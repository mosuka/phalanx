package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "phalanx-test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer os.RemoveAll(tmpDir)

	path := filepath.FromSlash(tmpDir)

	exists := FileExists(path)
	expected := true
	if exists != expected {
		t.Fatalf("expected %v, but %v\n", expected, exists)
	}
}

func TestIsFile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "phalanx-test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer os.RemoveAll(tmpDir)

	path := filepath.FromSlash(tmpDir)

	filePath := filepath.Join(path, "empty.txt")
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer file.Close()

	isFile := IsFile(filePath)
	expected := true
	if isFile != expected {
		t.Fatalf("expected %v, but %v\n", expected, isFile)
	}
}

func TestIsDie(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "phalanx-test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer os.RemoveAll(tmpDir)

	path := filepath.FromSlash(tmpDir)

	isDir := IsDir(path)
	expected := true
	if isDir != expected {
		t.Fatalf("expected %v, but %v\n", expected, isDir)
	}
}
