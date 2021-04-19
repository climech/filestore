// Package filestore provides a simple interface for managing a store of files,
// with local and remote store implementations.
package filestore

import (
	"context"
	"io"
)

type Filestore interface {
	// Get retrieves a file from the store and returns a read-only File
	// descriptor. The path is relative to the store root. File.Close should be
	// called as soon as the file is no longer in use. ErrFileNotFound is returned
	// if path does not exist.
	Get(ctx context.Context, path string) (File, error)

	// Insert inserts a file into the store. The destination path is relative to
	// the store root. Parent directories are created automatically if needed. If
	// dest already exists, the file is safely overwritten.
	Insert(ctx context.Context, src io.Reader, dest string) error

	// Remove permanently deletes the file from the store. The path is relative to
	// the store root. ErrFileNotFound is returned if path does not exist.
	Remove(ctx context.Context, path string) error

	// Close frees up any resources currently in use.
	Close() error
}
