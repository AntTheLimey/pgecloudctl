package main

import (
	"errors"
	"os"

	"github.com/AntTheLimey/pgecloudctl/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		var ee *cmd.ExitError
		if errors.As(err, &ee) {
			os.Exit(ee.Code())
		}
		os.Exit(1)
	}
}
