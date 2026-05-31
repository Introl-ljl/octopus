package remote

import (
	"fmt"
	"strconv"

	"github.com/bestruirui/octopus/internal/cli"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Manage relay logs",
	Long:  `View and clear relay logs.`,
}

var logListCmd = &cobra.Command{
	Use:   "list",
	Short: "List relay logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 {
			pageSize = 20
		}

		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			path := fmt.Sprintf("/api/v1/log/list?page=%d&page_size=%d", page, pageSize)
			var logs []map[string]interface{}
			if err := c.Get(path, &logs); err != nil {
				return err
			}
			if len(logs) == 0 {
				fmt.Println("No logs found.")
				return nil
			}
			f.PrintHeader("ID", "Model", "Channel", "Tokens", "Time(ms)", "Cost")
			for _, l := range logs {
				tokens := fmt.Sprintf("%v/%v", l["input_tokens"], l["output_tokens"])
				useTime := fmt.Sprintf("%v", l["use_time"])
				if t, ok := l["use_time"].(float64); ok {
					useTime = strconv.Itoa(int(t))
				}
				f.PrintRow(
					fmt.Sprintf("%v", l["id"]),
					fmt.Sprintf("%v", l["request_model_name"]),
					fmt.Sprintf("%v", l["channel_name"]),
					tokens,
					useTime,
					fmt.Sprintf("%v", l["cost"]),
				)
			}
			f.Flush()
			return nil
		})
	},
}

var logClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all relay logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !force && !confirmDestructive("Clear all relay logs?") {
			fmt.Println("Cancelled.")
			return nil
		}

		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			return c.Delete("/api/v1/log/clear", nil)
		})
	},
}

func init() {
	logListCmd.Flags().Int("page", 1, "page number")
	logListCmd.Flags().Int("page-size", 20, "items per page")
	logCmd.AddCommand(logListCmd)
	logCmd.AddCommand(logClearCmd)
	RemoteCmd.AddCommand(logCmd)
}
