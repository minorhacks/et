package exampleapi

import (
	"github.com/minorhacks/et"
	"github.com/minorhacks/et/sample/examplefs"
)

var files = map[int]string{
	0: "/var/log/foo.log",
	1: "/tmp/missing",
	2: "/tmp/some_file",
	3: "/tmp/some_other_file",
}

type ApiErr struct {
	et.Namespace
}

type ErrOversizedA struct {
	et.Member[ApiErr]
}

type ErrInternal struct {
	et.Member[ApiErr]
}

func MethodFoo(a, b int) (string, error) {
	if a > b {
		return "", et.Errorf[ErrOversizedA]("a (%d) is larger than b (%d)", a, b)
	}
	if a > 8 {
		return "", et.Errorf[ErrOversizedA]("a is larger than max supported value")
	}

	contents, err := examplefs.ReadFile(files[(a+b)%len(files)])
	if err != nil {
		return "", et.Errorf[ErrInternal]("a + b = %d resulted in error: %w", a+b, err)
	}

	return contents, nil
}
