package main

import (
	"os"

	"github.com/xylonx/sign-fxxker/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
