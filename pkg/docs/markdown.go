package docs

import (
	"fmt"
	"strings"
)

// GenerateMarkdown produces Markdown documentation from an APIDoc.
func GenerateMarkdown(doc *APIDoc) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", doc.Title))

	// Table of contents
	if len(doc.Routes) > 0 || len(doc.Types) > 0 {
		sb.WriteString("## Table of Contents\n\n")
		if len(doc.Routes) > 0 {
			sb.WriteString("- [Endpoints](#endpoints)\n")
			for _, r := range doc.Routes {
				anchor := routeAnchor(r)
				sb.WriteString(fmt.Sprintf("  - [%s %s](#%s)\n", r.Method, r.Path, anchor))
			}
		}
		if len(doc.Types) > 0 {
			sb.WriteString("- [Types](#types)\n")
			for _, t := range doc.Types {
				sb.WriteString(fmt.Sprintf("  - [%s](#%s)\n", t.Name, strings.ToLower(t.Name)))
			}
		}
		sb.WriteString("\n")
	}

	// Endpoints
	if len(doc.Routes) > 0 {
		sb.WriteString("## Endpoints\n\n")
		for _, r := range doc.Routes {
			writeRouteMarkdown(&sb, r)
		}
	}

	// Types
	if len(doc.Types) > 0 {
		sb.WriteString("## Types\n\n")
		for _, t := range doc.Types {
			writeTypeMarkdown(&sb, t)
		}
	}

	return sb.String()
}

func writeRouteMarkdown(sb *strings.Builder, r RouteDoc) {
	sb.WriteString(fmt.Sprintf("### %s `%s`\n\n", r.Method, r.Path))

	if r.HasAuth {
		sb.WriteString(fmt.Sprintf("**Authentication:** %s\n\n", r.AuthType))
	}
	if r.RateLimit != "" {
		sb.WriteString(fmt.Sprintf("**Rate Limit:** %s\n\n", r.RateLimit))
	}

	// Path parameters
	if len(r.PathParams) > 0 {
		sb.WriteString("**Path Parameters:**\n\n")
		sb.WriteString("| Parameter | Type |\n")
		sb.WriteString("|-----------|------|\n")
		for _, p := range r.PathParams {
			sb.WriteString(fmt.Sprintf("| `%s` | string |\n", p))
		}
		sb.WriteString("\n")
	}

	// Query parameters
	if len(r.QueryParams) > 0 {
		sb.WriteString("**Query Parameters:**\n\n")
		sb.WriteString("| Parameter | Type | Required |\n")
		sb.WriteString("|-----------|------|----------|\n")
		for _, qp := range r.QueryParams {
			req := "No"
			if qp.Required {
				req = "Yes"
			}
			sb.WriteString(fmt.Sprintf("| `%s` | %s | %s |\n", qp.Name, qp.Type, req))
		}
		sb.WriteString("\n")
	}

	// Request body
	if r.InputType != "" {
		sb.WriteString(fmt.Sprintf("**Request Body:** `%s`\n\n", r.InputType))
	}

	// Response
	if r.ReturnType != "" {
		sb.WriteString(fmt.Sprintf("**Response:** `%s`\n\n", r.ReturnType))
	}

	sb.WriteString("---\n\n")
}

func writeTypeMarkdown(sb *strings.Builder, t TypeDoc) {
	sb.WriteString(fmt.Sprintf("### %s\n\n", t.Name))

	if len(t.Fields) > 0 {
		sb.WriteString("| Field | Type | Required |\n")
		sb.WriteString("|-------|------|----------|\n")
		for _, f := range t.Fields {
			req := "No"
			if f.Required {
				req = "Yes"
			}
			sb.WriteString(fmt.Sprintf("| `%s` | %s | %s |\n", f.Name, f.Type, req))
		}
		sb.WriteString("\n")
	}
}

func routeAnchor(r RouteDoc) string {
	s := strings.ToLower(r.Method) + "-" + r.Path
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, " ", "-")
	return s
}
