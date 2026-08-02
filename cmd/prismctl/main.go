// Package main is the entry point for the prismctl CLI.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/prism-build/pkg/config"
)

const defaultDSN = "root:@tcp(127.0.0.1:3306)/prismcontrol"
const defaultDataDir = "~/.productbuildershq/prism"

func main() {
	if err := rootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prismctl",
		Short: "PRISM Control — Product Delivery Control Plane",
		Long:  "Coordinate cross-repository initiatives, roadmap items, assignments, and delivery evidence.",
	}

	cmd.PersistentFlags().String("dsn", "", "MySQL-compatible DSN for Dolt server mode (default: $PRISMCTL_DSN or "+defaultDSN+")")
	cmd.PersistentFlags().String("data-dir", "", "Data directory for embedded Dolt (default: $PRISMCTL_DATA or "+defaultDataDir+")")

	cmd.AddCommand(
		versionCmd(),
		configCmd(),
		dbCmd(),
		registryCmd(),
		programCmd(),
		initiativeCmd(),
		phaseCmd(),
		rmiCmd(),
		workCmd(),
		contextCmd(),
		exportCmd(),
		ingestCmd(),
		reportCmd(),
		validateCmd(),
		mcpCmd(),
		dashboardCmd(),
		roadmapCmd(),
		releaseCmd(),
		workflowCmd(),
		specCmd(),
		maturityCmd(),
	)

	// DB-backed orchestration moved to the visionstudio CLI
	// (INIT-VISIONSTUDIO-001 Phase 4). The prismctl commands keep working
	// against the legacy prismcontrol database until the data migration,
	// but print a pointer to their successor. Cobra only prints the notice
	// on the executed leaf command, so it is propagated to all children.
	for _, c := range cmd.Commands() {
		switch c.Name() {
		case "db", "registry", "program", "initiative", "phase", "rmi", "work":
			markDeprecated(c, fmt.Sprintf("orchestration has moved to the visionstudio CLI; use `visionstudio %s` instead", c.Name()))
		}
	}
	return cmd
}

func markDeprecated(c *cobra.Command, msg string) {
	c.Deprecated = msg
	for _, child := range c.Commands() {
		markDeprecated(child, msg)
	}
}

func getDataDir(cmd *cobra.Command) string {
	dir, _ := cmd.Flags().GetString("data-dir")
	if dir != "" {
		return expandHome(dir)
	}
	if env := os.Getenv("PRISMCTL_DATA"); env != "" {
		return expandHome(env)
	}
	if os.Getenv("PRISMCTL_DSN") != "" {
		return ""
	}
	if cfg, err := config.Load(); err == nil && cfg.DSN != "" {
		return ""
	}
	return expandHome(defaultDataDir)
}

func expandHome(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("prismctl v0.1.0-dev")
		},
	}
}
