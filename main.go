package main

import (
	"os"

	"github.com/kostayne/go-nginx-analyzer/cmd"
	"github.com/kostayne/go-nginx-analyzer/generator"
)

func main() {
	if len(os.Args) <= 1 || os.Args[1] != "--gen" {
		cmd.Execute()
	} else {
		generator.GenerateAccessLog()
	}
}
