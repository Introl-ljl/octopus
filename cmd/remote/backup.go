package remote

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"strings"

	"github.com/bestruirui/octopus/internal/cli"
	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup and restore database",
	Long:  `Export database to a JSON backup file or import from a backup file.`,
}

var backupExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export database backup",
	RunE: func(cmd *cobra.Command, args []string) error {
		includeLogs, _ := cmd.Flags().GetBool("include-logs")
		includeStats, _ := cmd.Flags().GetBool("include-stats")

		session := cli.GetSession()
		if session == nil {
			return cli.ErrUnauthorized
		}

		profile := cli.GetActiveProfile()
		if profile == nil {
			return cli.NewCLIError(cli.ExitCodeInput, "no active profile")
		}

		query := fmt.Sprintf("include_logs=%v&include_stats=%v", includeLogs, includeStats)
		url := strings.TrimRight(profile.BaseURL, "/") + "/api/v1/setting/export?" + query

		req, err := httpReq("GET", url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+session.Token)

		resp, err := httpDo(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			return cli.NewCLIError(cli.ExitCodeServerError, fmt.Sprintf("export failed: %s", string(body)))
		}

		filename := "octopus-export.json"
		if cd := resp.Header.Get("Content-Disposition"); cd != "" {
			if parts := strings.Split(cd, "filename=\""); len(parts) > 1 {
				filename = strings.TrimRight(parts[1], "\"")
			}
		}

		outFile, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer outFile.Close()

		written, _ := io.Copy(outFile, resp.Body)
		fmt.Printf("Database exported to '%s' (%d bytes)\n", filename, written)
		return nil
	},
}

var backupImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import database backup",
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath, _ := cmd.Flags().GetString("file")

		if err := cli.RequireFlag(filePath, "file"); err != nil {
			return err
		}

		if !force && !confirmDestructive(fmt.Sprintf("Import from '%s'?", filePath)) {
			fmt.Println("Cancelled.")
			return nil
		}

		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		session := cli.GetSession()
		if session == nil {
			return cli.ErrUnauthorized
		}

		profile := cli.GetActiveProfile()
		if profile == nil {
			return cli.NewCLIError(cli.ExitCodeInput, "no active profile")
		}

		var buf strings.Builder
		w := multipart.NewWriter(&buf)
		part, _ := w.CreateFormFile("file", filePath)
		io.Copy(part, file)
		w.Close()

		url := strings.TrimRight(profile.BaseURL, "/") + "/api/v1/setting/import"
		req, err := httpReq("POST", url, strings.NewReader(buf.String()))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", w.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+session.Token)

		resp, err := httpDo(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			return cli.NewCLIError(cli.ExitCodeServerError, fmt.Sprintf("import failed: %s", string(body)))
		}

		fmt.Println("Database import completed.")
		return nil
	},
}

func init() {
	backupExportCmd.Flags().Bool("include-logs", false, "include relay logs in export")
	backupExportCmd.Flags().Bool("include-stats", false, "include statistics in export")
	backupImportCmd.Flags().String("file", "", "path to backup JSON file")
	backupCmd.AddCommand(backupExportCmd)
	backupCmd.AddCommand(backupImportCmd)
	RemoteCmd.AddCommand(backupCmd)
}
