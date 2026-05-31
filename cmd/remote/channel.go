package remote

import (
	"fmt"
	"strconv"

	"github.com/bestruirui/octopus/internal/cli"
	"github.com/spf13/cobra"
)

var channelCmd = &cobra.Command{
	Use:   "channel",
	Short: "Manage channels",
	Long:  `Manage LLM provider channels: list, create, update, enable, delete, fetch models, and sync.`,
}

var channelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all channels",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			var channels []map[string]interface{}
			if err := c.Get("/api/v1/channel/list", &channels); err != nil {
				return err
			}
			f.PrintHeader("ID", "Name", "Type", "Enabled", "Model")
			for _, ch := range channels {
				f.PrintRow(
					fmt.Sprintf("%v", ch["id"]),
					fmt.Sprintf("%v", ch["name"]),
					fmt.Sprintf("%v", ch["type"]),
					fmt.Sprintf("%v", ch["enabled"]),
					fmt.Sprintf("%v", ch["model"]),
				)
			}
			f.Flush()
			return nil
		})
	},
}

var channelCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new channel",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		chType, _ := cmd.Flags().GetInt("type")
		model, _ := cmd.Flags().GetString("model")
		baseURL, _ := cmd.Flags().GetString("base-url")
		apiKey, _ := cmd.Flags().GetString("api-key")

		if err := cli.RequireFlag(name, "name"); err != nil {
			return err
		}
		if err := cli.RequireFlag(model, "model"); err != nil {
			return err
		}
		if err := cli.RequireFlag(baseURL, "base-url"); err != nil {
			return err
		}
		if err := cli.RequireFlag(apiKey, "api-key"); err != nil {
			return err
		}

		body := map[string]interface{}{
			"name": name,
			"type": chType,
			"model": model,
			"base_urls": []map[string]interface{}{
				{"url": baseURL, "delay": 0},
			},
			"keys": []map[string]interface{}{
				{"enabled": true, "channel_key": apiKey, "remark": ""},
			},
		}

		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			var result map[string]interface{}
			if err := c.Post("/api/v1/channel/create", body, &result); err != nil {
				return err
			}
			f.PrintSuccess(fmt.Sprintf("Channel '%s' created (ID: %v)", name, result["id"]))
			return nil
		})
	},
}

var channelEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable or disable a channel",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetInt("id")
		enabled, _ := cmd.Flags().GetBool("enabled")

		if id == 0 {
			return cli.NewCLIError(cli.ExitCodeInput, "--id is required")
		}

		body := map[string]interface{}{
			"id":      id,
			"enabled": enabled,
		}

		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			if err := c.Post("/api/v1/channel/enable", body, nil); err != nil {
				return err
			}
			f.PrintSuccess(fmt.Sprintf("Channel %d enabled: %v", id, enabled))
			return nil
		})
	},
}

var channelDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a channel",
	RunE: func(cmd *cobra.Command, args []string) error {
		idStr, _ := cmd.Flags().GetString("id")
		id, err := strconv.Atoi(idStr)
		if err != nil || id == 0 {
			return cli.NewCLIError(cli.ExitCodeInput, "--id is required and must be a number")
		}

		if !force && !confirmDestructive(fmt.Sprintf("Delete channel %d?", id)) {
			fmt.Println("Cancelled.")
			return nil
		}

		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			if err := c.Delete(fmt.Sprintf("/api/v1/channel/delete/%d", id), nil); err != nil {
				return err
			}
			f.PrintSuccess(fmt.Sprintf("Channel %d deleted.", id))
			return nil
		})
	},
}

var channelSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync channel model data",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			if err := c.Post("/api/v1/channel/sync", nil, nil); err != nil {
				return err
			}
			f.PrintSuccess("Channel sync triggered.")
			return nil
		})
	},
}

var channelLastSyncCmd = &cobra.Command{
	Use:   "last-sync-time",
	Short: "Show last channel sync time",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			var result string
			if err := c.Get("/api/v1/channel/last-sync-time", &result); err != nil {
				return err
			}
			fmt.Println("Last sync time:", result)
			return nil
		})
	},
}

var channelUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a channel",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetInt("id")
		if id == 0 {
			return cli.NewCLIError(cli.ExitCodeInput, "--id is required")
		}
		body := map[string]interface{}{"id": id}
		if cmd.Flags().Changed("name") {
			body["name"], _ = cmd.Flags().GetString("name")
		}
		if cmd.Flags().Changed("model") {
			body["model"], _ = cmd.Flags().GetString("model")
		}
		if cmd.Flags().Changed("enabled") {
			body["enabled"], _ = cmd.Flags().GetBool("enabled")
		}
		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			var result map[string]interface{}
			if err := c.Post("/api/v1/channel/update", body, &result); err != nil {
				return err
			}
			f.PrintSuccess(fmt.Sprintf("Channel %d updated.", id))
			return nil
		})
	},
}

var channelFetchModelCmd = &cobra.Command{
	Use:   "fetch-model",
	Short: "Fetch models for a channel configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		chType, _ := cmd.Flags().GetInt("type")
		baseURL, _ := cmd.Flags().GetString("base-url")
		apiKey, _ := cmd.Flags().GetString("api-key")

		if err := cli.RequireFlag(baseURL, "base-url"); err != nil {
			return err
		}
		if err := cli.RequireFlag(apiKey, "api-key"); err != nil {
			return err
		}

		body := map[string]interface{}{
			"type": chType,
			"base_urls": []map[string]interface{}{
				{"url": baseURL, "delay": 0},
			},
			"keys": []map[string]interface{}{
				{"enabled": true, "channel_key": apiKey},
			},
		}
		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			var models []string
			if err := c.Post("/api/v1/channel/fetch-model", body, &models); err != nil {
				return err
			}
			for _, m := range models {
				fmt.Println(m)
			}
			return nil
		})
	},
}

func init() {
	channelCreateCmd.Flags().String("name", "", "channel name")
	channelCreateCmd.Flags().Int("type", 0, "channel type (0=OpenAI Chat, 1=OpenAI Response, 2=Anthropic, 3=Gemini)")
	channelCreateCmd.Flags().String("model", "", "model name pattern")
	channelCreateCmd.Flags().String("base-url", "", "provider base URL")
	channelCreateCmd.Flags().String("api-key", "", "provider API key")
	channelEnableCmd.Flags().Int("id", 0, "channel ID")
	channelEnableCmd.Flags().Bool("enabled", true, "enable or disable")
	channelDeleteCmd.Flags().String("id", "", "channel ID to delete")
	channelUpdateCmd.Flags().Int("id", 0, "channel ID")
	channelUpdateCmd.Flags().String("name", "", "channel name")
	channelUpdateCmd.Flags().String("model", "", "model name pattern")
	channelUpdateCmd.Flags().Bool("enabled", false, "enable or disable")
	channelFetchModelCmd.Flags().Int("type", 0, "channel type")
	channelFetchModelCmd.Flags().String("base-url", "", "provider base URL")
	channelFetchModelCmd.Flags().String("api-key", "", "provider API key")

	channelCmd.AddCommand(channelListCmd)
	channelCmd.AddCommand(channelCreateCmd)
	channelCmd.AddCommand(channelUpdateCmd)
	channelCmd.AddCommand(channelEnableCmd)
	channelCmd.AddCommand(channelDeleteCmd)
	channelCmd.AddCommand(channelFetchModelCmd)
	channelCmd.AddCommand(channelSyncCmd)
	channelCmd.AddCommand(channelLastSyncCmd)
	RemoteCmd.AddCommand(channelCmd)
}
