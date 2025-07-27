package examplefs

import (
	"strings"

	"github.com/minorhacks/et"
)

type FsErr struct {
	et.Namespace
}

type ErrUnsupported struct {
	et.Member[FsErr]
}

type ErrNotFound struct {
	et.Member[FsErr]
}

func ReadFile(filename string) (string, error) {
	if !strings.HasPrefix(filename, "/tmp/") {
		return "", et.Errorf[ErrUnsupported]("only /tmp can be read")
	}

	if filename == "/tmp/missing" {
		return "", et.Errorf[ErrNotFound]("file not found: %q", filename)
	}

	return strings.TrimPrefix(filename, "/tmp/"), nil
}
