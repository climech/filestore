package filestore

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestFilestoreLocal(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	var store Filestore
	store, err = NewFilestoreLocal(dir)
	if err != nil {
		t.Fatal(err)
	}

	fileText := "Hello world!\n"
	filePath := filepath.Join("subdir", "test.txt")
	ctx := context.Background()

	if err := store.Insert(ctx, strings.NewReader(fileText), filePath); err != nil {
		t.Fatal(err)
	}

	// Check if the file was saved correctly on the filesystem.
	if b, err := ioutil.ReadFile(filepath.Join(dir, filePath)); err != nil {
		t.Fatal(err)
	} else if string(b) != fileText {
		t.Fatalf(`file contents mismatch; got "%v",  want "%v"`, string(b), fileText)
	}

	// Get an existing file.
	if f, err := store.Get(ctx, filePath); err != nil {
		t.Errorf("couldn't get file '%s': %v", filePath, err)
	} else {
		defer f.Close()
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(f); err != nil {
			t.Errorf("couldn't read data from File: %v", err)
		}
		text := buf.String()
		if text != fileText {
			t.Errorf(`file contents mismatch; got "%v",  want "%v"`, text, fileText)
		}
	}

	// Get a non-existent file.
	if _, err := store.Get(ctx, "non-existent.txt"); err == nil {
		t.Errorf("no error returned when trying to get a non-existent file")
	} else if err != ErrFileNotFound {
		t.Errorf("wrong error when trying to get a non-existent file; "+
			"want ErrFileNotFound, got err of type %s: %v", reflect.TypeOf(err), err)
	}

	// Delete existing file.
	if err := store.Remove(ctx, filePath); err != nil {
		t.Errorf("couldn't delete file '%s': %v", filePath, err)
	} else if _, err := os.Stat(filepath.Join(dir, filePath)); os.IsExist(err) {
		t.Errorf("file '%s' exists after deletion", filePath)
	}

	// Delete non-existent file.
	if err := store.Remove(ctx, "non-existent.txt"); err == nil {
		t.Errorf("no error returned when trying to delete a non-existent file")
	} else if err != ErrFileNotFound {
		t.Errorf("wrong error when trying to remove a non-existent file; "+
			"want ErrFileNotFound, got err of type %s: %v", reflect.TypeOf(err), err)
	}
}
