package urfs

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

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
