package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
)

func Prepare(p string, overwrite bool) (io.ReadWriteCloser, error) {
	_, err := os.Stat(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if path.Base(p) != p {
				dir := path.Dir(p)
				if err = os.MkdirAll(dir, 0700); err != nil {
					return nil, fmt.Errorf("mkdir all %q: %w", dir, err)
				}
			}
			return create(p)
		}
		return nil, fmt.Errorf("stat %q: %w", p, err)
	}
	stat, err := os.Stat(p)
	if err != nil {
		return nil, fmt.Errorf("exists stat %q: %w", p, err)
	}
	if stat.IsDir() {
		return nil, fmt.Errorf("%q %w", p, ErrConfigFileIsDir)
	}
	if overwrite {
		return create(p)
	}
	f, err := os.Open(p)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", p, err)
	}
	return f, nil
}

var ErrConfigFileIsDir = errors.New("is a directory")

func create(p string) (io.ReadWriteCloser, error) {
	f, err := os.Create(p)
	if err != nil {
		return nil, fmt.Errorf("create %q: %w", p, err)
	}
	return f, nil
}
