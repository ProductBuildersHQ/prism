package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func programCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "program",
		Aliases: []string{"prog"},
		Short:   "Manage programs",
	}
	cmd.AddCommand(
		programCreateCmd(),
		programListCmd(),
		programGetCmd(),
		programUpdateCmd(),
		programMigrateCmd(),
	)
	return cmd
}

func programCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new program",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			id, _ := cmd.Flags().GetString("id")
			name, _ := cmd.Flags().GetString("name")
			org, _ := cmd.Flags().GetString("org")
			desc, _ := cmd.Flags().GetString("description")

			if id == "" || name == "" {
				return fmt.Errorf("--id and --name are required")
			}
			if org == "" {
				org = "default"
			}

			prog, err := svc.CreateProgram(cmd.Context(), id, name, org, desc)
			if err != nil {
				return err
			}
			cmd.Printf("Created program: %s (%s)\n", prog.ID, prog.Name)
			return nil
		},
	}
	cmd.Flags().String("id", "", "Program ID (e.g. PROG-DELIVERY) (required)")
	cmd.Flags().String("name", "", "Program display name (required)")
	cmd.Flags().String("org", "", "Organization (default: 'default')")
	cmd.Flags().String("description", "", "Description")
	return cmd
}

func programListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all programs",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			progs, err := svc.ListPrograms(cmd.Context())
			if err != nil {
				return err
			}
			if len(progs) == 0 {
				cmd.Println("No programs found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "ID\tNAME\tORG\tDESCRIPTION")
			for _, p := range progs {
				desc := p.Description
				if desc == "" {
					desc = "-"
				}
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.ID, p.Name, p.Organization, desc)
			}
			return w.Flush()
		},
	}
}

func programGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <program-id>",
		Short: "Show program details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			prog, err := svc.GetProgram(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			cmd.Printf("Program:      %s\n", prog.ID)
			cmd.Printf("Name:         %s\n", prog.Name)
			cmd.Printf("Organization: %s\n", prog.Organization)
			if prog.Description != "" {
				cmd.Printf("Description:  %s\n", prog.Description)
			}
			cmd.Printf("Created:      %s\n", prog.CreatedAt.Format("2006-01-02 15:04"))
			return nil
		},
	}
}

func programUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <program-id>",
		Short: "Update program fields",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			prog, err := svc.Store.GetProgram(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			if cmd.Flags().Changed("name") {
				prog.Name, _ = cmd.Flags().GetString("name")
			}
			if cmd.Flags().Changed("description") {
				prog.Description, _ = cmd.Flags().GetString("description")
			}
			if err := svc.UpdateProgram(cmd.Context(), prog); err != nil {
				return err
			}
			cmd.Printf("Updated %s\n", prog.ID)
			return nil
		},
	}
	cmd.Flags().String("name", "", "Program display name")
	cmd.Flags().String("description", "", "Description")
	return cmd
}

func programMigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate-strings",
		Short: "Convert free-text program strings on initiatives to Program entities",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			inits, err := svc.ListInitiatives(cmd.Context())
			if err != nil {
				return err
			}

			seen := map[string]bool{}
			for _, init := range inits {
				if init.ProgramID != "" && !strings.HasPrefix(init.ProgramID, "PROG-") {
					seen[init.ProgramID] = true
				}
			}

			if len(seen) == 0 {
				cmd.Println("No free-text program values to migrate.")
				return nil
			}

			nameToID := map[string]string{}
			for name := range seen {
				slug := strings.ToUpper(strings.ReplaceAll(
					strings.ReplaceAll(name, " ", "-"), "_", "-"))
				id := "PROG-" + slug
				_, err := svc.CreateProgram(cmd.Context(), id, name, "default", "")
				if err != nil {
					cmd.PrintErrf("Warning: could not create program %s: %v\n", id, err)
					continue
				}
				nameToID[name] = id
				cmd.Printf("Created program: %s -> %s\n", name, id)
			}

			for _, init := range inits {
				if newID, ok := nameToID[init.ProgramID]; ok {
					init.ProgramID = newID
					if err := svc.UpdateInitiative(cmd.Context(), init); err != nil {
						cmd.PrintErrf("Warning: could not update initiative %s: %v\n", init.ID, err)
					} else {
						cmd.Printf("Updated initiative %s -> program %s\n", init.ID, newID)
					}
				}
			}

			return nil
		},
	}
}
