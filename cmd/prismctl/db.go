package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/prism-control/pkg/service"
)

func dbCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Database management",
	}
	cmd.AddCommand(dbServeCmd())
	addDoltDBCommands(cmd)
	return cmd
}

func dbServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve [directory]",
		Short: "Start a Dolt SQL server",
		Long: `Start a Dolt SQL server bound to localhost. Press Ctrl-C to stop.

The server uses --data-dir (or PRISMCTL_DATA, or the default data directory)
as the multi-database root. Subdirectories containing .dolt are served as
databases. Other sessions connect via PRISMCTL_DSN.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var dir string
			if len(args) > 0 {
				dir = expandHome(args[0])
			} else {
				dir = getDataDir(cmd)
				if dir == "" {
					dir = expandHome(defaultDataDir)
				}
			}

			absDir, err := filepath.Abs(dir)
			if err != nil {
				return fmt.Errorf("resolve path: %w", err)
			}

			port, _ := cmd.Flags().GetInt("port")
			cmd.Printf("Starting Dolt SQL server on 127.0.0.1:%d (dir: %s)...\n", port, absDir)
			cmd.Printf("Connect with: export PRISMCTL_DSN=\"root:@tcp(127.0.0.1:%d)/prismcontrol\"\n", port)
			return service.DBServe(cmd.Context(), absDir, port)
		},
	}
	cmd.Flags().Int("port", 3306, "Port for the SQL server")
	return cmd
}
