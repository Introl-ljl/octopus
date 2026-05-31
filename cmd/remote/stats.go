package remote

import (
	"fmt"

	"github.com/bestruirui/octopus/internal/cli"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "View statistics",
	Long:  `View usage statistics: today, daily, hourly, total, per API key, and per model.`,
}

var statsTodayCmd = &cobra.Command{
	Use:   "today",
	Short: "Show today's stats",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			var result map[string]interface{}
			if err := c.Get("/api/v1/stats/today", &result); err != nil {
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

var statsDailyCmd = &cobra.Command{
	Use:   "daily",
	Short: "Show daily stats",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			var results []map[string]interface{}
			if err := c.Get("/api/v1/stats/daily", &results); err != nil {
				return err
			}
			f.PrintHeader("Date", "Input Token", "Output Token", "Cost", "Success", "Failed")
			for _, r := range results {
				f.PrintRow(
					fmt.Sprintf("%v", r["date"]),
					fmt.Sprintf("%v", r["input_token"]),
					fmt.Sprintf("%v", r["output_token"]),
					fmt.Sprintf("%v", r["output_cost"]),
					fmt.Sprintf("%v", r["request_success"]),
					fmt.Sprintf("%v", r["request_failed"]),
				)
			}
			f.Flush()
			return nil
		})
	},
}

var statsHourlyCmd = &cobra.Command{
	Use:   "hourly",
	Short: "Show hourly stats",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			var results []map[string]interface{}
			if err := c.Get("/api/v1/stats/hourly", &results); err != nil {
				return err
			}
			f.PrintHeader("Date", "Hour", "Input Token", "Output Token", "Cost", "Success", "Failed")
			for _, r := range results {
				f.PrintRow(
					fmt.Sprintf("%v", r["date"]),
					fmt.Sprintf("%v", r["hour"]),
					fmt.Sprintf("%v", r["input_token"]),
					fmt.Sprintf("%v", r["output_token"]),
					fmt.Sprintf("%v", r["output_cost"]),
					fmt.Sprintf("%v", r["request_success"]),
					fmt.Sprintf("%v", r["request_failed"]),
				)
			}
			f.Flush()
			return nil
		})
	},
}

var statsTotalCmd = &cobra.Command{
	Use:   "total",
	Short: "Show total stats",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			var result map[string]interface{}
			if err := c.Get("/api/v1/stats/total", &result); err != nil {
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

var statsApikeyCmd = &cobra.Command{
	Use:   "apikey",
	Short: "Show per-API-key stats",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			var results []map[string]interface{}
			if err := c.Get("/api/v1/stats/apikey", &results); err != nil {
				return err
			}
			f.PrintHeader("API Key ID", "Input Token", "Output Token", "Cost", "Success", "Failed")
			for _, r := range results {
				f.PrintRow(
					fmt.Sprintf("%v", r["api_key_id"]),
					fmt.Sprintf("%v", r["input_token"]),
					fmt.Sprintf("%v", r["output_token"]),
					fmt.Sprintf("%v", r["output_cost"]),
					fmt.Sprintf("%v", r["request_success"]),
					fmt.Sprintf("%v", r["request_failed"]),
				)
			}
			f.Flush()
			return nil
		})
	},
}

var statsModelCmd = &cobra.Command{
	Use:   "model",
	Short: "Show per-model stats",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			var results []map[string]interface{}
			if err := c.Get("/api/v1/stats/model", &results); err != nil {
				return err
			}
			f.PrintHeader("Model", "Channel ID", "Input Token", "Output Token", "Cost", "Success", "Failed")
			for _, r := range results {
				f.PrintRow(
					fmt.Sprintf("%v", r["name"]),
					fmt.Sprintf("%v", r["channel_id"]),
					fmt.Sprintf("%v", r["input_token"]),
					fmt.Sprintf("%v", r["output_token"]),
					fmt.Sprintf("%v", r["output_cost"]),
					fmt.Sprintf("%v", r["request_success"]),
					fmt.Sprintf("%v", r["request_failed"]),
				)
			}
			f.Flush()
			return nil
		})
	},
}

func init() {
	statsCmd.AddCommand(statsTodayCmd)
	statsCmd.AddCommand(statsDailyCmd)
	statsCmd.AddCommand(statsHourlyCmd)
	statsCmd.AddCommand(statsTotalCmd)
	statsCmd.AddCommand(statsApikeyCmd)
	statsCmd.AddCommand(statsModelCmd)
	RemoteCmd.AddCommand(statsCmd)
}
