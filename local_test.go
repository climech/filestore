package filestore_test

import (
	"bytes"
	"context"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/climech/filestore"
)

func TestFilestoreLocal(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	filePerm := fs.FileMode(0601)
	dirPerm := fs.FileMode(0701)
	store, err := filestore.NewFilestoreLocal(
		filestore.SetLocalDir(dir),
		filestore.SetLocalFilePerm(filePerm),
		filestore.SetLocalDirPerm(dirPerm),
	)
	if err != nil {
		t.Fatal(err)
	}

	text := "Hello world!\n"
	fp := filepath.Join("subdir", "test.txt")
	ctx := context.Background()

	if err := store.Insert(ctx, strings.NewReader(text), fp); err != nil {
		t.Fatal(err)
	}

	// Check if the file was saved correctly on the filesystem.
	if b, err := ioutil.ReadFile(filepath.Join(dir, fp)); err != nil {
		t.Fatal(err)
	} else if string(b) != text {
		t.Fatalf(`file contents mismatch; got "%v",  want "%v"`, string(b), text)
	}

	checkPermissions := func(fp string, perm fs.FileMode) {
		info, err := os.Stat(fp)
		if err != nil {
			t.Fatal(err)
		}
		permWant := perm.Perm()
		permGot := info.Mode().Perm()
		if permWant != permGot {
			t.Errorf(`incorrect permissions for '%s': want %s, got %s`,
				fp, permWant, permGot)
		}
	}
	checkPermissions(filepath.Join(dir, fp), filePerm)
	checkPermissions(filepath.Join(dir, filepath.Dir(fp)), dirPerm)

	// Get an existing file.
	if f, err := store.Get(ctx, fp); err != nil {
		t.Errorf("couldn't get file '%s': %v", fp, err)
	} else {
		defer f.Close()
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(f); err != nil {
			t.Errorf("couldn't read data from File: %v", err)
		}
		s := buf.String()
		if s != text {
			t.Errorf(`file contents mismatch; got "%v",  want "%v"`, s, text)
		}
	}

	// Get a non-existent file.
	if _, err := store.Get(ctx, "non-existent.txt"); err == nil {
		t.Errorf("no error returned when trying to get a non-existent file")
	} else if err != filestore.ErrFileNotFound {
		t.Errorf("wrong error when trying to get a non-existent file; "+
			"want ErrFileNotFound, got err of type %s: %v", reflect.TypeOf(err), err)
	}

	// Delete existing file.
	if err := store.Remove(ctx, fp); err != nil {
		t.Errorf("couldn't delete file '%s': %v", fp, err)
	} else if _, err := os.Stat(filepath.Join(dir, fp)); os.IsExist(err) {
		t.Errorf("file '%s' exists after deletion", fp)
	}

	// Delete non-existent file.
	if err := store.Remove(ctx, "non-existent.txt"); err == nil {
		t.Errorf("no error returned when trying to delete a non-existent file")
	} else if err != filestore.ErrFileNotFound {
		t.Errorf("wrong error when trying to remove a non-existent file; "+
			"want ErrFileNotFound, got err of type %s: %v", reflect.TypeOf(err), err)
	}
}
