package storage

import "os"

// Int shows available actions for interface
type Int interface {
	OpenFile(path string, flag int, perm os.FileMode) (*os.File, error)
	CreateFile(path string) (*os.File, error)
	ResultPath(dir, name, ext string) string
}
