package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/prism-control/pkg/service"
	"github.com/ProductBuildersHQ/prism-control/pkg/store/doltstore"
)

//go:embed views.sql
var viewsSQL string

func dbCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Database management",
	}
	cmd.AddCommand(dbInitCmd(), dbServeCmd(), dbCreateViewsCmd())
	return cmd
}

func dbInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [directory]",
		Short: "Bootstrap a Dolt database for PRISM Control",
		Long: `Initialize a Dolt database and run schema migration.

In embedded mode (--data-dir or PRISMCTL_DATA set), creates the database
directory and runs migration automatically — no server needed.

In server mode, requires a running Dolt SQL server (see 'prismctl db serve')
and the --migrate flag.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dataDir := getDataDir(cmd)

			if dataDir != "" {
				return dbInitEmbedded(cmd, dataDir)
			}

			dir := defaultDataDir
			if len(args) > 0 {
				dir = args[0]
			}
			return dbInitServer(cmd, dir)
		},
	}
	cmd.Flags().Bool("migrate", false, "Also run schema migration (server mode only)")
	return cmd
}

func dbInitEmbedded(cmd *cobra.Command, dataDir string) error {
	absDir, err := filepath.Abs(dataDir)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	cmd.Printf("Initializing embedded Dolt database at %s...\n", absDir)
	ds, err := doltstore.NewEmbedded(dataDir)
	if err != nil {
		return fmt.Errorf("init embedded: %w", err)
	}
	defer func() {
		if err := ds.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: close database: %v\n", err)
		}
	}()

	cmd.Println("Running schema migration...")
	if err := ds.Migrate(cmd.Context()); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	cmd.Println("Embedded Dolt database initialized and migrated.")
	return nil
}

func dbInitServer(cmd *cobra.Command, dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	cmd.Printf("Initializing Dolt database at %s...\n", absDir)
	if err := service.DBInit(absDir); err != nil {
		return err
	}
	cmd.Println("Dolt database initialized.")

	migrateFlag, _ := cmd.Flags().GetBool("migrate")
	if !migrateFlag {
		cmd.Println("Run 'prismctl db serve' then 'prismctl db init --migrate' to create tables.")
		return nil
	}

	dsn := getDSN(cmd)
	cmd.Printf("Running schema migration (DSN: %s)...\n", dsn)
	ds, err := doltstore.New(dsn)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer func() {
		if err := ds.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: close database: %v\n", err)
		}
	}()

	if err := ds.Migrate(cmd.Context()); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	cmd.Println("Schema migration complete.")
	return nil
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

func dbCreateViewsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create-views",
		Short: "Create read-only SQL views for VisionStudio and other consumers",
		Long: `Execute the bundled SQL view definitions against the database.

Views created:
  v_initiative_summary  — initiative rollup (RMI counts, phase count, repos)
  v_phase_progress      — per-phase progress with derived status
  v_rmi_detail          — flat RMI detail with denormalized references
  v_active_assignments  — currently active work assignments

These views use only base tables (no Dolt system tables) and are safe
for read-only consumers such as VisionStudio.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			dataDir := getDataDir(cmd)

			var ds *doltstore.DoltStore
			var err error
			if dataDir != "" {
				ds, err = doltstore.NewEmbedded(dataDir)
			} else {
				ds, err = doltstore.New(getDSN(cmd))
			}
			if err != nil {
				return fmt.Errorf("connect to database: %w", err)
			}
			defer func() {
				if cerr := ds.Close(); cerr != nil {
					fmt.Fprintf(os.Stderr, "warning: close database: %v\n", cerr)
				}
			}()

			stmts := splitSQLStatements(viewsSQL)
			for _, stmt := range stmts {
				if err := ds.ExecSQL(cmd.Context(), stmt); err != nil {
					return fmt.Errorf("execute view statement: %w", err)
				}
			}

			cmd.Println("Created views: v_initiative_summary, v_phase_progress, v_rmi_detail, v_active_assignments")
			return nil
		},
	}
}

// splitSQLStatements splits a SQL script on semicolons, skipping
// empty statements and comments-only fragments.
func splitSQLStatements(sql string) []string {
	raw := strings.Split(sql, ";")
	var stmts []string
	for _, s := range raw {
		trimmed := strings.TrimSpace(s)
		// Skip empty or comment-only fragments
		if trimmed == "" {
			continue
		}
		lines := strings.Split(trimmed, "\n")
		hasSQL := false
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "--") {
				hasSQL = true
				break
			}
		}
		if hasSQL {
			stmts = append(stmts, trimmed)
		}
	}
	return stmts
}
