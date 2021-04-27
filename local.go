package filestore

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
)

// FilestoreLocal implements Filestore using a local directory.
type FilestoreLocal struct {
	// Dir is the root directory for the filestore.
	Dir string

	// FileMode holds the permission bits given to new files.
	FilePerm fs.FileMode

	// DirMode holds the permission bits given to new directories.
	DirPerm fs.FileMode
}

func (f *FilestoreLocal) makeAbsLocalPath(path string) string {
	return filepath.Join(f.Dir, path)
}

// NewFilestoreLocal creates a new FilestoreLocal initialized with the given
// root directory.
func NewFilestoreLocal(config ...LocalConfigFunc) (*FilestoreLocal, error) {
	f := &FilestoreLocal{
		Dir:      "",
		FilePerm: 0644,
		DirPerm:  0755,
	}
	for _, c := range config {
		c(f)
	}

	if f.Dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("couldn't set filestore root directory: %v", err)
		}
		f.Dir = cwd
	}

	if stat, err := os.Stat(f.Dir); err != nil {
		return nil, err
	} else if !stat.IsDir() {
		return nil, fmt.Errorf("'%s' is not a directory", f.Dir)
	}

	return f, nil
}

func (f *FilestoreLocal) Get(_ context.Context, path string) (File, error) {
	file, err := os.Open(f.makeAbsLocalPath(path))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if stat.IsDir() {
		return nil, ErrFileNotFound
	}
	return file, nil
}

func (f *FilestoreLocal) Insert(_ context.Context, r io.Reader, dest string) error {
	destAbs := f.makeAbsLocalPath(dest)
	destDir := filepath.Dir(destAbs)
	if err := os.MkdirAll(destDir, f.DirPerm.Perm()); err != nil {
		return err
	}

	// Write to temporary file first, in case dest already exists and is currently
	// in use. Use the same directory as dest to ensure both files are stored on
	// the same device.
	tmpFile, err := ioutil.TempFile(destDir, ".tmp_*")
	if err != nil {
		return err
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name()) // in case error occurs before file is moved

	if err := tmpFile.Chmod(f.FilePerm.Perm()); err != nil {
		return err
	}
	w := bufio.NewWriter(tmpFile)
	if _, err := w.ReadFrom(r); err != nil {
		return err
	}
	if err := w.Flush(); err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	// Safely move the file to dest.
	if err := os.Rename(tmpFile.Name(), destAbs); err != nil {
		return err
	}

	return nil
}

func (f *FilestoreLocal) Remove(_ context.Context, path string) error {
	err := os.Remove(f.makeAbsLocalPath(path))
	if errors.Is(err, fs.ErrNotExist) {
		return ErrFileNotFound
	}
	return nil
}

func (f *FilestoreLocal) Close() error {
	return nil
}
