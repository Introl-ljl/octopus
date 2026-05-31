package remote

import (
	"fmt"

	"github.com/bestruirui/octopus/internal/cli"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check current authentication status",
	Long:  `Verify the current CLI session is authenticated and the server is reachable.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		session := cli.GetSession()
		if session == nil {
			fmt.Println("Not authenticated. Use 'octopus remote login' to authenticate.")
			return nil
		}

		profile := cli.GetActiveProfile()
		if profile == nil {
			fmt.Println("No active profile.")
			return nil
		}

		client := cli.NewClient(profile.BaseURL)
		client.SetToken(session.Token)

		var result map[string]interface{}
		err := client.Get("/api/v1/user/status", &result)
		if err != nil {
			if _, ok := err.(*cli.CLIError); ok {
				fmt.Println("Session expired or invalid. Use 'octopus remote login' to re-authenticate.")
				cli.ClearSession()
				return nil
			}
			return err
		}

		fmt.Println("Authenticated as:", session.Username)
		fmt.Println("Profile:", profile.Name)
		fmt.Println("Server:", profile.BaseURL)
		if session.ExpireAt != "" {
			fmt.Println("Token expires at:", session.ExpireAt)
		}
		return nil
	},
}

func init() {
	RemoteCmd.AddCommand(statusCmd)
}
