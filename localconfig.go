package filestore

import (
	"io/fs"
)

type LocalConfigFunc func(*FilestoreLocal)

func SetLocalDir(dir string) LocalConfigFunc {
	return func(f *FilestoreLocal) {
		f.Dir = dir
	}
}

func SetLocalFilePerm(perm fs.FileMode) LocalConfigFunc {
	return func(f *FilestoreLocal) {
		f.FilePerm = perm
	}
}

func SetLocalDirPerm(perm fs.FileMode) LocalConfigFunc {
	return func(f *FilestoreLocal) {
		f.DirPerm = perm
	}
}
