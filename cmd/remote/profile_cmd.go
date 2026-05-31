package remote

import (
	"fmt"

	"github.com/bestruirui/octopus/internal/cli"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage local connection profiles",
	Long:  `Configure and manage named connection profiles for remote server access.`,
}

var profileAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new connection profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		server, _ := cmd.Flags().GetString("url")
		output, _ := cmd.Flags().GetString("output")

		if err := cli.RequireFlag(name, "name"); err != nil {
			return err
		}
		if err := cli.RequireFlag(server, "server"); err != nil {
			return err
		}
		if err := cli.ValidateURL(server); err != nil {
			return err
		}
		if output == "" {
			output = "table"
		}
		if err := cli.ValidateEnum(output, []string{"table", "json"}, "output"); err != nil {
			return err
		}

		cli.AddProfile(name, server, output)
		fmt.Printf("Profile '%s' added. Server: %s\n", name, server)
		return nil
	},
}

var profileUseCmd = &cobra.Command{
	Use:   "use",
	Short: "Set a profile as active",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cli.SetActiveProfile(name)
		fmt.Printf("Profile '%s' is now active.\n", name)
		return nil
	},
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all connection profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		profiles := cli.GetProfiles()
		fmt := cli.GetFormatter()
		fmt.PrintHeader("Name", "Server", "Output", "Active")
		for _, p := range profiles {
			active := ""
			if p.Active {
				active = "*"
			}
			fmt.PrintRow(p.Name, p.BaseURL, p.DefaultOutput, active)
		}
		fmt.Flush()
		return nil
	},
}

var profileRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a connection profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cli.RemoveProfile(name)
		fmt.Printf("Profile '%s' removed.\n", name)
		return nil
	},
}

func init() {
	profileAddCmd.Flags().String("name", "", "profile name")
	profileAddCmd.Flags().String("url", "", "server URL (e.g., http://127.0.0.1:8080)")
	profileAddCmd.Flags().String("output", "table", "default output mode (table or json)")
	profileCmd.AddCommand(profileAddCmd)
	profileCmd.AddCommand(profileUseCmd)
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileRemoveCmd)
	RemoteCmd.AddCommand(profileCmd)
}
