package main

import (
	"fmt"
	"os"

	"github.com/deeone/go-research/cmd/outreach"
)

func main() {
	if err := outreach.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}