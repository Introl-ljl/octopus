package remote

import (
	"fmt"

	"github.com/bestruirui/octopus/internal/cli"
	"github.com/spf13/cobra"
)

var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "Manage groups",
	Long:  `Manage model groups: list, create, update, and delete.`,
}

var groupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all groups",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			var groups []map[string]interface{}
			if err := c.Get("/api/v1/group/list", &groups); err != nil {
				return err
			}
			f.PrintHeader("ID", "Name", "Mode", "Match Regex")
			for _, g := range groups {
				f.PrintRow(
					fmt.Sprintf("%v", g["id"]),
					fmt.Sprintf("%v", g["name"]),
					fmt.Sprintf("%v", g["mode"]),
					fmt.Sprintf("%v", g["match_regex"]),
				)
			}
			f.Flush()
			return nil
		})
	},
}

var groupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new group",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		mode, _ := cmd.Flags().GetInt("mode")

		if err := cli.RequireFlag(name, "name"); err != nil {
			return err
		}

		body := map[string]interface{}{
			"name": name,
			"mode": mode,
		}

		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			var result map[string]interface{}
			if err := c.Post("/api/v1/group/create", body, &result); err != nil {
				return err
			}
			f.PrintSuccess(fmt.Sprintf("Group '%s' created (ID: %v)", name, result["id"]))
			return nil
		})
	},
}

var groupDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a group",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetInt("id")
		if id == 0 {
			return cli.NewCLIError(cli.ExitCodeInput, "--id is required")
		}

		if !force && !confirmDestructive(fmt.Sprintf("Delete group %d?", id)) {
			fmt.Println("Cancelled.")
			return nil
		}

		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			return c.Delete(fmt.Sprintf("/api/v1/group/delete/%d", id), nil)
		})
	},
}

var groupUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a group",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetInt("id")
		if id == 0 {
			return cli.NewCLIError(cli.ExitCodeInput, "--id is required")
		}
		body := map[string]interface{}{"id": id}
		if cmd.Flags().Changed("name") {
			body["name"], _ = cmd.Flags().GetString("name")
		}
		if cmd.Flags().Changed("mode") {
			body["mode"], _ = cmd.Flags().GetInt("mode")
		}
		return runWithClient(func(c *cli.Client, f cli.Formatter) error {
			var result map[string]interface{}
			if err := c.Post("/api/v1/group/update", body, &result); err != nil {
				return err
			}
			f.PrintSuccess(fmt.Sprintf("Group %d updated.", id))
			return nil
		})
	},
}

func init() {
	groupCreateCmd.Flags().String("name", "", "group name")
	groupCreateCmd.Flags().Int("mode", 1, "routing mode (1=RoundRobin, 2=Random, 3=Failover, 4=Weighted)")
	groupDeleteCmd.Flags().Int("id", 0, "group ID to delete")
	groupUpdateCmd.Flags().Int("id", 0, "group ID")
	groupUpdateCmd.Flags().String("name", "", "group name")
	groupUpdateCmd.Flags().Int("mode", 0, "routing mode")
	groupCmd.AddCommand(groupListCmd)
	groupCmd.AddCommand(groupCreateCmd)
	groupCmd.AddCommand(groupUpdateCmd)
	groupCmd.AddCommand(groupDeleteCmd)
	RemoteCmd.AddCommand(groupCmd)
}
