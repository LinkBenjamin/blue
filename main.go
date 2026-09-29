package main

import (
	"bam/cmd"
	_ "bam/cmd/dev"
)

func main() {
	cmd.Execute()
}
