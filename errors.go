package filestore

type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrFileNotFound Error = "file was not found"
)
