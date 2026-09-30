package ticket

import (
	"fmt"
	"strings"

	"bam/cmd"

	"github.com/spf13/cobra"
)

var TicketCmd = &cobra.Command{
	Use:   "ticket",
	Short: "Create ServiceNow tickets",
}

var groupMembershipCmd = &cobra.Command{
	Use:   "group-membership <group-name> <user-name>",
	Short: "Request adding a user to a group in ServiceNow",
	Args:  cobra.ExactArgs(2),
	RunE: func(command *cobra.Command, args []string) error {
		groupName := strings.TrimSpace(args[0])
		userName := strings.TrimSpace(args[1])
		if groupName == "" || userName == "" {
			return fmt.Errorf("group name and user name cannot be empty")
		}

		requestID, err := createGroupMembershipRequest(command.Context(), serviceNowHTTPClient, serviceNowSettings, groupName, userName)
		if err != nil {
			return err
		}
		fmt.Printf("ServiceNow request ID: %s\n", requestID)
		return nil
	},
}

func init() {
	cmd.RegisterSubcommand(TicketCmd)
	TicketCmd.AddCommand(groupMembershipCmd)
}
