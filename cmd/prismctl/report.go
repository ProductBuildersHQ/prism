package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/prism-control/pkg/report"
)

func reportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate reports",
	}
	cmd.AddCommand(reportInitiativeCmd())
	return cmd
}

func reportInitiativeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "initiative <id>",
		Short: "Generate a full initiative report",
		Long: `Compute an end-to-end initiative report: duration, phases, RMI progress,
commit distribution by type and repo, releases, and unattributed residual.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			r, err := report.Generate(cmd.Context(), svc.Store, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			switch format {
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(r)
			case "markdown":
				fmt.Print(r.Markdown())
				return nil
			default:
				return fmt.Errorf("unknown format: %s (use json or markdown)", format)
			}
		},
	}
	cmd.Flags().String("format", "json", "Output format: json or markdown")
	return cmd
}
