package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/NaruseYuki/ramen-github-manager/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration interactively",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.ConfigPath()
		if err != nil {
			return err
		}

		reader := bufio.NewReader(os.Stdin)
		newCfg := config.DefaultConfig()

		fmt.Println("📝 rgm configuration setup")
		fmt.Println()

		// Owner
		fmt.Print("GitHub owner (user or org): ")
		owner, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		owner = strings.TrimSpace(owner)
		if owner == "" {
			return fmt.Errorf("owner is required")
		}
		newCfg.Owner = owner

		// Repositories
		fmt.Println("\nAdd repositories (empty name to finish):")
		for i := 1; ; i++ {
			fmt.Printf("  Repo %d name: ", i)
			name, err := reader.ReadString('\n')
			if err != nil {
				break // EOF is acceptable here
			}
			name = strings.TrimSpace(name)
			if name == "" {
				break
			}

			fmt.Printf("  Repo %d alias (optional): ", i)
			alias, _ := reader.ReadString('\n')
			alias = strings.TrimSpace(alias)

			if err := config.AddRepo(newCfg, name, alias); err != nil {
				fmt.Printf("  ⚠️  %s (skipped)\n", err)
				continue
			}
			fmt.Println()
		}

		if err := config.Save(newCfg); err != nil {
			return err
		}

		fmt.Printf("\n✅ Configuration created at %s\n\n", path)
		fmt.Printf("Owner: %s\n", newCfg.Owner)
		if len(newCfg.Repositories) > 0 {
			fmt.Println("Repositories:")
			for _, r := range newCfg.Repositories {
				alias := r.Alias
				if alias == "" {
					alias = "—"
				}
				fmt.Printf("  %-12s → %s\n", alias, r.Name)
			}
		} else {
			fmt.Println("No repositories added. Use 'rgm config add-repo' to add them.")
		}
		return nil
	},
}

// loadCfg loads the config for config subcommands (since PersistentPreRunE is skipped).
func loadCfg() error {
	var err error
	cfg, err = config.Load()
	return err
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := loadCfg(); err != nil {
			return err
		}
		path, _ := config.ConfigPath()
		fmt.Printf("📁 Config: %s\n\n", path)
		fmt.Printf("Owner: %s\n\n", cfg.Owner)
		fmt.Println("Repositories:")
		if len(cfg.Repositories) == 0 {
			fmt.Println("  (none)")
		}
		for _, r := range cfg.Repositories {
			alias := r.Alias
			if alias == "" {
				alias = "—"
			}
			fmt.Printf("  %-12s → %s\n", alias, r.Name)
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

var configAddRepoCmd = &cobra.Command{
	Use:   "add-repo <name>",
	Short: "Add a repository to the configuration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := loadCfg(); err != nil {
			return err
		}
		alias, _ := cmd.Flags().GetString("alias")

		if err := config.AddRepo(cfg, args[0], alias); err != nil {
			return err
		}
		if err := config.Save(cfg); err != nil {
			return err
		}

		display := args[0]
		if alias != "" {
			display = fmt.Sprintf("%s (%s)", args[0], alias)
		}
		fmt.Printf("✅ Added repository: %s\n", display)
		return nil
	},
}

var configRemoveRepoCmd = &cobra.Command{
	Use:   "remove-repo <name-or-alias>",
	Short: "Remove a repository from the configuration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := loadCfg(); err != nil {
			return err
		}
		if err := config.RemoveRepo(cfg, args[0]); err != nil {
			return err
		}
		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Printf("✅ Removed repository: %s\n", args[0])
		return nil
	},
}

var configSetOwnerCmd = &cobra.Command{
	Use:   "set-owner <owner>",
	Short: "Change the GitHub owner (user or org)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := loadCfg(); err != nil {
			return err
		}
		cfg.Owner = args[0]
		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Printf("✅ Owner set to: %s\n", args[0])
		return nil
	},
}

func init() {
	configAddRepoCmd.Flags().String("alias", "", "Short alias for the repository")

	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configAddRepoCmd)
	configCmd.AddCommand(configRemoveRepoCmd)
	configCmd.AddCommand(configSetOwnerCmd)
}
