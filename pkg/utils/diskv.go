package utils

import (
	"os"
	"path/filepath"

	"github.com/peterbourgon/diskv/v3"
)

const (
	STROAGE_PREFIX = ".groundcover"
)

var PresistentStorage *diskv.Diskv = NewStorage()

func NewStorage() *diskv.Diskv {
	var err error

	var baseDir string
	if baseDir, err = os.UserHomeDir(); err != nil {
		baseDir = os.TempDir()
	}

	diskv := diskv.New(diskv.Options{
		BasePath:  filepath.Join(baseDir, STROAGE_PREFIX),
		Transform: func(s string) []string { return []string{} },
	})

	return diskv
}
