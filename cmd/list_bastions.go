package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"mytunnel/internal/config"
)

// listBastionsCmd represents the list-bastions command
var listBastionsCmd = &cobra.Command{
	Use:   "list-bastions",
	Short: "List all configured bastion servers",
	Long: `List all bastion servers that have been configured in MyTunnel.
Displays the name, host, user, port, and authentication type for each bastion.`,
	RunE: runListBastions,
}

func init() {
	rootCmd.AddCommand(listBastionsCmd)
}

func runListBastions(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.Bastions) == 0 {
		fmt.Println("No bastion servers configured")
		return nil
	}

	// Create tabwriter for formatted output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tHOST\tUSER\tPORT\tAUTH TYPE")
	fmt.Fprintln(w, "----\t----\t----\t----\t---------")

	for name, bastion := range cfg.Bastions {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n",
			name,
			bastion.Host,
			bastion.User,
			bastion.Port,
			bastion.AuthType)
	}

	return w.Flush()
} 