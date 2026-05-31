package remote

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/bestruirui/octopus/internal/cli"
)

func getClientAndFormatter() (*cli.Client, cli.Formatter, error) {
	session := cli.GetSession()
	if session == nil {
		return nil, nil, cli.ErrUnauthorized
	}

	profile := cli.GetActiveProfile()
	if profile == nil {
		return nil, nil, cli.NewCLIError(cli.ExitCodeInput, "no active profile. Configure one with 'octopus remote profile add'")
	}

	baseURL := profile.BaseURL
	if serverURL != "" {
		baseURL = serverURL
	}
	client := cli.NewClient(baseURL)
	client.SetToken(session.Token)

	var f cli.Formatter
	if outputMode == "json" {
		f = cli.NewJSONFormatter()
	} else if profile.DefaultOutput == "json" {
		f = cli.NewJSONFormatter()
	} else {
		f = cli.NewTableFormatter()
	}

	return client, f, nil
}

func runWithClient(fn func(*cli.Client, cli.Formatter) error) error {
	client, formatter, err := getClientAndFormatter()
	if err != nil {
		return err
	}
	return fn(client, formatter)
}

func confirmDestructive(msg string) bool {
	fmt.Fprintf(os.Stderr, "WARNING: %s\n", msg)
	fmt.Fprint(os.Stderr, "Are you sure? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	return response == "y" || response == "yes"
}

func httpReq(method, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, url, body)
}

func httpDo(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	return client.Do(req)
}

func init() {
	var _ = fmt.Println
}

var confirmDestructiveVar = confirmDestructive
