// Tafcha CLI - Pipe text to get a URL
package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/rayenfassatoui/tafcha-cli/internal/cli"
)

var (
	// Flags
	apiURL  string
	expiry  string
	timeout time.Duration
	quiet   bool

	// Version info (set via ldflags)
	version = "dev"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "tafcha",
		Short: "Pipe text to get a shareable URL",
		Long: `Tafcha is a CLI tool for sharing text snippets.

Pipe any text to tafcha and get a short URL back.
The URL returns the exact plain text when accessed.

Examples:
  echo "hello world" | tafcha
  cat file.txt | tafcha --expiry 1d
  tafcha < script.sh --expiry 1w`,
		RunE:          run,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
	}

	// Flags
	rootCmd.Flags().StringVarP(&apiURL, "api", "a", "https://tafcha.dev", "API server URL")
	rootCmd.Flags().StringVarP(&expiry, "expiry", "e", "", "Expiry duration (e.g., 10m, 12h, 3d, 1w)")
	rootCmd.Flags().DurationVarP(&timeout, "timeout", "t", 30*time.Second, "Request timeout")
	rootCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Only output the URL (no extra info)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	// Check if stdin has data (is a pipe)
	stat, err := os.Stdin.Stat()
	if err != nil {
		return fmt.Errorf("checking stdin: %w", err)
	}

	if (stat.Mode() & os.ModeCharDevice) != 0 {
		// stdin is a terminal, not a pipe
		return fmt.Errorf("no input provided - pipe text to tafcha\n\nExample: echo \"hello\" | tafcha")
	}

	// Read all input from stdin
	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("reading stdin: %w", err)
	}

	if len(content) == 0 {
		return fmt.Errorf("empty input - nothing to upload")
	}

	// Create client and upload
	client := cli.NewClient(apiURL, timeout)
	resp, err := client.Create(content, expiry)
	if err != nil {
		return err
	}

	// Output result
	if quiet {
		fmt.Println(resp.URL)
	} else {
		fmt.Printf("%s\n", resp.URL)
		fmt.Fprintf(os.Stderr, "Expires: %s\n", resp.ExpiresAt.Local().Format("2006-01-02 15:04:05"))
	}

	return nil
}
