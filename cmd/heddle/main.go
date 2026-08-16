// Command heddle reads a weaving draft and reports the cloth it makes.
package main

import (
	"os"

	"Heddle/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
