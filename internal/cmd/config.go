package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/NaruseYuki/ramen-github-manager/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize default configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.ConfigPath()
		if err != nil {
			return err
		}

		defaultCfg := config.DefaultConfig()
		if err := config.Save(defaultCfg); err != nil {
			return err
		}

		fmt.Printf("✅ Configuration created at %s\n\n", path)
		fmt.Println("Default repositories:")
		for _, r := range defaultCfg.Repositories {
			fmt.Printf("  %-8s → %s\n", r.Alias, r.Name)
		}
		fmt.Printf("\nOwner: %s\n", defaultCfg.Owner)
		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := config.ConfigPath()
		fmt.Printf("📁 Config: %s\n\n", path)
		fmt.Printf("Owner: %s\n\n", cfg.Owner)
		fmt.Println("Repositories:")
		for _, r := range cfg.Repositories {
			alias := r.Alias
			if alias == "" {
				alias = "—"
			}
			fmt.Printf("  %-8s → %s\n", alias, r.Name)
		}
		fmt.Printf("\nDefaults:\n")
		fmt.Printf("  Sort:  %s\n", cfg.Defaults.Sort)
		fmt.Printf("  Limit: %d\n", cfg.Defaults.Limit)
		fmt.Printf("  State: %s\n", cfg.Defaults.State)
		return nil
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show configuration file path",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.ConfigPath()
		if err != nil {
			return err
		}
		fmt.Println(path)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configPathCmd)
}
