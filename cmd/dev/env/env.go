package env

import (
	"bam/cmd"

	"github.com/spf13/cobra"
)

var EnvCmd = &cobra.Command{
	Use:   "env",
	Short: "Environment inspection and tool verification",
}

func init() {
	cmd.RegisterSubcommand(EnvCmd)
}
