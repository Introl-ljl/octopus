package remote

import (
	"fmt"

	"github.com/bestruirui/octopus/internal/cli"
	"github.com/spf13/cobra"
)

var modelCmd = &cobra.Command{
	Use:   "model",
	Short: "Manage model prices",
	Long:  `Manage model pricing entries: list, create, update, delete, and sync pricing.`,
}

var modelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all model prices",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			var models []map[string]interface{}
			if err := c.Get("/api/v1/model/list", &models); err != nil {
				return err
			}
			f.PrintHeader("Name", "Input", "Output", "Cache Read", "Cache Write")
			for _, m := range models {
				f.PrintRow(
					fmt.Sprintf("%v", m["name"]),
					fmt.Sprintf("%v", m["input"]),
					fmt.Sprintf("%v", m["output"]),
					fmt.Sprintf("%v", m["cache_read"]),
					fmt.Sprintf("%v", m["cache_write"]),
				)
			}
			f.Flush()
			return nil
		})
	},
}

var modelCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a model pricing entry",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		inputPrice, _ := cmd.Flags().GetFloat64("input")
		outputPrice, _ := cmd.Flags().GetFloat64("output")

		if err := cli.RequireFlag(name, "name"); err != nil {
			return err
		}

		body := map[string]interface{}{
			"name":        name,
			"input":       inputPrice,
			"output":      outputPrice,
			"cache_read":  0,
			"cache_write": 0,
		}

		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			if err := c.Post("/api/v1/model/create", body, nil); err != nil {
				return err
			}
			f.PrintSuccess(fmt.Sprintf("Model '%s' created.", name))
			return nil
		})
	},
}

var modelDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a model pricing entry",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")

		if err := cli.RequireFlag(name, "name"); err != nil {
			return err
		}

		if !force && !confirmDestructive(fmt.Sprintf("Delete model '%s'?", name)) {
			fmt.Println("Cancelled.")
			return nil
		}

		body := map[string]interface{}{"name": name}
		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			return c.Post("/api/v1/model/delete", body, nil)
		})
	},
}

var modelUpdatePriceCmd = &cobra.Command{
	Use:   "update-price",
	Short: "Sync model pricing from upstream",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			if err := c.Post("/api/v1/model/update-price", struct{}{}, nil); err != nil {
				return err
			}
			f.PrintSuccess("Model pricing sync triggered.")
			return nil
		})
	},
}

var modelLastUpdateCmd = &cobra.Command{
	Use:   "last-update-time",
	Short: "Show last model price update time",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			var result string
			if err := c.Get("/api/v1/model/last-update-time", &result); err != nil {
				return err
			}
			fmt.Println("Last update time:", result)
			return nil
		})
	},
}

var modelChannelCmd = &cobra.Command{
	Use:   "channel",
	Short: "List model-to-channel associations",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			var results []map[string]interface{}
			if err := c.Get("/api/v1/model/channel", &results); err != nil {
				return err
			}
			f.PrintHeader("Model", "Channel ID", "Channel Name", "Enabled")
			for _, r := range results {
				f.PrintRow(
					fmt.Sprintf("%v", r["name"]),
					fmt.Sprintf("%v", r["channel_id"]),
					fmt.Sprintf("%v", r["channel_name"]),
					fmt.Sprintf("%v", r["enabled"]),
				)
			}
			f.Flush()
			return nil
		})
	},
}

var modelUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a model pricing entry",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		if err := cli.RequireFlag(name, "name"); err != nil {
			return err
		}
		body := map[string]interface{}{"name": name}
		if cmd.Flags().Changed("input") {
			body["input"], _ = cmd.Flags().GetFloat64("input")
		}
		if cmd.Flags().Changed("output") {
			body["output"], _ = cmd.Flags().GetFloat64("output")
		}
		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			if err := c.Post("/api/v1/model/update", body, nil); err != nil {
				return err
			}
			f.PrintSuccess(fmt.Sprintf("Model '%s' updated.", name))
			return nil
		})
	},
}

func init() {
	modelCreateCmd.Flags().String("name", "", "model name")
	modelCreateCmd.Flags().Float64("input", 0, "input price")
	modelCreateCmd.Flags().Float64("output", 0, "output price")
	modelDeleteCmd.Flags().String("name", "", "model name to delete")
	modelUpdateCmd.Flags().String("name", "", "model name")
	modelUpdateCmd.Flags().Float64("input", 0, "input price")
	modelUpdateCmd.Flags().Float64("output", 0, "output price")

	modelCmd.AddCommand(modelListCmd)
	modelCmd.AddCommand(modelChannelCmd)
	modelCmd.AddCommand(modelCreateCmd)
	modelCmd.AddCommand(modelUpdateCmd)
	modelCmd.AddCommand(modelDeleteCmd)
	modelCmd.AddCommand(modelUpdatePriceCmd)
	modelCmd.AddCommand(modelLastUpdateCmd)
	RemoteCmd.AddCommand(modelCmd)
}
