package remote

import (
	"fmt"

	"github.com/bestruirui/octopus/internal/cli"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear current session",
	Long:  `Remove the stored authentication token for the active profile.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		session := cli.GetSession()
		if session == nil {
			fmt.Println("No active session.")
			return nil
		}
		cli.ClearSession()
		fmt.Println("Logged out. Session cleared for active profile.")
		return nil
	},
}

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage user account",
	Long:  `Change management password or username.`,
}

var changePasswordCmd = &cobra.Command{
	Use:   "change-password",
	Short: "Change management password",
	RunE: func(cmd *cobra.Command, args []string) error {
		oldPass, _ := cmd.Flags().GetString("old-password")
		newPass, _ := cmd.Flags().GetString("new-password")

		if err := cli.RequireFlag(oldPass, "old-password"); err != nil {
			return err
		}
		if err := cli.RequireFlag(newPass, "new-password"); err != nil {
			return err
		}

		body := map[string]interface{}{
			"old_password": oldPass,
			"new_password": newPass,
		}

		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			if err := c.Post("/api/v1/user/change-password", body, nil); err != nil {
				return err
			}
			f.PrintSuccess("Password changed successfully.")
			return nil
		})
	},
}

var changeUsernameCmd = &cobra.Command{
	Use:   "change-username",
	Short: "Change management username",
	RunE: func(cmd *cobra.Command, args []string) error {
		newUsername, _ := cmd.Flags().GetString("new-username")

		if err := cli.RequireFlag(newUsername, "new-username"); err != nil {
			return err
		}

		body := map[string]interface{}{
			"new_username": newUsername,
		}

		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			if err := c.Post("/api/v1/user/change-username", body, nil); err != nil {
				return err
			}
			f.PrintSuccess("Username changed successfully.")
			return nil
		})
	},
}

func init() {
	changePasswordCmd.Flags().String("old-password", "", "current password")
	changePasswordCmd.Flags().String("new-password", "", "new password")
	changeUsernameCmd.Flags().String("new-username", "", "new username")

	userCmd.AddCommand(changePasswordCmd)
	userCmd.AddCommand(changeUsernameCmd)

	RemoteCmd.AddCommand(logoutCmd)
	RemoteCmd.AddCommand(userCmd)
}
