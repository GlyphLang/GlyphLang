package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/glyphlang/glyph/pkg/ir"
	"github.com/spf13/cobra"
)

// runIR handles the ir command — exports ServiceIR or ServiceCatalog as JSON.
func runIR(cmd *cobra.Command, args []string) error {
	filePath := args[0]
	format, _ := cmd.Flags().GetString("format")
	output, _ := cmd.Flags().GetString("output")
	pretty, _ := cmd.Flags().GetBool("pretty")

	source, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	module, err := parseSource(string(source))
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	analyzer := ir.NewAnalyzer()
	service, err := analyzer.Analyze(module)
	if err != nil {
		return fmt.Errorf("IR analysis error: %w", err)
	}

	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	service.Name = base[:len(base)-len(ext)]

	var outputData []byte

	switch format {
	case "json":
		if pretty {
			outputData, err = json.MarshalIndent(service, "", "  ")
		} else {
			outputData, err = json.Marshal(service)
		}
		if err != nil {
			return fmt.Errorf("failed to serialize IR: %w", err)
		}

	case "catalog":
		catalog := ir.ToCatalog(service)
		if pretty {
			outputData, err = json.MarshalIndent(catalog, "", "  ")
		} else {
			outputData, err = json.Marshal(catalog)
		}
		if err != nil {
			return fmt.Errorf("failed to serialize catalog: %w", err)
		}

	case "compact":
		catalog := ir.ToCatalog(service)
		outputData = []byte(catalogToCompact(catalog))

	default:
		return fmt.Errorf("unknown format: %s (use: json, catalog, compact)", format)
	}

	if output != "" {
		if err := os.WriteFile(output, outputData, 0644); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		opCount := len(service.Routes) + len(service.Functions) + len(service.Commands) +
			len(service.Events) + len(service.CronJobs) + len(service.Queues)
		printSuccess(fmt.Sprintf("IR written to %s", output))
		printInfo(fmt.Sprintf("Service: %s | Operations: %d | Types: %d",
			service.Name, opCount, len(service.Types)))
		return nil
	}

	fmt.Println(string(outputData))
	return nil
}

// catalogToCompact renders a ServiceCatalog as a human-readable summary
// using GlyphLang @-notation style consistent with pkg/context/context.go.
func catalogToCompact(catalog *ir.ServiceCatalog) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s v%s\n", catalog.Service, catalog.Version))

	if len(catalog.Types) > 0 {
		sb.WriteString("\n## Types\n")
		for _, t := range catalog.Types {
			sb.WriteString(fmt.Sprintf(": %s {\n", t.Name))
			for _, f := range t.Fields {
				req := ""
				if f.Required {
					req = "!"
				}
				sb.WriteString(fmt.Sprintf("  %s: %s%s\n", f.Name, f.Type, req))
			}
			sb.WriteString("}\n")
		}
	}

	if len(catalog.Providers) > 0 {
		sb.WriteString("\n## Providers\n")
		for _, p := range catalog.Providers {
			sb.WriteString(fmt.Sprintf("%% %s: %s\n", p.Name, p.ProviderType))
		}
	}

	routes := filterOps(catalog.Operations, "route")
	if len(routes) > 0 {
		sb.WriteString("\n## Routes\n")
		for _, op := range routes {
			auth := ""
			if op.Auth != nil {
				auth = fmt.Sprintf(" [+auth(%s)]", op.Auth.Type)
			}
			ret := ""
			if op.Returns != "" {
				ret = " -> " + op.Returns
			}
			sb.WriteString(fmt.Sprintf("%s%s%s\n", op.Notation, ret, auth))
		}
	}

	functions := filterOps(catalog.Operations, "function")
	if len(functions) > 0 {
		sb.WriteString("\n## Functions\n")
		for _, op := range functions {
			sb.WriteString(op.Notation + "\n")
		}
	}

	commands := filterOps(catalog.Operations, "command")
	if len(commands) > 0 {
		sb.WriteString("\n## Commands\n")
		for _, op := range commands {
			sb.WriteString(op.Notation + "\n")
		}
	}

	events := filterOps(catalog.Operations, "event")
	if len(events) > 0 {
		sb.WriteString("\n## Events\n")
		for _, op := range events {
			sb.WriteString(op.Notation + "\n")
		}
	}

	crons := filterOps(catalog.Operations, "cron")
	if len(crons) > 0 {
		sb.WriteString("\n## Cron Jobs\n")
		for _, op := range crons {
			sb.WriteString(op.Notation + "\n")
		}
	}

	queues := filterOps(catalog.Operations, "queue")
	if len(queues) > 0 {
		sb.WriteString("\n## Queues\n")
		for _, op := range queues {
			sb.WriteString(op.Notation + "\n")
		}
	}

	grpc := filterOps(catalog.Operations, "grpc")
	if len(grpc) > 0 {
		sb.WriteString("\n## gRPC\n")
		for _, op := range grpc {
			sb.WriteString(op.Notation + "\n")
		}
	}

	graphql := filterOps(catalog.Operations, "graphql")
	if len(graphql) > 0 {
		sb.WriteString("\n## GraphQL\n")
		for _, op := range graphql {
			sb.WriteString(op.Notation + "\n")
		}
	}

	return sb.String()
}

func filterOps(ops []ir.CatalogOp, kind string) []ir.CatalogOp {
	var result []ir.CatalogOp
	for _, op := range ops {
		if op.Kind == kind {
			result = append(result, op)
		}
	}
	return result
}
