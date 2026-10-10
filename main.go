package main

import (
	"os"

	"github.com/gomicro/align/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
