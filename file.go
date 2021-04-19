package filestore

type File interface {
	Read(b []byte) (int, error)
	Close() error
}
