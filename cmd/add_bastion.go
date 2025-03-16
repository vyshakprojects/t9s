package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"mytunnel/internal/config"
)

var (
	host     string
	user     string
	port     int
	authType string
	keyPath  string
	password string
)

// addBastionCmd represents the add-bastion command
var addBastionCmd = &cobra.Command{
	Use:   "add-bastion",
	Short: "Add a new bastion server configuration",
	Long: `Add a new bastion server configuration to your MyTunnel config file.
You can specify either SSH key authentication or password authentication.

Example:
  mytunnel add-bastion --name my-bastion --host bastion.example.com --user admin --auth-type key --key-path ~/.ssh/id_rsa
  mytunnel add-bastion --name my-bastion --host bastion.example.com --user admin --auth-type password --password mypass`,
	RunE: runAddBastion,
}

func init() {
	rootCmd.AddCommand(addBastionCmd)

	addBastionCmd.Flags().StringVar(&bastionName, "name", "", "name of the bastion server")
	addBastionCmd.Flags().StringVar(&host, "host", "", "hostname of the bastion server")
	addBastionCmd.Flags().StringVar(&user, "user", "", "username for SSH connection")
	addBastionCmd.Flags().IntVar(&port, "port", 22, "SSH port number")
	addBastionCmd.Flags().StringVar(&authType, "auth-type", "key", "authentication type (key or password)")
	addBastionCmd.Flags().StringVar(&keyPath, "key-path", "", "path to SSH private key")
	addBastionCmd.Flags().StringVar(&password, "password", "", "SSH password (if using password auth)")

	addBastionCmd.MarkFlagRequired("name")
	addBastionCmd.MarkFlagRequired("host")
	addBastionCmd.MarkFlagRequired("user")
}

func runAddBastion(cmd *cobra.Command, args []string) error {
	// Load existing config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate auth type
	if authType != "key" && authType != "password" {
		return fmt.Errorf("invalid auth-type: must be either 'key' or 'password'")
	}

	// Validate auth credentials
	if authType == "key" && keyPath == "" {
		return fmt.Errorf("key-path is required when using key authentication")
	}
	if authType == "password" && password == "" {
		return fmt.Errorf("password is required when using password authentication")
	}

	// Create new bastion config
	bastion := &config.BastionConfig{
		Host:     host,
		User:     user,
		Port:     port,
		AuthType: authType,
		KeyPath:  keyPath,
		Password: password,
	}

	// Add to config
	cfg.AddBastion(bastionName, bastion)

	// Save config
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Successfully added bastion server '%s'\n", bastionName)
	return nil
} 