package main

import (
	"bam/cmd"
	_ "bam/cmd/dev"
	_ "bam/cmd/ticket"
)

func main() {
	cmd.Execute()
}
