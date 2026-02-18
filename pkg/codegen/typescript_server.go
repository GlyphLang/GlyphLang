package codegen

import (
	"fmt"
	"strings"

	"github.com/glyphlang/glyph/pkg/ir"
)

// TypeScriptServerGenerator generates TypeScript/Express server code from the Semantic IR.
type TypeScriptServerGenerator struct {
	host string
	port int
}

// NewTypeScriptServerGenerator creates a new TypeScript/Express code generator.
func NewTypeScriptServerGenerator(host string, port int) *TypeScriptServerGenerator {
	if host == "" {
		host = "0.0.0.0"
	}
	if port == 0 {
		port = 3000
	}
	return &TypeScriptServerGenerator{host: host, port: port}
}

// Generate produces TypeScript/Express source code from a Semantic IR.
func (g *TypeScriptServerGenerator) Generate(service *ir.ServiceIR) string {
	var sb strings.Builder

	g.tsWriteHeader(&sb)
	g.tsWriteImports(&sb, service)
	g.tsWriteModels(&sb, service)
	g.tsWriteProviderStubs(&sb, service)
	sb.WriteString("\nconst app = express();\n")
	sb.WriteString("app.use(express.json());\n\n")
	g.tsWriteRoutes(&sb, service)
	g.tsWriteCronJobs(&sb, service)
	g.tsWriteEventHandlers(&sb, service)
	g.tsWriteQueueWorkers(&sb, service)
	g.tsWriteMain(&sb)

	return sb.String()
}

// GeneratePackageJSON produces a package.json file.
func (g *TypeScriptServerGenerator) GeneratePackageJSON(service *ir.ServiceIR) string {
	var sb strings.Builder
	sb.WriteString("{\n")
	sb.WriteString("  \"name\": \"glyph-generated-api\",\n")
	sb.WriteString("  \"version\": \"1.0.0\",\n")
	sb.WriteString("  \"description\": \"Auto-generated from GlyphLang\",\n")
	sb.WriteString("  \"main\": \"dist/app.js\",\n")
	sb.WriteString("  \"scripts\": {\n")
	sb.WriteString("    \"dev\": \"ts-node src/app.ts\",\n")
	sb.WriteString("    \"build\": \"tsc\",\n")
	sb.WriteString("    \"start\": \"node dist/app.js\"\n")
	sb.WriteString("  },\n")

	// Dependencies
	deps := []string{
		`"express": "^4.18.0"`,
	}
	for _, p := range service.Providers {
		switch p.ProviderType {
		case "Database":
			deps = append(deps, `"pg": "^8.11.0"`)
		case "Redis":
			deps = append(deps, `"redis": "^4.6.0"`)
		case "MongoDB":
			deps = append(deps, `"mongodb": "^6.3.0"`)
		case "LLM":
			deps = append(deps, `"@anthropic-ai/sdk": "^0.30.0"`)
		}
	}
	if len(service.CronJobs) > 0 {
		deps = append(deps, `"node-cron": "^3.0.0"`)
	}

	sb.WriteString("  \"dependencies\": {\n")
	for i, dep := range deps {
		sb.WriteString("    " + dep)
		if i < len(deps)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString("  },\n")

	// Dev dependencies
	devDeps := []string{
		`"typescript": "^5.0.0"`,
		`"ts-node": "^10.0.0"`,
		`"@types/node": "^20.0.0"`,
		`"@types/express": "^4.17.0"`,
	}
	if len(service.CronJobs) > 0 {
		devDeps = append(devDeps, `"@types/node-cron": "^3.0.0"`)
	}

	sb.WriteString("  \"devDependencies\": {\n")
	for i, dep := range devDeps {
		sb.WriteString("    " + dep)
		if i < len(devDeps)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString("  }\n")
	sb.WriteString("}\n")
	return sb.String()
}

// GenerateTSConfig produces a tsconfig.json file.
func (g *TypeScriptServerGenerator) GenerateTSConfig() string {
	return `{
  "compilerOptions": {
    "target": "ES2020",
    "module": "commonjs",
    "lib": ["ES2020"],
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist"]
}
`
}

func (g *TypeScriptServerGenerator) tsWriteHeader(sb *strings.Builder) {
	sb.WriteString("// Auto-generated TypeScript/Express server from GlyphLang\n")
	sb.WriteString("// Do not edit manually\n\n")
}

func (g *TypeScriptServerGenerator) tsWriteImports(sb *strings.Builder, service *ir.ServiceIR) {
	sb.WriteString("import express, { Request, Response } from 'express';\n")

	for _, p := range service.Providers {
		switch p.ProviderType {
		case "Database":
			sb.WriteString("import { Pool } from 'pg';\n")
		case "Redis":
			sb.WriteString("import { createClient } from 'redis';\n")
		case "MongoDB":
			sb.WriteString("import { MongoClient, Db } from 'mongodb';\n")
		}
	}

	if len(service.CronJobs) > 0 {
		sb.WriteString("import cron from 'node-cron';\n")
	}
	if len(service.Events) > 0 {
		sb.WriteString("import { EventEmitter } from 'events';\n")
	}

	sb.WriteString("\n")
}

func (g *TypeScriptServerGenerator) tsWriteModels(sb *strings.Builder, service *ir.ServiceIR) {
	for _, t := range service.Types {
		g.tsWriteModel(sb, t)
		sb.WriteString("\n")
	}
}

func (g *TypeScriptServerGenerator) tsWriteModel(sb *strings.Builder, t ir.TypeSchema) {
	fmt.Fprintf(sb, "interface %s {\n", t.Name)
	for _, f := range t.Fields {
		tsType := irTypeToTypeScript(f.Type)
		if !f.Required {
			fmt.Fprintf(sb, "  %s?: %s;\n", f.Name, tsType)
		} else if f.HasDefault {
			fmt.Fprintf(sb, "  %s: %s;\n", f.Name, tsType)
		} else {
			fmt.Fprintf(sb, "  %s: %s;\n", f.Name, tsType)
		}
	}
	sb.WriteString("}\n")
}

func (g *TypeScriptServerGenerator) tsWriteProviderStubs(sb *strings.Builder, service *ir.ServiceIR) {
	for _, p := range service.Providers {
		switch p.ProviderType {
		case "Database":
			sb.WriteString("\n// Database provider stub - replace with actual implementation\n")
			sb.WriteString("class TableProxy {\n")
			sb.WriteString("  constructor(private name: string) {}\n")
			sb.WriteString("  async Get(id: any): Promise<any> { throw new Error('Not implemented'); }\n")
			sb.WriteString("  async Find(filter?: any): Promise<any[]> { throw new Error('Not implemented'); }\n")
			sb.WriteString("  async Create(data: any): Promise<any> { throw new Error('Not implemented'); }\n")
			sb.WriteString("  async Update(id: any, data: any): Promise<any> { throw new Error('Not implemented'); }\n")
			sb.WriteString("  async Delete(id: any): Promise<void> { throw new Error('Not implemented'); }\n")
			sb.WriteString("  async Where(filter: any): Promise<any[]> { throw new Error('Not implemented'); }\n")
			sb.WriteString("}\n\n")
			sb.WriteString("class DatabaseProvider {\n")
			sb.WriteString("  private tables: Record<string, TableProxy> = {};\n")
			sb.WriteString("  [key: string]: any;\n\n")
			sb.WriteString("  constructor() {\n")
			sb.WriteString("    return new Proxy(this, {\n")
			sb.WriteString("      get: (target, prop: string) => {\n")
			sb.WriteString("        if (prop in target) return (target as any)[prop];\n")
			sb.WriteString("        if (!target.tables[prop]) target.tables[prop] = new TableProxy(prop);\n")
			sb.WriteString("        return target.tables[prop];\n")
			sb.WriteString("      }\n")
			sb.WriteString("    });\n")
			sb.WriteString("  }\n")
			sb.WriteString("}\n\n")
			sb.WriteString("const dbProvider = new DatabaseProvider();\n")
			sb.WriteString("function getDb(): DatabaseProvider { return dbProvider; }\n\n")
		case "Redis":
			sb.WriteString("\n// Redis provider stub\n")
			sb.WriteString("const redisClient = createClient();\n")
			sb.WriteString("function getRedis() { return redisClient; }\n\n")
		default:
			if !p.IsStandard {
				fmt.Fprintf(sb, "\n// Custom provider stub: %s\n", p.ProviderType)
				fmt.Fprintf(sb, "class %sProvider {\n", p.ProviderType)
				sb.WriteString("  // Implement provider methods as needed\n")
				sb.WriteString("}\n\n")
				varName := tsCamelCase(p.ProviderType) + "Provider"
				fmt.Fprintf(sb, "const %s = new %sProvider();\n", varName, p.ProviderType)
				fmt.Fprintf(sb, "function get%s(): %sProvider { return %s; }\n\n", p.ProviderType, p.ProviderType, varName)
			}
		}
	}
}

func (g *TypeScriptServerGenerator) tsWriteRoutes(sb *strings.Builder, service *ir.ServiceIR) {
	for _, route := range service.Routes {
		g.tsWriteRoute(sb, route)
	}
}

func (g *TypeScriptServerGenerator) tsWriteRoute(sb *strings.Builder, route ir.RouteHandler) {
	method := strings.ToLower(route.Method.String())
	expressPath := glyphPathToExpress(route.Path)

	// Auth/rate limit comments
	if route.Auth != nil {
		fmt.Fprintf(sb, "// Requires %s authentication\n", route.Auth.AuthType)
	}
	if route.RateLimit != nil {
		fmt.Fprintf(sb, "// Rate limited: %d requests per %s\n", route.RateLimit.Requests, route.RateLimit.Window)
	}

	fmt.Fprintf(sb, "app.%s('%s', async (req: Request, res: Response) => {\n", method, expressPath)

	// Extract path params
	for _, pp := range route.PathParams {
		tsWriteIndent(sb, 1)
		fmt.Fprintf(sb, "const %s = req.params.%s;\n", pp, pp)
	}

	// Inject providers
	for _, prov := range route.Providers {
		tsWriteIndent(sb, 1)
		depFunc := tsProviderToGetterFunc(prov.ProviderType)
		fmt.Fprintf(sb, "const %s = %s();\n", prov.Name, depFunc)
	}

	// Extract input body for POST/PUT/PATCH
	if route.InputType != nil && (method == "post" || method == "put" || method == "patch") {
		inputTypeName := tsTypeNameForInput(route.InputType)
		tsWriteIndent(sb, 1)
		fmt.Fprintf(sb, "const input: %s = req.body;\n", inputTypeName)
	}

	// Write body statements
	g.tsWriteStatements(sb, route.Body, 1)

	sb.WriteString("});\n\n")
}

func (g *TypeScriptServerGenerator) tsWriteStatements(sb *strings.Builder, stmts []ir.StmtIR, indent int) {
	if len(stmts) == 0 {
		tsWriteIndent(sb, indent)
		sb.WriteString("// no-op\n")
		return
	}
	for _, stmt := range stmts {
		g.tsWriteStatement(sb, stmt, indent)
	}
}

func (g *TypeScriptServerGenerator) tsWriteStatement(sb *strings.Builder, stmt ir.StmtIR, indent int) {
	switch stmt.Kind {
	case ir.StmtAssign:
		tsWriteIndent(sb, indent)
		fmt.Fprintf(sb, "const %s = ", stmt.Assign.Target)
		g.tsWriteExpr(sb, stmt.Assign.Value)
		sb.WriteString(";\n")
	case ir.StmtReassign:
		tsWriteIndent(sb, indent)
		fmt.Fprintf(sb, "%s = ", stmt.Assign.Target)
		g.tsWriteExpr(sb, stmt.Assign.Value)
		sb.WriteString(";\n")
	case ir.StmtReturn:
		tsWriteIndent(sb, indent)
		sb.WriteString("return res.json(")
		g.tsWriteExpr(sb, stmt.Return.Value)
		sb.WriteString(");\n")
	case ir.StmtIf:
		tsWriteIndent(sb, indent)
		sb.WriteString("if (")
		g.tsWriteExpr(sb, stmt.If.Condition)
		sb.WriteString(") {\n")
		g.tsWriteStatements(sb, stmt.If.Then, indent+1)
		tsWriteIndent(sb, indent)
		sb.WriteString("}")
		if len(stmt.If.Else) > 0 {
			sb.WriteString(" else {\n")
			g.tsWriteStatements(sb, stmt.If.Else, indent+1)
			tsWriteIndent(sb, indent)
			sb.WriteString("}")
		}
		sb.WriteString("\n")
	case ir.StmtFor:
		tsWriteIndent(sb, indent)
		if stmt.For.KeyVar != "" {
			fmt.Fprintf(sb, "for (const [%s, %s] of ", stmt.For.KeyVar, stmt.For.ValueVar)
			g.tsWriteExpr(sb, stmt.For.Iterable)
			sb.WriteString(".entries()) {\n")
		} else {
			fmt.Fprintf(sb, "for (const %s of ", stmt.For.ValueVar)
			g.tsWriteExpr(sb, stmt.For.Iterable)
			sb.WriteString(") {\n")
		}
		g.tsWriteStatements(sb, stmt.For.Body, indent+1)
		tsWriteIndent(sb, indent)
		sb.WriteString("}\n")
	case ir.StmtWhile:
		tsWriteIndent(sb, indent)
		sb.WriteString("while (")
		g.tsWriteExpr(sb, stmt.While.Condition)
		sb.WriteString(") {\n")
		g.tsWriteStatements(sb, stmt.While.Body, indent+1)
		tsWriteIndent(sb, indent)
		sb.WriteString("}\n")
	case ir.StmtExpr:
		if stmt.ExprStmt != nil {
			tsWriteIndent(sb, indent)
			g.tsWriteExpr(sb, *stmt.ExprStmt)
			sb.WriteString(";\n")
		}
	case ir.StmtValidate:
		tsWriteIndent(sb, indent)
		sb.WriteString("// validate: ")
		g.tsWriteExpr(sb, stmt.Validate.Call)
		sb.WriteString("\n")
	case ir.StmtBreak:
		tsWriteIndent(sb, indent)
		sb.WriteString("break;\n")
	case ir.StmtContinue:
		tsWriteIndent(sb, indent)
		sb.WriteString("continue;\n")
	}
}

func (g *TypeScriptServerGenerator) tsWriteExpr(sb *strings.Builder, expr ir.ExprIR) {
	switch expr.Kind {
	case ir.ExprInt:
		fmt.Fprintf(sb, "%d", expr.IntVal)
	case ir.ExprFloat:
		fmt.Fprintf(sb, "%g", expr.FloatVal)
	case ir.ExprString:
		fmt.Fprintf(sb, "%q", expr.StringVal)
	case ir.ExprBool:
		if expr.BoolVal {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
	case ir.ExprNull:
		sb.WriteString("null")
	case ir.ExprVar:
		sb.WriteString(expr.VarName)
	case ir.ExprBinary:
		sb.WriteString("(")
		g.tsWriteExpr(sb, expr.BinOp.Left)
		sb.WriteString(" ")
		sb.WriteString(binOpToTypeScript(expr.BinOp.Op))
		sb.WriteString(" ")
		g.tsWriteExpr(sb, expr.BinOp.Right)
		sb.WriteString(")")
	case ir.ExprUnary:
		sb.WriteString(unaryOpToTypeScript(expr.UnaryOp.Op))
		g.tsWriteExpr(sb, expr.UnaryOp.Right)
	case ir.ExprFieldAccess:
		g.tsWriteExpr(sb, expr.FieldAccess.Object)
		fmt.Fprintf(sb, ".%s", expr.FieldAccess.Field)
	case ir.ExprIndexAccess:
		g.tsWriteExpr(sb, expr.IndexAccess.Object)
		sb.WriteString("[")
		g.tsWriteExpr(sb, expr.IndexAccess.Index)
		sb.WriteString("]")
	case ir.ExprCall:
		// Method call detection — same pattern as Python generator
		if len(expr.Call.Args) > 0 && !strings.Contains(expr.Call.Name, ".") && isObjectReceiver(expr.Call.Args[0]) {
			g.tsWriteExpr(sb, expr.Call.Args[0])
			fmt.Fprintf(sb, ".%s(", expr.Call.Name)
			for i, arg := range expr.Call.Args[1:] {
				if i > 0 {
					sb.WriteString(", ")
				}
				g.tsWriteExpr(sb, arg)
			}
			sb.WriteString(")")
		} else {
			sb.WriteString(expr.Call.Name)
			sb.WriteString("(")
			for i, arg := range expr.Call.Args {
				if i > 0 {
					sb.WriteString(", ")
				}
				g.tsWriteExpr(sb, arg)
			}
			sb.WriteString(")")
		}
	case ir.ExprObject:
		sb.WriteString("{ ")
		for i, f := range expr.Object.Fields {
			if i > 0 {
				sb.WriteString(", ")
			}
			fmt.Fprintf(sb, "%s: ", f.Key)
			g.tsWriteExpr(sb, f.Value)
		}
		sb.WriteString(" }")
	case ir.ExprArray:
		sb.WriteString("[")
		for i, el := range expr.Array.Elements {
			if i > 0 {
				sb.WriteString(", ")
			}
			g.tsWriteExpr(sb, el)
		}
		sb.WriteString("]")
	case ir.ExprLambda:
		sb.WriteString("(")
		for i, p := range expr.Lambda.Params {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(p.Name)
		}
		sb.WriteString(") => ")
		g.tsWriteExpr(sb, expr.Lambda.Body)
	}
}

func (g *TypeScriptServerGenerator) tsWriteCronJobs(sb *strings.Builder, service *ir.ServiceIR) {
	if len(service.CronJobs) == 0 {
		return
	}
	sb.WriteString("// --- Cron Jobs ---\n\n")
	for _, cronJob := range service.CronJobs {
		name := cronJob.Name
		if name == "" {
			name = "cronTask"
		}
		fmt.Fprintf(sb, "cron.schedule('%s', async () => {\n", cronJob.Schedule)
		// Inject providers
		for _, prov := range cronJob.Providers {
			tsWriteIndent(sb, 1)
			depFunc := tsProviderToGetterFunc(prov.ProviderType)
			fmt.Fprintf(sb, "const %s = %s();\n", prov.Name, depFunc)
		}
		g.tsWriteNonRouteStatements(sb, cronJob.Body, 1)
		fmt.Fprintf(sb, "}); // %s\n\n", name)
	}
}

func (g *TypeScriptServerGenerator) tsWriteEventHandlers(sb *strings.Builder, service *ir.ServiceIR) {
	if len(service.Events) == 0 {
		return
	}
	sb.WriteString("// --- Event Handlers ---\n")
	sb.WriteString("const eventBus = new EventEmitter();\n\n")
	for _, ev := range service.Events {
		funcName := "handle_" + strings.ReplaceAll(ev.EventType, ".", "_")
		fmt.Fprintf(sb, "eventBus.on('%s', async (event: any) => {\n", ev.EventType)
		for _, prov := range ev.Providers {
			tsWriteIndent(sb, 1)
			depFunc := tsProviderToGetterFunc(prov.ProviderType)
			fmt.Fprintf(sb, "const %s = %s();\n", prov.Name, depFunc)
		}
		g.tsWriteNonRouteStatements(sb, ev.Body, 1)
		fmt.Fprintf(sb, "}); // %s\n\n", funcName)
	}
}

func (g *TypeScriptServerGenerator) tsWriteQueueWorkers(sb *strings.Builder, service *ir.ServiceIR) {
	if len(service.Queues) == 0 {
		return
	}
	sb.WriteString("// --- Queue Workers ---\n")
	sb.WriteString("// Implement with BullMQ, SQS, or your preferred task queue\n\n")
	for _, q := range service.Queues {
		funcName := "worker_" + strings.ReplaceAll(q.QueueName, ".", "_")
		fmt.Fprintf(sb, "async function %s(message: any): Promise<void> {\n", funcName)
		for _, prov := range q.Providers {
			tsWriteIndent(sb, 1)
			depFunc := tsProviderToGetterFunc(prov.ProviderType)
			fmt.Fprintf(sb, "const %s = %s();\n", prov.Name, depFunc)
		}
		g.tsWriteNonRouteStatements(sb, q.Body, 1)
		sb.WriteString("}\n\n")
	}
}

// tsWriteNonRouteStatements writes statements for non-route contexts (cron, events, queues)
// where return statements should use plain `return` instead of `res.json()`.
func (g *TypeScriptServerGenerator) tsWriteNonRouteStatements(sb *strings.Builder, stmts []ir.StmtIR, indent int) {
	if len(stmts) == 0 {
		tsWriteIndent(sb, indent)
		sb.WriteString("// no-op\n")
		return
	}
	for _, stmt := range stmts {
		if stmt.Kind == ir.StmtReturn {
			tsWriteIndent(sb, indent)
			sb.WriteString("return ")
			g.tsWriteExpr(sb, stmt.Return.Value)
			sb.WriteString(";\n")
		} else {
			g.tsWriteStatement(sb, stmt, indent)
		}
	}
}

func (g *TypeScriptServerGenerator) tsWriteMain(sb *strings.Builder) {
	fmt.Fprintf(sb, "app.listen(%d, '%s', () => {\n", g.port, g.host)
	fmt.Fprintf(sb, "  console.log(`Server running on http://%s:%d`);\n", g.host, g.port)
	sb.WriteString("});\n")
}

// --- Helper functions ---

func irTypeToTypeScript(t ir.TypeRef) string {
	switch t.Kind {
	case ir.TypeInt, ir.TypeFloat:
		return "number"
	case ir.TypeString:
		return "string"
	case ir.TypeBool:
		return "boolean"
	case ir.TypeArray:
		if t.Inner != nil {
			return irTypeToTypeScript(*t.Inner) + "[]"
		}
		return "any[]"
	case ir.TypeOptional:
		if t.Inner != nil {
			return irTypeToTypeScript(*t.Inner) + " | null"
		}
		return "any | null"
	case ir.TypeNamed:
		return t.Name
	case ir.TypeProvider:
		return t.Name
	case ir.TypeUnion:
		if len(t.Elements) > 0 {
			var parts []string
			for _, elem := range t.Elements {
				parts = append(parts, irTypeToTypeScript(elem))
			}
			return strings.Join(parts, " | ")
		}
		return "any"
	case ir.TypeAny:
		return "any"
	default:
		return "any"
	}
}

func glyphPathToExpress(path string) string {
	// Express uses :param syntax (same as Glyph), so no conversion needed
	return path
}

func tsProviderToGetterFunc(providerType string) string {
	switch providerType {
	case "Database":
		return "getDb"
	case "Redis":
		return "getRedis"
	default:
		return "get" + providerType
	}
}

func tsTypeNameForInput(t *ir.TypeRef) string {
	if t == nil {
		return "any"
	}
	if t.Kind == ir.TypeNamed {
		return t.Name
	}
	return "any"
}

func tsCamelCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func binOpToTypeScript(op ir.BinOp) string {
	switch op {
	case ir.OpAdd:
		return "+"
	case ir.OpSub:
		return "-"
	case ir.OpMul:
		return "*"
	case ir.OpDiv:
		return "/"
	case ir.OpMod:
		return "%"
	case ir.OpEq:
		return "==="
	case ir.OpNe:
		return "!=="
	case ir.OpLt:
		return "<"
	case ir.OpLe:
		return "<="
	case ir.OpGt:
		return ">"
	case ir.OpGe:
		return ">="
	case ir.OpAnd:
		return "&&"
	case ir.OpOr:
		return "||"
	default:
		return "+"
	}
}

func unaryOpToTypeScript(op ir.UnOp) string {
	switch op {
	case ir.OpNot:
		return "!"
	case ir.OpNeg:
		return "-"
	default:
		return ""
	}
}

func tsWriteIndent(sb *strings.Builder, level int) {
	for i := 0; i < level; i++ {
		sb.WriteString("  ")
	}
}
