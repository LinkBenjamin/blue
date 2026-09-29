package env

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// Map tool aliases to executable binary names
var toolBinaries = map[string][]string{
	"nodejs": {"node"},
	"python": {"python3", "python", "py"},
	"cobol":  {"cobc"},
}

var verifyCmd = &cobra.Command{
	Use:       "verify [tool]",
	Short:     "Verify present development tools on the workstation",
	ValidArgs: []string{"nodejs", "python", "cobol"},
	Args:      cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			tool := strings.ToLower(args[0])
			binaries, valid := toolBinaries[tool]
			if !valid {
				return fmt.Errorf("invalid tool '%s'. Allowed options: nodejs, python, cobol", tool)
			}

			// Single tool check: Output true or false
			found := isToolInstalled(binaries)
			fmt.Println(found)
			return nil
		}

		// Omitted argument: Verbose summary response
		fmt.Println("Environment Tools Summary:")
		fmt.Println("--------------------------")
		for tool, binaries := range toolBinaries {
			if isToolInstalled(binaries) {
				fmt.Printf("  [FOUND]     %s\n", tool)
			} else {
				fmt.Printf("  [NOT FOUND] %s\n", tool)
			}
		}
		return nil
	},
}

func init() {
	EnvCmd.AddCommand(verifyCmd)
}

// isToolInstalled checks if any executable matching the tool list exists in PATH
func isToolInstalled(binaries []string) bool {
	for _, bin := range binaries {
		if _, err := exec.LookPath(bin); err == nil {
			return true
		}
	}
	return false
}
