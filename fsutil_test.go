package urfs

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestPathExists(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "com.bengfort.urfs-")
	if err != nil {
		t.Error(err.Error())
	}

	defer os.RemoveAll(tmpdir)

	// create a random directory path that doesn't exist
	path := filepath.Join(tmpdir, "foo")

	if PathExists(path) {
		t.Fatal("a non-existant path returned true")
	}

	// create a file
	if err := ioutil.WriteFile(path, []byte("bar"), 0644); err != nil {
		t.Error(err.Error())
	}

	// ensure the file path exists
	if !PathExists(path) {
		t.Fatal("an existing path returned false")
	}

	// create a directory
	path = filepath.Join(tmpdir, "baz")
	if err := os.Mkdir(path, 0755); err != nil {
		t.Error(err.Error())
	}

	// ensure the directory path exists
	if !PathExists(path) {
		t.Fatal("an existing path returned false")
	}
}

// TestMkdir tests the creation of a single file and ensures no error is
// raised if the directory already exists.
func TestMkdir(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "com.bengfort.urfs-")
	if err != nil {
		t.Error(err.Error())
	}

	defer os.RemoveAll(tmpdir)

	// create a random directory path
	path := filepath.Join(tmpdir, "testing123")

	// check the path does not exist
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("%s alaready exists", path)
	}

	// create the directory
	if err := Mkdir(path); err != nil {
		t.Fatal(err.Error())
	}

	// check the path does exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("%s not correctly created", path)
	}

	// create again, but no error should occur
	if err := Mkdir(path); err != nil {
		t.Fatal(err.Error())
	}

	// check the path does exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("%s not correctly created", path)
	}
}

// TestMkdirAll tests to ensure a deep path can be created
func TestMkdirAll(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "com.bengfort.urfs-")
	if err != nil {
		t.Error(err.Error())
	}

	defer os.RemoveAll(tmpdir)

	// create a random deep directory path
	path := filepath.Join(tmpdir, "path", "to", "testing123")

	// check the path does not exist
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("%s alaready exists", path)
	}

	// create the directory
	if err := Mkdir(path); err != nil {
		t.Fatal(err.Error())
	}

	// check the path does exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("%s not correctly created", path)
	}

	// create again, but no error should occur
	if err := Mkdir(path); err != nil {
		t.Fatal(err.Error())
	}

	// check the path does exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("%s not correctly created", path)
	}
}
