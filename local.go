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
	dir string
}

func (l *FilestoreLocal) makeLocalPath(path string) string {
	return filepath.Join(l.dir, path)
}

// NewFilestoreLocal creates a new FilestoreLocal initialized with the given
// root directory.
func NewFilestoreLocal(dir string) (*FilestoreLocal, error) {
	if stat, err := os.Stat(dir); err != nil {
		return nil, err
	} else if !stat.IsDir() {
		return nil, fmt.Errorf("'%s' is not a directory", dir)
	}
	return &FilestoreLocal{dir}, nil
}

func (l *FilestoreLocal) Get(_ context.Context, path string) (File, error) {
	f, err := os.Open(l.makeLocalPath(path))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if stat.IsDir() {
		return nil, ErrFileNotFound
	}
	return f, nil
}

func (l *FilestoreLocal) Insert(_ context.Context, r io.Reader, dest string) error {
	// Write to temporary file first, in case dest already exists and is currently
	// in use. Use the same directory as dest to ensure both files are stored on
	// the same device.
	tmpFile, err := ioutil.TempFile(l.dir, ".tmp_*")
	if err != nil {
		return err
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name()) // in case error occurs before file is moved

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

	destAbs := l.makeLocalPath(dest)

	// Make sure parent dirs exist.
	if err := os.MkdirAll(filepath.Dir(destAbs), 0755); err != nil {
		return err
	}

	// Safely move the file to dest.
	if err := os.Rename(tmpFile.Name(), destAbs); err != nil {
		return err
	}

	return nil
}

func (l *FilestoreLocal) Remove(_ context.Context, path string) error {
	err := os.Remove(l.makeLocalPath(path))
	if errors.Is(err, fs.ErrNotExist) {
		return ErrFileNotFound
	}
	return nil
}

func (l *FilestoreLocal) Close() error {
	return nil
}
