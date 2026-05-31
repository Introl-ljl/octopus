package remote

import (
	"fmt"

	"github.com/bestruirui/octopus/internal/cli"
	"github.com/spf13/cobra"
)

var settingCmd = &cobra.Command{
	Use:   "setting",
	Short: "Manage system settings",
	Long:  `View and modify system settings, and manage circuit breaker state.`,
}

var settingListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			var settings []map[string]interface{}
			if err := c.Get("/api/v1/setting/list", &settings); err != nil {
				return err
			}
			f.PrintHeader("Key", "Value")
			for _, s := range settings {
				f.PrintRow(
					fmt.Sprintf("%v", s["key"]),
					fmt.Sprintf("%v", s["value"]),
				)
			}
			f.Flush()
			return nil
		})
	},
}

var settingSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a setting value",
	RunE: func(cmd *cobra.Command, args []string) error {
		key, _ := cmd.Flags().GetString("key")
		value, _ := cmd.Flags().GetString("value")

		if err := cli.RequireFlag(key, "key"); err != nil {
			return err
		}
		if err := cli.RequireFlag(value, "value"); err != nil {
			return err
		}

		body := map[string]interface{}{
			"key":   key,
			"value": value,
		}

		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			if err := c.Post("/api/v1/setting/set", body, nil); err != nil {
				return err
			}
			f.PrintSuccess(fmt.Sprintf("Setting '%s' updated.", key))
			return nil
		})
	},
}

var circuitBreakerResetCmd = &cobra.Command{
	Use:   "reset-circuit-breaker",
	Short: "Reset circuit breaker state",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			var result map[string]interface{}
			if err := c.Post("/api/v1/setting/circuit-breaker/reset", struct{}{}, &result); err != nil {
				return err
			}
			f.PrintSuccess(fmt.Sprintf("Circuit breaker reset. Reset count: %v", result["reset_count"]))
			return nil
		})
	},
}

func init() {
	settingSetCmd.Flags().String("key", "", "setting key")
	settingSetCmd.Flags().String("value", "", "setting value")
	settingCmd.AddCommand(settingListCmd)
	settingCmd.AddCommand(settingSetCmd)
	settingCmd.AddCommand(circuitBreakerResetCmd)
	RemoteCmd.AddCommand(settingCmd)
}
