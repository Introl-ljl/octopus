package remote

import (
	"fmt"

	"github.com/bestruirui/octopus/internal/cli"
	"github.com/spf13/cobra"
)

var apikeyCmd = &cobra.Command{
	Use:   "apikey",
	Short: "Manage API keys",
	Long:  `Manage relay API keys: list, create, update, delete, and view stats.`,
}

var apikeyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all API keys",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			var keys []map[string]interface{}
			if err := c.Get("/api/v1/apikey/list", &keys); err != nil {
				return err
			}
			f.PrintHeader("ID", "Name", "Key", "Enabled", "Expire")
			for _, k := range keys {
				keyStr := fmt.Sprintf("%v", k["api_key"])
				f.PrintRow(
					fmt.Sprintf("%v", k["id"]),
					fmt.Sprintf("%v", k["name"]),
					cli.RedactFields(map[string]interface{}{"api_key": keyStr})["api_key"].(string),
					fmt.Sprintf("%v", k["enabled"]),
					fmt.Sprintf("%v", k["expire_at"]),
				)
			}
			f.Flush()
			return nil
		})
	},
}

var apikeyCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")

		if err := cli.RequireFlag(name, "name"); err != nil {
			return err
		}

		body := map[string]interface{}{
			"name": name,
		}

		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			var result map[string]interface{}
			if err := c.Post("/api/v1/apikey/create", body, &result); err != nil {
				return err
			}
			fmt.Printf("API Key created:\n")
			fmt.Printf("  Name:  %v\n", result["name"])
			fmt.Printf("  Key:   %v\n", result["api_key"])
			fmt.Println("Save the key now — it won't be shown again.")
			return nil
		})
	},
}

var apikeyDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetInt("id")
		if id == 0 {
			return cli.NewCLIError(cli.ExitCodeInput, "--id is required")
		}

		if !force && !confirmDestructive(fmt.Sprintf("Delete API key %d?", id)) {
			fmt.Println("Cancelled.")
			return nil
		}

		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			return c.Delete(fmt.Sprintf("/api/v1/apikey/delete/%d", id), nil)
		})
	},
}

var apikeyUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetInt("id")
		if id == 0 {
			return cli.NewCLIError(cli.ExitCodeInput, "--id is required")
		}
		body := map[string]interface{}{"id": id}
		if cmd.Flags().Changed("name") {
			body["name"], _ = cmd.Flags().GetString("name")
		}
		if cmd.Flags().Changed("enabled") {
			body["enabled"], _ = cmd.Flags().GetBool("enabled")
		}
		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			var result map[string]interface{}
			if err := c.Post("/api/v1/apikey/update", body, &result); err != nil {
				return err
			}
			f.PrintSuccess(fmt.Sprintf("API key %d updated.", id))
			return nil
		})
	},
}

var apikeyStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show API key statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			var result map[string]interface{}
			if err := c.Get("/api/v1/apikey/stats", &result); err != nil {
				return err
			}
			f.PrintHeader("Metric", "Value")
			for k, v := range result {
				f.PrintRow(k, fmt.Sprintf("%v", v))
			}
			f.Flush()
			return nil
		})
	},
}

func init() {
	apikeyCreateCmd.Flags().String("name", "", "API key name")
	apikeyDeleteCmd.Flags().Int("id", 0, "API key ID to delete")
	apikeyUpdateCmd.Flags().Int("id", 0, "API key ID")
	apikeyUpdateCmd.Flags().String("name", "", "API key name")
	apikeyUpdateCmd.Flags().Bool("enabled", true, "enable or disable")
	apikeyCmd.AddCommand(apikeyListCmd)
	apikeyCmd.AddCommand(apikeyCreateCmd)
	apikeyCmd.AddCommand(apikeyUpdateCmd)
	apikeyCmd.AddCommand(apikeyDeleteCmd)
	apikeyCmd.AddCommand(apikeyStatsCmd)
	RemoteCmd.AddCommand(apikeyCmd)
}
