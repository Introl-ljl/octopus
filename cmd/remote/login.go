package remote

import (
	"fmt"

	"github.com/bestruirui/octopus/internal/cli"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the Octopus server",
	Long: `Authenticate with the Octopus server using management credentials.
Stores the session token in the active profile for subsequent commands.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		var expire int
		if cmd.Flags().Changed("expire") {
			expire, _ = cmd.Flags().GetInt("expire")
		}

		if err := cli.RequireFlag(username, "username"); err != nil {
			return err
		}

		profile := cli.GetActiveProfile()
		if profile == nil {
			return cli.NewCLIError(cli.ExitCodeInput, "no active profile. Configure one with 'octopus remote profile add'")
		}

		baseURL := profile.BaseURL
		if serverURL != "" {
			baseURL = serverURL
		}
		client := cli.NewClient(baseURL)

		var result struct {
			Token    string `json:"token"`
			ExpireAt string `json:"expire_at"`
		}
		body := map[string]interface{}{
			"username": username,
			"password": password,
		}
		if cmd.Flags().Changed("expire") {
			body["expire"] = expire
		}

		err := client.Post("/api/v1/user/login", body, &result)
		if err != nil {
			return err
		}

		cli.SaveSession(cli.Session{
			ProfileName: profile.Name,
			Username:    username,
			Token:       result.Token,
			ExpireAt:    result.ExpireAt,
		})

		fmt.Println("Login successful. Session token stored for profile:", profile.Name)
		if result.ExpireAt != "" {
			fmt.Println("Token expires at:", result.ExpireAt)
		}
		return nil
	},
}

func init() {
	loginCmd.Flags().String("username", "", "management username")
	loginCmd.Flags().String("password", "", "management password")
	loginCmd.Flags().Int("expire", 0, "token expiration in minutes (0=default 15, -1=max 30 days)")
	RemoteCmd.AddCommand(loginCmd)
}
