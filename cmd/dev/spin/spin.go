package spin

import (
	"github.com/spf13/cobra"
)

var SpinCmd = &cobra.Command{
	Use:   "spin",
	Short: "Spin up a development environment/template",
}

func init() {
	// Register child leaf nodes under SpinCmd
	SpinCmd.AddCommand(pythonFastApiCmd)
	SpinCmd.AddCommand(s3Cmd)
	SpinCmd.AddCommand(vpcCmd)
}
