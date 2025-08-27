package main

import (
	"os"

	"github.com/Kostayne/go-nginx-analyzer/cmd"
	"github.com/Kostayne/go-nginx-analyzer/generator"
)

func main() {
	if len(os.Args) <= 1 || os.Args[1] != "--gen" {
		cmd.Execute()
	} else {
		generator.GenerateAccessLog()
	}
}
