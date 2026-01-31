package docs

import (
	"fmt"
	"html"
	"strings"
)

// GenerateHTML produces a self-contained HTML documentation page from an APIDoc.
func GenerateHTML(doc *APIDoc) string {
	var sb strings.Builder

	sb.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n")
	sb.WriteString("<meta charset=\"UTF-8\">\n")
	sb.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	sb.WriteString(fmt.Sprintf("<title>%s</title>\n", html.EscapeString(doc.Title)))
	writeCSS(&sb)
	sb.WriteString("</head>\n<body>\n")

	// Sidebar
	sb.WriteString("<nav class=\"sidebar\">\n")
	sb.WriteString(fmt.Sprintf("<h2>%s</h2>\n", html.EscapeString(doc.Title)))
	sb.WriteString("<input type=\"text\" id=\"search\" placeholder=\"Search...\" onkeyup=\"filterDocs()\">\n")

	if len(doc.Routes) > 0 {
		sb.WriteString("<h3>Endpoints</h3>\n<ul>\n")
		for _, r := range doc.Routes {
			anchor := routeAnchor(r)
			sb.WriteString(fmt.Sprintf("<li><a href=\"#%s\"><span class=\"method method-%s\">%s</span> %s</a></li>\n",
				anchor, strings.ToLower(r.Method), r.Method, html.EscapeString(r.Path)))
		}
		sb.WriteString("</ul>\n")
	}

	if len(doc.Types) > 0 {
		sb.WriteString("<h3>Types</h3>\n<ul>\n")
		for _, t := range doc.Types {
			sb.WriteString(fmt.Sprintf("<li><a href=\"#%s\">%s</a></li>\n",
				strings.ToLower(t.Name), html.EscapeString(t.Name)))
		}
		sb.WriteString("</ul>\n")
	}
	sb.WriteString("</nav>\n")

	// Main content
	sb.WriteString("<main class=\"content\">\n")
	sb.WriteString(fmt.Sprintf("<h1>%s</h1>\n", html.EscapeString(doc.Title)))

	if len(doc.Routes) > 0 {
		sb.WriteString("<h2 id=\"endpoints\">Endpoints</h2>\n")
		for _, r := range doc.Routes {
			writeRouteHTML(&sb, r)
		}
	}

	if len(doc.Types) > 0 {
		sb.WriteString("<h2 id=\"types\">Types</h2>\n")
		for _, t := range doc.Types {
			writeTypeHTML(&sb, t)
		}
	}

	sb.WriteString("</main>\n")
	writeScript(&sb)
	sb.WriteString("</body>\n</html>\n")

	return sb.String()
}

func writeRouteHTML(sb *strings.Builder, r RouteDoc) {
	anchor := routeAnchor(r)
	sb.WriteString(fmt.Sprintf("<section class=\"endpoint\" id=\"%s\">\n", anchor))
	sb.WriteString(fmt.Sprintf("<h3><span class=\"method method-%s\">%s</span> <code>%s</code></h3>\n",
		strings.ToLower(r.Method), r.Method, html.EscapeString(r.Path)))

	if r.HasAuth {
		sb.WriteString(fmt.Sprintf("<p class=\"badge auth\">Auth: %s</p>\n", html.EscapeString(r.AuthType)))
	}
	if r.RateLimit != "" {
		sb.WriteString(fmt.Sprintf("<p class=\"badge rate-limit\">Rate Limit: %s</p>\n", html.EscapeString(r.RateLimit)))
	}

	if len(r.PathParams) > 0 {
		sb.WriteString("<h4>Path Parameters</h4>\n<table>\n<tr><th>Parameter</th><th>Type</th></tr>\n")
		for _, p := range r.PathParams {
			sb.WriteString(fmt.Sprintf("<tr><td><code>%s</code></td><td>string</td></tr>\n", html.EscapeString(p)))
		}
		sb.WriteString("</table>\n")
	}

	if len(r.QueryParams) > 0 {
		sb.WriteString("<h4>Query Parameters</h4>\n<table>\n<tr><th>Parameter</th><th>Type</th><th>Required</th></tr>\n")
		for _, qp := range r.QueryParams {
			req := "No"
			if qp.Required {
				req = "Yes"
			}
			sb.WriteString(fmt.Sprintf("<tr><td><code>%s</code></td><td>%s</td><td>%s</td></tr>\n",
				html.EscapeString(qp.Name), html.EscapeString(qp.Type), req))
		}
		sb.WriteString("</table>\n")
	}

	if r.InputType != "" {
		sb.WriteString(fmt.Sprintf("<h4>Request Body</h4>\n<p><code>%s</code></p>\n", html.EscapeString(r.InputType)))
	}

	if r.ReturnType != "" {
		sb.WriteString(fmt.Sprintf("<h4>Response</h4>\n<p><code>%s</code></p>\n", html.EscapeString(r.ReturnType)))
	}

	sb.WriteString("</section>\n")
}

func writeTypeHTML(sb *strings.Builder, t TypeDoc) {
	sb.WriteString(fmt.Sprintf("<section class=\"type-def\" id=\"%s\">\n", strings.ToLower(t.Name)))
	sb.WriteString(fmt.Sprintf("<h3>%s</h3>\n", html.EscapeString(t.Name)))

	if len(t.Fields) > 0 {
		sb.WriteString("<table>\n<tr><th>Field</th><th>Type</th><th>Required</th></tr>\n")
		for _, f := range t.Fields {
			req := "No"
			if f.Required {
				req = "Yes"
			}
			sb.WriteString(fmt.Sprintf("<tr><td><code>%s</code></td><td>%s</td><td>%s</td></tr>\n",
				html.EscapeString(f.Name), html.EscapeString(f.Type), req))
		}
		sb.WriteString("</table>\n")
	}

	sb.WriteString("</section>\n")
}

func writeCSS(sb *strings.Builder) {
	sb.WriteString(`<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; display: flex; min-height: 100vh; color: #333; }
.sidebar { width: 280px; background: #1a1a2e; color: #eee; padding: 20px; position: fixed; height: 100vh; overflow-y: auto; }
.sidebar h2 { margin-bottom: 16px; font-size: 1.2em; }
.sidebar h3 { margin-top: 16px; margin-bottom: 8px; font-size: 0.9em; text-transform: uppercase; color: #888; }
.sidebar ul { list-style: none; }
.sidebar li { margin-bottom: 4px; }
.sidebar a { color: #ccc; text-decoration: none; font-size: 0.9em; display: block; padding: 4px 8px; border-radius: 4px; }
.sidebar a:hover { background: #2a2a4e; color: #fff; }
#search { width: 100%; padding: 8px; border: 1px solid #444; border-radius: 4px; background: #2a2a4e; color: #eee; margin-bottom: 8px; }
.content { margin-left: 280px; padding: 40px; max-width: 900px; width: 100%; }
h1 { margin-bottom: 24px; }
h2 { margin-top: 32px; margin-bottom: 16px; border-bottom: 2px solid #eee; padding-bottom: 8px; }
.endpoint, .type-def { margin-bottom: 32px; padding: 20px; border: 1px solid #e0e0e0; border-radius: 8px; }
.endpoint h3, .type-def h3 { margin-bottom: 12px; }
h4 { margin-top: 12px; margin-bottom: 8px; font-size: 0.95em; }
table { border-collapse: collapse; width: 100%; margin-bottom: 12px; }
th, td { border: 1px solid #ddd; padding: 8px 12px; text-align: left; }
th { background: #f5f5f5; font-weight: 600; }
code { background: #f0f0f0; padding: 2px 6px; border-radius: 3px; font-size: 0.9em; }
.method { display: inline-block; padding: 2px 8px; border-radius: 3px; font-weight: 700; font-size: 0.8em; color: #fff; }
.method-get { background: #61affe; }
.method-post { background: #49cc90; }
.method-put { background: #fca130; }
.method-delete { background: #f93e3e; }
.method-patch { background: #50e3c2; }
.badge { display: inline-block; padding: 4px 10px; border-radius: 4px; font-size: 0.85em; margin-right: 8px; margin-bottom: 8px; }
.auth { background: #fff3cd; color: #856404; }
.rate-limit { background: #d1ecf1; color: #0c5460; }
</style>
`)
}

func writeScript(sb *strings.Builder) {
	sb.WriteString(`<script>
function filterDocs() {
  var query = document.getElementById('search').value.toLowerCase();
  var sections = document.querySelectorAll('.endpoint, .type-def');
  sections.forEach(function(s) {
    s.style.display = s.textContent.toLowerCase().includes(query) ? '' : 'none';
  });
  var links = document.querySelectorAll('.sidebar li');
  links.forEach(function(li) {
    li.style.display = li.textContent.toLowerCase().includes(query) ? '' : 'none';
  });
}
</script>
`)
}
