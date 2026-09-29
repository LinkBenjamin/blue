package dev

import (
	"bam/cmd"
	"bam/cmd/dev/env"
	"bam/cmd/dev/spin"

	"github.com/spf13/cobra"
)

var DevCmd = &cobra.Command{
	Use:   "dev",
	Short: "Pre-packaged Development Templates",
}

var EnvCmd = &cobra.Command{
	Use:   "env",
	Short: "Local Dev Environment Utilities",
}

func init() {
	cmd.RegisterSubcommand(DevCmd)
	DevCmd.AddCommand(spin.SpinCmd) // Attach spin under dev
	EnvCmd.AddCommand(env.EnvCmd)   // Attach env under dev
}
