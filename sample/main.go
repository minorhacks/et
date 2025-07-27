package main

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/minorhacks/et"
	"github.com/minorhacks/et/sample/exampleapi"
	"github.com/minorhacks/et/sample/examplefs"
)

func main() {
	apiErrCounts := map[string]int{}
	fsErrCounts := map[string]int{}

	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			_, err := exampleapi.MethodFoo(i, j)
			if err == nil {
				slog.Info("Success", slog.Int("a", i), slog.Int("b", j))
				continue
			}

			apiErr := et.AsKind[exampleapi.ApiErr]()
			if errors.As(err, &apiErr) {
				apiErrCounts[apiErr.Tag()]++
			}

			fsErr := et.AsKind[examplefs.FsErr]()
			if errors.As(err, &fsErr) {
				fsErrCounts[fsErr.Tag()]++
			}

			for err != nil {
				if errors.Is(err, et.OfType[exampleapi.ErrOversizedA]()) {
					slog.Error("Bad params", slog.Int("a", i), slog.Int("b", j))
				} else if errors.Is(err, et.OfKind[examplefs.FsErr]()) {
					slog.Error("Underlying failure", slog.Any("err", err))
				}
				err = errors.Unwrap(err)
			}
		}
	}
	fmt.Println("exampleapi error counts")
	for k, v := range apiErrCounts {
		fmt.Printf("%s: %d\n", k, v)
	}
	fmt.Println("")

	fmt.Println("examplefs error counts")
	for k, v := range fsErrCounts {
		fmt.Printf("%s: %d\n", k, v)
	}
	fmt.Println("")
}
