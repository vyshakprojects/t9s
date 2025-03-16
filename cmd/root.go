package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"mytunnel/internal/config"
	"mytunnel/internal/ssh"
	"mytunnel/internal/ui"
)

var (
	cfgFile     string
	bastionName string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mytunnel",
	Short: "A k9s-like SSH tunnel manager",
	Long: `MyTunnel is an interactive SSH tunnel manager that helps you create and manage
SSH tunnels through bastion servers with a k9s-like interface.

Features:
- Interactive terminal UI with Vim-style navigation
- Bastion server configuration management
- Real-time port monitoring and tunnel management
- SSH key and password authentication support`,
	RunE: runRoot,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mytunnel/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&bastionName, "bastion", "", "bastion server to connect to")
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	if cfgFile == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		cfgFile = filepath.Join(home, ".mytunnel", "config.yaml")
	}
}

func runRoot(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// If no bastion is specified and there's only one, use it
	if bastionName == "" {
		if len(cfg.Bastions) == 0 {
			return fmt.Errorf("no bastions configured. Use 'mytunnel add-bastion' to add one")
		}
		if len(cfg.Bastions) == 1 {
			for name := range cfg.Bastions {
				bastionName = name
				break
			}
		} else {
			return fmt.Errorf("multiple bastions configured, please specify one with --bastion")
		}
	}

	// Get the specified bastion
	bastion, ok := cfg.Bastions[bastionName]
	if !ok {
		return fmt.Errorf("bastion '%s' not found", bastionName)
	}

	// Create tunnel manager
	tunnelManager := ssh.NewTunnelManager()

	// Create and run UI
	ui := ui.NewUI(tunnelManager, bastion)
	
	// Set some default ports for testing (you would replace this with actual port discovery)
	ui.SetPorts([]int{22, 80, 443, 3306, 5432, 6379, 8080, 8443})

	return ui.Run()
} 