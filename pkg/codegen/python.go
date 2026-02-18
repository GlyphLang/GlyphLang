package codegen

import (
	"fmt"
	"strings"

	"github.com/glyphlang/glyph/pkg/ir"
)

// PythonGenerator generates Python/FastAPI server code from the Semantic IR.
type PythonGenerator struct {
	host string
	port int
}

// NewPythonGenerator creates a new Python code generator.
func NewPythonGenerator(host string, port int) *PythonGenerator {
	if host == "" {
		host = "0.0.0.0"
	}
	if port == 0 {
		port = 8000
	}
	return &PythonGenerator{host: host, port: port}
}

// Generate produces Python/FastAPI source code from a Semantic IR.
func (g *PythonGenerator) Generate(service *ir.ServiceIR) string {
	var sb strings.Builder

	g.writeHeader(&sb)
	g.writeImports(&sb, service)
	sb.WriteString("\napp = FastAPI()\n\n")
	g.writeModels(&sb, service)
	g.writeProviderStubs(&sb, service)
	g.writeRoutes(&sb, service)
	g.writeGraphQL(&sb, service)
	g.writeGRPC(&sb, service)
	g.writeWebSockets(&sb, service)
	g.writeCronJobs(&sb, service)
	g.writeEventHandlers(&sb, service)
	g.writeQueueWorkers(&sb, service)
	g.writeMain(&sb)

	return sb.String()
}

// GenerateRequirements produces a requirements.txt file.
func (g *PythonGenerator) GenerateRequirements(service *ir.ServiceIR) string {
	deps := []string{
		"fastapi>=0.100.0",
		"uvicorn>=0.23.0",
		"pydantic>=2.0.0",
	}

	for _, p := range service.Providers {
		switch p.ProviderType {
		case "Database":
			deps = append(deps, "sqlalchemy>=2.0.0", "databases>=0.9.0")
		case "Redis":
			deps = append(deps, "redis>=5.0.0")
		case "MongoDB":
			deps = append(deps, "motor>=3.3.0", "pymongo>=4.6.0")
		case "LLM":
			deps = append(deps, "anthropic>=0.40.0")
		}
	}

	if len(service.CronJobs) > 0 {
		deps = append(deps, "apscheduler>=3.10.0")
	}
	if len(service.Queues) > 0 {
		deps = append(deps, "celery>=5.3.0")
	}
	if len(service.WebSocket) > 0 {
		deps = append(deps, "websockets>=12.0")
	}
	if len(service.GraphQL) > 0 {
		deps = append(deps, "strawberry-graphql>=0.220.0")
	}
	if len(service.GRPC) > 0 {
		deps = append(deps, "grpcio>=1.60.0", "grpcio-tools>=1.60.0", "protobuf>=4.25.0")
	}

	return strings.Join(deps, "\n") + "\n"
}

func (g *PythonGenerator) writeHeader(sb *strings.Builder) {
	sb.WriteString("# Auto-generated Python/FastAPI server from GlyphLang\n")
	sb.WriteString("# Do not edit manually\n\n")
}

func (g *PythonGenerator) writeImports(sb *strings.Builder, service *ir.ServiceIR) {
	sb.WriteString("from fastapi import FastAPI, HTTPException, Depends, Request\n")
	sb.WriteString("from pydantic import BaseModel\n")
	sb.WriteString("from typing import Optional, List, Any\n")
	sb.WriteString("import uuid\n")
	sb.WriteString("import time\n")

	for _, p := range service.Providers {
		switch p.ProviderType {
		case "Database":
			sb.WriteString("from sqlalchemy.orm import Session\n")
		case "Redis":
			sb.WriteString("import redis\n")
		}
	}

	if len(service.CronJobs) > 0 {
		sb.WriteString("from apscheduler.schedulers.asyncio import AsyncIOScheduler\n")
		sb.WriteString("from apscheduler.triggers.cron import CronTrigger\n")
	}

	if len(service.WebSocket) > 0 {
		sb.WriteString("from fastapi import WebSocket, WebSocketDisconnect\n")
	}
	if len(service.GraphQL) > 0 {
		sb.WriteString("import strawberry\n")
		sb.WriteString("from strawberry.fastapi import GraphQLRouter\n")
	}
	if len(service.GRPC) > 0 {
		sb.WriteString("import grpc\n")
		sb.WriteString("from grpc import aio as grpc_aio\n")
		sb.WriteString("from concurrent import futures\n")
	}

	sb.WriteString("\n")
}

func (g *PythonGenerator) writeModels(sb *strings.Builder, service *ir.ServiceIR) {
	for _, t := range service.Types {
		g.writeModel(sb, t)
		sb.WriteString("\n")
	}
}

func (g *PythonGenerator) writeModel(sb *strings.Builder, t ir.TypeSchema) {
	fmt.Fprintf(sb, "\nclass %s(BaseModel):\n", t.Name)
	if len(t.Fields) == 0 {
		sb.WriteString("    pass\n")
		return
	}
	for _, f := range t.Fields {
		pyType := irTypeToPython(f.Type)
		if !f.Required {
			// Only wrap in Optional if the IR type isn't already Optional
			// (e.g., `str?` produces TypeOptional(TypeString), which irTypeToPython
			// already renders as "Optional[str]")
			if f.Type.Kind != ir.TypeOptional {
				pyType = fmt.Sprintf("Optional[%s]", pyType)
			}
			fmt.Fprintf(sb, "    %s: %s = None\n", f.Name, pyType)
		} else if f.HasDefault {
			switch f.Default.Kind {
			case ir.ExprBool:
				def := "False"
				if f.Default.BoolVal {
					def = "True"
				}
				fmt.Fprintf(sb, "    %s: %s = %s\n", f.Name, pyType, def)
			case ir.ExprString:
				fmt.Fprintf(sb, "    %s: %s = %q\n", f.Name, pyType, f.Default.StringVal)
			case ir.ExprInt:
				fmt.Fprintf(sb, "    %s: %s = %d\n", f.Name, pyType, f.Default.IntVal)
			case ir.ExprFloat:
				fmt.Fprintf(sb, "    %s: %s = %g\n", f.Name, pyType, f.Default.FloatVal)
			default:
				fmt.Fprintf(sb, "    %s: %s\n", f.Name, pyType)
			}
		} else {
			fmt.Fprintf(sb, "    %s: %s\n", f.Name, pyType)
		}
	}
}

func (g *PythonGenerator) writeProviderStubs(sb *strings.Builder, service *ir.ServiceIR) {
	for _, p := range service.Providers {
		switch p.ProviderType {
		case "Database":
			sb.WriteString("\n# Database provider stub - replace with actual implementation\n")
			sb.WriteString("class DatabaseProvider:\n")
			sb.WriteString("    \"\"\"Abstract database provider. Implement with SQLAlchemy, Prisma, etc.\"\"\"\n")
			sb.WriteString("    def __init__(self):\n")
			sb.WriteString("        self._tables = {}\n\n")
			sb.WriteString("    def __getattr__(self, name):\n")
			sb.WriteString("        if name.startswith('_'):\n")
			sb.WriteString("            raise AttributeError(name)\n")
			sb.WriteString("        if name not in self._tables:\n")
			sb.WriteString("            self._tables[name] = TableProxy(name)\n")
			sb.WriteString("        return self._tables[name]\n\n")
			sb.WriteString("class TableProxy:\n")
			sb.WriteString("    def __init__(self, name): self.name = name\n")
			sb.WriteString("    def Get(self, id): raise NotImplementedError\n")
			sb.WriteString("    def Find(self, filter=None): raise NotImplementedError\n")
			sb.WriteString("    def Create(self, data): raise NotImplementedError\n")
			sb.WriteString("    def Update(self, id, data): raise NotImplementedError\n")
			sb.WriteString("    def Delete(self, id): raise NotImplementedError\n")
			sb.WriteString("    def Where(self, filter): raise NotImplementedError\n\n")
			sb.WriteString("db_provider = DatabaseProvider()\n\n")
			sb.WriteString("def get_db() -> DatabaseProvider:\n")
			sb.WriteString("    return db_provider\n\n")
		case "Redis":
			sb.WriteString("\n# Redis provider stub\n")
			sb.WriteString("redis_client = redis.Redis(host='localhost', port=6379, decode_responses=True)\n\n")
			sb.WriteString("def get_redis() -> redis.Redis:\n")
			sb.WriteString("    return redis_client\n\n")
		default:
			if !p.IsStandard {
				fmt.Fprintf(sb, "\n# Custom provider stub: %s\n", p.ProviderType)
				fmt.Fprintf(sb, "class %sProvider:\n", p.ProviderType)
				sb.WriteString("    \"\"\"Custom provider - implement methods as needed.\"\"\"\n")
				sb.WriteString("    pass\n\n")
				varName := strings.ToLower(p.ProviderType) + "_provider"
				fmt.Fprintf(sb, "%s = %sProvider()\n\n", varName, p.ProviderType)
				fmt.Fprintf(sb, "def get_%s() -> %sProvider:\n", strings.ToLower(p.ProviderType), p.ProviderType)
				fmt.Fprintf(sb, "    return %s\n\n", varName)
			}
		}
	}
}

func (g *PythonGenerator) writeRoutes(sb *strings.Builder, service *ir.ServiceIR) {
	for _, route := range service.Routes {
		g.writeRoute(sb, route)
	}
}

func (g *PythonGenerator) writeRoute(sb *strings.Builder, route ir.RouteHandler) {
	// Determine FastAPI decorator
	method := strings.ToLower(route.Method.String())
	fastapiPath := glyphPathToFastAPI(route.Path)

	// Build function parameters
	var params []string
	for _, pp := range route.PathParams {
		params = append(params, fmt.Sprintf("%s: str", pp))
	}

	// Add provider dependencies
	for _, prov := range route.Providers {
		depFunc := providerToDependsFunc(prov.ProviderType)
		params = append(params, fmt.Sprintf("%s = Depends(%s)", prov.Name, depFunc))
	}

	// Add input body if present
	if route.InputType != nil && (method == "post" || method == "put" || method == "patch") {
		inputTypeName := irTypeNameForInput(route.InputType)
		params = append(params, fmt.Sprintf("input: %s", inputTypeName))
	}

	paramStr := strings.Join(params, ", ")

	// Status code
	statusCode := ""
	if method == "post" {
		statusCode = ", status_code=201"
	}

	fmt.Fprintf(sb, "\n@app.%s(\"%s\"%s)\n", method, fastapiPath, statusCode)
	fmt.Fprintf(sb, "async def %s(%s):\n", routeToFuncName(route), paramStr)

	// Auth check comment
	if route.Auth != nil {
		fmt.Fprintf(sb, "    # Requires %s authentication\n", route.Auth.AuthType)
	}
	if route.RateLimit != nil {
		fmt.Fprintf(sb, "    # Rate limited: %d requests per %s\n", route.RateLimit.Requests, route.RateLimit.Window)
	}

	// Write body
	g.writeStatements(sb, route.Body, 1)
	sb.WriteString("\n")
}

func (g *PythonGenerator) writeStatements(sb *strings.Builder, stmts []ir.StmtIR, indent int) {
	if len(stmts) == 0 {
		writeIndent(sb, indent)
		sb.WriteString("pass\n")
		return
	}
	for _, stmt := range stmts {
		g.writeStatement(sb, stmt, indent)
	}
}

func (g *PythonGenerator) writeStatement(sb *strings.Builder, stmt ir.StmtIR, indent int) {
	switch stmt.Kind {
	case ir.StmtAssign:
		writeIndent(sb, indent)
		fmt.Fprintf(sb, "%s = ", stmt.Assign.Target)
		g.writeExpr(sb, stmt.Assign.Value)
		sb.WriteString("\n")
	case ir.StmtReassign:
		writeIndent(sb, indent)
		fmt.Fprintf(sb, "%s = ", stmt.Assign.Target)
		g.writeExpr(sb, stmt.Assign.Value)
		sb.WriteString("\n")
	case ir.StmtReturn:
		writeIndent(sb, indent)
		sb.WriteString("return ")
		g.writeExpr(sb, stmt.Return.Value)
		sb.WriteString("\n")
	case ir.StmtIf:
		writeIndent(sb, indent)
		sb.WriteString("if ")
		g.writeExpr(sb, stmt.If.Condition)
		sb.WriteString(":\n")
		g.writeStatements(sb, stmt.If.Then, indent+1)
		if len(stmt.If.Else) > 0 {
			writeIndent(sb, indent)
			sb.WriteString("else:\n")
			g.writeStatements(sb, stmt.If.Else, indent+1)
		}
	case ir.StmtFor:
		writeIndent(sb, indent)
		if stmt.For.KeyVar != "" {
			fmt.Fprintf(sb, "for %s, %s in enumerate(", stmt.For.KeyVar, stmt.For.ValueVar)
		} else {
			fmt.Fprintf(sb, "for %s in ", stmt.For.ValueVar)
		}
		g.writeExpr(sb, stmt.For.Iterable)
		if stmt.For.KeyVar != "" {
			sb.WriteString(")")
		}
		sb.WriteString(":\n")
		g.writeStatements(sb, stmt.For.Body, indent+1)
	case ir.StmtWhile:
		writeIndent(sb, indent)
		sb.WriteString("while ")
		g.writeExpr(sb, stmt.While.Condition)
		sb.WriteString(":\n")
		g.writeStatements(sb, stmt.While.Body, indent+1)
	case ir.StmtExpr:
		if stmt.ExprStmt != nil {
			writeIndent(sb, indent)
			g.writeExpr(sb, *stmt.ExprStmt)
			sb.WriteString("\n")
		}
	case ir.StmtValidate:
		writeIndent(sb, indent)
		sb.WriteString("# validate: ")
		g.writeExpr(sb, stmt.Validate.Call)
		sb.WriteString("\n")
	case ir.StmtSwitch:
		if stmt.Switch != nil {
			for i, c := range stmt.Switch.Cases {
				writeIndent(sb, indent)
				if i == 0 {
					sb.WriteString("if ")
				} else {
					sb.WriteString("elif ")
				}
				g.writeExpr(sb, stmt.Switch.Value)
				sb.WriteString(" == ")
				g.writeExpr(sb, c.Value)
				sb.WriteString(":\n")
				g.writeStatements(sb, c.Body, indent+1)
			}
			if len(stmt.Switch.Default) > 0 {
				writeIndent(sb, indent)
				sb.WriteString("else:\n")
				g.writeStatements(sb, stmt.Switch.Default, indent+1)
			}
		}
	case ir.StmtBreak:
		writeIndent(sb, indent)
		sb.WriteString("break\n")
	case ir.StmtContinue:
		writeIndent(sb, indent)
		sb.WriteString("continue\n")
	}
}

func (g *PythonGenerator) writeExpr(sb *strings.Builder, expr ir.ExprIR) {
	switch expr.Kind {
	case ir.ExprInt:
		fmt.Fprintf(sb, "%d", expr.IntVal)
	case ir.ExprFloat:
		fmt.Fprintf(sb, "%g", expr.FloatVal)
	case ir.ExprString:
		fmt.Fprintf(sb, "%q", expr.StringVal)
	case ir.ExprBool:
		if expr.BoolVal {
			sb.WriteString("True")
		} else {
			sb.WriteString("False")
		}
	case ir.ExprNull:
		sb.WriteString("None")
	case ir.ExprVar:
		sb.WriteString(expr.VarName)
	case ir.ExprBinary:
		sb.WriteString("(")
		g.writeExpr(sb, expr.BinOp.Left)
		sb.WriteString(" ")
		sb.WriteString(binOpToPython(expr.BinOp.Op))
		sb.WriteString(" ")
		g.writeExpr(sb, expr.BinOp.Right)
		sb.WriteString(")")
	case ir.ExprUnary:
		sb.WriteString(unaryOpToPython(expr.UnaryOp.Op))
		g.writeExpr(sb, expr.UnaryOp.Right)
	case ir.ExprFieldAccess:
		g.writeExpr(sb, expr.FieldAccess.Object)
		fmt.Fprintf(sb, ".%s", expr.FieldAccess.Field)
	case ir.ExprIndexAccess:
		g.writeExpr(sb, expr.IndexAccess.Object)
		sb.WriteString("[")
		g.writeExpr(sb, expr.IndexAccess.Index)
		sb.WriteString("]")
	case ir.ExprCall:
		// The Glyph parser converts method calls (obj.method(args)) into function
		// calls where the object is the first argument. Detect this pattern and
		// render as a method call: if the first arg is a field access or variable
		// and the function name is simple (no dots), it's a method call.
		if len(expr.Call.Args) > 0 && !strings.Contains(expr.Call.Name, ".") && isObjectReceiver(expr.Call.Args[0]) {
			g.writeExpr(sb, expr.Call.Args[0])
			fmt.Fprintf(sb, ".%s(", expr.Call.Name)
			for i, arg := range expr.Call.Args[1:] {
				if i > 0 {
					sb.WriteString(", ")
				}
				g.writeExpr(sb, arg)
			}
			sb.WriteString(")")
		} else {
			sb.WriteString(expr.Call.Name)
			sb.WriteString("(")
			for i, arg := range expr.Call.Args {
				if i > 0 {
					sb.WriteString(", ")
				}
				g.writeExpr(sb, arg)
			}
			sb.WriteString(")")
		}
	case ir.ExprObject:
		sb.WriteString("{")
		for i, f := range expr.Object.Fields {
			if i > 0 {
				sb.WriteString(", ")
			}
			fmt.Fprintf(sb, "%q: ", f.Key)
			g.writeExpr(sb, f.Value)
		}
		sb.WriteString("}")
	case ir.ExprArray:
		sb.WriteString("[")
		for i, el := range expr.Array.Elements {
			if i > 0 {
				sb.WriteString(", ")
			}
			g.writeExpr(sb, el)
		}
		sb.WriteString("]")
	case ir.ExprLambda:
		sb.WriteString("lambda ")
		for i, p := range expr.Lambda.Params {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(p.Name)
		}
		sb.WriteString(": ")
		g.writeExpr(sb, expr.Lambda.Body)
	case ir.ExprPipe:
		if expr.Pipe != nil {
			// Python pipe: right(left)
			g.writeExpr(sb, expr.Pipe.Right)
			sb.WriteString("(")
			g.writeExpr(sb, expr.Pipe.Left)
			sb.WriteString(")")
		}
	case ir.ExprMatch:
		if expr.Match != nil {
			// Python 3.10+ match/case
			g.writeMatchExpr(sb, expr.Match)
		}
	case ir.ExprAsync:
		if expr.Async != nil {
			// Python async: wrap in an async function and call
			sb.WriteString("asyncio.create_task((async lambda: None)())")
		}
	case ir.ExprAwait:
		if expr.Await != nil {
			sb.WriteString("await ")
			g.writeExpr(sb, expr.Await.Expr)
		}
	}
}

func (g *PythonGenerator) writeGraphQL(sb *strings.Builder, service *ir.ServiceIR) {
	if len(service.GraphQL) == 0 {
		return
	}
	sb.WriteString("\n# --- GraphQL Resolvers ---\n\n")

	// Separate queries, mutations, subscriptions
	var queries, mutations, subscriptions []ir.GraphQLDef
	for _, gql := range service.GraphQL {
		switch gql.Operation {
		case ir.GraphQLQuery:
			queries = append(queries, gql)
		case ir.GraphQLMutation:
			mutations = append(mutations, gql)
		case ir.GraphQLSubscription:
			subscriptions = append(subscriptions, gql)
		}
	}

	if len(queries) > 0 {
		sb.WriteString("@strawberry.type\n")
		sb.WriteString("class Query:\n")
		for _, q := range queries {
			g.writeGraphQLResolver(sb, q, 1)
		}
		sb.WriteString("\n")
	}

	if len(mutations) > 0 {
		sb.WriteString("@strawberry.type\n")
		sb.WriteString("class Mutation:\n")
		for _, m := range mutations {
			g.writeGraphQLResolver(sb, m, 1)
		}
		sb.WriteString("\n")
	}

	if len(subscriptions) > 0 {
		sb.WriteString("@strawberry.type\n")
		sb.WriteString("class Subscription:\n")
		for _, s := range subscriptions {
			g.writeGraphQLResolver(sb, s, 1)
		}
		sb.WriteString("\n")
	}

	// Create schema and mount router
	schemaArgs := []string{}
	if len(queries) > 0 {
		schemaArgs = append(schemaArgs, "query=Query")
	}
	if len(mutations) > 0 {
		schemaArgs = append(schemaArgs, "mutation=Mutation")
	}
	if len(subscriptions) > 0 {
		schemaArgs = append(schemaArgs, "subscription=Subscription")
	}
	fmt.Fprintf(sb, "schema = strawberry.Schema(%s)\n", strings.Join(schemaArgs, ", "))
	sb.WriteString("graphql_app = GraphQLRouter(schema)\n")
	sb.WriteString("app.include_router(graphql_app, prefix=\"/graphql\")\n\n")
}

func (g *PythonGenerator) writeGraphQLResolver(sb *strings.Builder, gql ir.GraphQLDef, indent int) {
	writeIndent(sb, indent)
	sb.WriteString("@strawberry.field\n")
	writeIndent(sb, indent)

	// Build params
	var params []string
	params = append(params, "self")
	for _, p := range gql.Params {
		pyType := irTypeToPython(p.Type)
		params = append(params, fmt.Sprintf("%s: %s", p.Name, pyType))
	}
	retType := "Any"
	if gql.ReturnType != nil {
		retType = irTypeToPython(*gql.ReturnType)
	}
	fmt.Fprintf(sb, "async def %s(%s) -> %s:\n", gql.FieldName, strings.Join(params, ", "), retType)

	if gql.Auth != nil {
		writeIndent(sb, indent+1)
		fmt.Fprintf(sb, "# Requires %s authentication\n", gql.Auth.AuthType)
	}

	// Inject providers
	for _, prov := range gql.Providers {
		writeIndent(sb, indent+1)
		depFunc := providerToDependsFunc(prov.ProviderType)
		fmt.Fprintf(sb, "%s = %s()\n", prov.Name, depFunc)
	}

	g.writeStatements(sb, gql.Body, indent+1)
}

// writeGRPC generates gRPC service implementations for Python.
// Uses grpcio library. Helper functions used here (irTypeNameForInput, writeIndent,
// providerToDependsFunc, writeStatements) are all defined in this file (python.go).
func (g *PythonGenerator) writeGRPC(sb *strings.Builder, service *ir.ServiceIR) {
	if len(service.GRPC) == 0 {
		return
	}
	sb.WriteString("\n# --- gRPC Services ---\n")
	sb.WriteString("# Generate .proto files and compile with grpc_tools.protoc\n\n")

	for _, svc := range service.GRPC {
		// Write proto stub comment
		fmt.Fprintf(sb, "# Proto definition for %s:\n", svc.Name)
		fmt.Fprintf(sb, "# service %s {\n", svc.Name)
		for _, m := range svc.Methods {
			streamPrefix := ""
			switch m.StreamType {
			case ir.GRPCServerStream:
				streamPrefix = "stream "
			case ir.GRPCClientStream:
				streamPrefix = "// client stream: "
			case ir.GRPCBidirectional:
				streamPrefix = "// bidirectional stream: "
			}
			inputName := irTypeNameForInput(&m.InputType)
			returnName := irTypeNameForInput(&m.ReturnType)
			fmt.Fprintf(sb, "#   rpc %s (%s) returns (%s%s);\n", m.Name, inputName, streamPrefix, returnName)
		}
		sb.WriteString("# }\n\n")

		// Write service implementation class
		fmt.Fprintf(sb, "class %sServicer:\n", svc.Name)
		fmt.Fprintf(sb, "    \"\"\"gRPC service implementation for %s.\"\"\"\n\n", svc.Name)

		hasContent := false
		for _, h := range svc.Handlers {
			hasContent = true
			fmt.Fprintf(sb, "    async def %s(self, request, context):\n", h.MethodName)
			if h.Auth != nil {
				writeIndent(sb, 2)
				fmt.Fprintf(sb, "# Requires %s authentication\n", h.Auth.AuthType)
			}
			for _, prov := range h.Providers {
				writeIndent(sb, 2)
				depFunc := providerToDependsFunc(prov.ProviderType)
				fmt.Fprintf(sb, "%s = %s()\n", prov.Name, depFunc)
			}
			g.writeStatements(sb, h.Body, 2)
			sb.WriteString("\n")
		}

		// Write methods without handlers as stubs
		for _, m := range svc.Methods {
			handlerFound := false
			for _, h := range svc.Handlers {
				if h.MethodName == m.Name {
					handlerFound = true
					break
				}
			}
			if !handlerFound {
				hasContent = true
				fmt.Fprintf(sb, "    async def %s(self, request, context):\n", m.Name)
				sb.WriteString("        raise NotImplementedError\n\n")
			}
		}

		if !hasContent {
			sb.WriteString("    pass\n")
		}
		sb.WriteString("\n")
	}
}

func (g *PythonGenerator) writeWebSockets(sb *strings.Builder, service *ir.ServiceIR) {
	if len(service.WebSocket) == 0 {
		return
	}
	sb.WriteString("\n# --- WebSocket Handlers ---\n")
	sb.WriteString("# Connection manager for WebSocket clients\n")
	sb.WriteString("class ConnectionManager:\n")
	sb.WriteString("    def __init__(self):\n")
	sb.WriteString("        self.active_connections: list[WebSocket] = []\n\n")
	sb.WriteString("    async def connect(self, websocket: WebSocket):\n")
	sb.WriteString("        await websocket.accept()\n")
	sb.WriteString("        self.active_connections.append(websocket)\n\n")
	sb.WriteString("    def disconnect(self, websocket: WebSocket):\n")
	sb.WriteString("        self.active_connections.remove(websocket)\n\n")
	sb.WriteString("    async def broadcast(self, message: str):\n")
	sb.WriteString("        for connection in self.active_connections:\n")
	sb.WriteString("            await connection.send_text(message)\n\n")
	sb.WriteString("manager = ConnectionManager()\n\n")

	for _, ws := range service.WebSocket {
		fastapiPath := glyphPathToFastAPI(ws.Path)
		fmt.Fprintf(sb, "\n@app.websocket(\"%s\")\n", fastapiPath)
		sb.WriteString("async def websocket_endpoint(websocket: WebSocket):\n")

		// Find connect handler
		hasConnect := false
		for _, ev := range ws.Events {
			if ev.EventType == ir.WSConnect {
				hasConnect = true
				break
			}
		}
		if hasConnect {
			for _, ev := range ws.Events {
				if ev.EventType == ir.WSConnect {
					sb.WriteString("    await manager.connect(websocket)\n")
					g.writeStatements(sb, ev.Body, 1)
					break
				}
			}
		} else {
			sb.WriteString("    await manager.connect(websocket)\n")
		}

		sb.WriteString("    try:\n")
		sb.WriteString("        while True:\n")
		sb.WriteString("            data = await websocket.receive_text()\n")

		// Message handler
		for _, ev := range ws.Events {
			if ev.EventType == ir.WSMessage {
				g.writeStatements(sb, ev.Body, 3)
				break
			}
		}

		sb.WriteString("    except WebSocketDisconnect:\n")
		// Disconnect handler
		hasDisconnect := false
		for _, ev := range ws.Events {
			if ev.EventType == ir.WSDisconnect {
				hasDisconnect = true
				sb.WriteString("        manager.disconnect(websocket)\n")
				g.writeStatements(sb, ev.Body, 2)
				break
			}
		}
		if !hasDisconnect {
			sb.WriteString("        manager.disconnect(websocket)\n")
		}

		// Error handler
		for _, ev := range ws.Events {
			if ev.EventType == ir.WSError {
				sb.WriteString("    except Exception as e:\n")
				g.writeStatements(sb, ev.Body, 2)
				sb.WriteString("        manager.disconnect(websocket)\n")
				break
			}
		}
		sb.WriteString("\n")
	}
}

func (g *PythonGenerator) writeMatchExpr(sb *strings.Builder, match *ir.MatchExpr) {
	// Generate as if/elif chain for broad compatibility
	for i, c := range match.Cases {
		if c.Pattern.Kind == ir.PatternWildcard {
			// Wildcard is the default/else case
			continue
		}
		if i == 0 {
			sb.WriteString("(")
		} else {
			sb.WriteString(" if ")
		}
		g.writeExpr(sb, c.Body)
	}
	// Find wildcard case for the else branch
	hasWildcard := false
	for _, c := range match.Cases {
		if c.Pattern.Kind == ir.PatternWildcard {
			sb.WriteString(" if True else ")
			g.writeExpr(sb, c.Body)
			hasWildcard = true
			break
		}
	}
	if !hasWildcard && len(match.Cases) > 0 {
		sb.WriteString(" if True else None")
	}
	sb.WriteString(")")
}

func (g *PythonGenerator) writeMatchStatements(sb *strings.Builder, match *ir.MatchExpr, indent int) {
	for i, c := range match.Cases {
		writeIndent(sb, indent)
		if c.Pattern.Kind == ir.PatternWildcard {
			sb.WriteString("else:\n")
			writeIndent(sb, indent+1)
			g.writeExpr(sb, c.Body)
			sb.WriteString("\n")
			continue
		}
		if i == 0 {
			sb.WriteString("if ")
		} else {
			sb.WriteString("elif ")
		}
		g.writeExpr(sb, match.Value)
		sb.WriteString(" == ")
		g.writePattern(sb, c.Pattern)
		if c.Guard != nil {
			sb.WriteString(" and ")
			g.writeExpr(sb, *c.Guard)
		}
		sb.WriteString(":\n")
		writeIndent(sb, indent+1)
		g.writeExpr(sb, c.Body)
		sb.WriteString("\n")
	}
}

func (g *PythonGenerator) writePattern(sb *strings.Builder, pat ir.PatternIR) {
	switch pat.Kind {
	case ir.PatternLiteral:
		if pat.StrVal != "" {
			fmt.Fprintf(sb, "%q", pat.StrVal)
		} else if pat.BoolVal {
			sb.WriteString("True")
		} else if pat.IntVal != 0 {
			fmt.Fprintf(sb, "%d", pat.IntVal)
		} else if pat.FloatVal != 0 {
			fmt.Fprintf(sb, "%g", pat.FloatVal)
		} else {
			sb.WriteString("0")
		}
	case ir.PatternVariable:
		sb.WriteString(pat.VarName)
	case ir.PatternWildcard:
		sb.WriteString("_")
	case ir.PatternObject:
		sb.WriteString("{")
		for i, f := range pat.Fields {
			if i > 0 {
				sb.WriteString(", ")
			}
			fmt.Fprintf(sb, "%q: ", f.Key)
			g.writePattern(sb, f.Pattern)
		}
		sb.WriteString("}")
	case ir.PatternArray:
		sb.WriteString("[")
		for i, el := range pat.Elements {
			if i > 0 {
				sb.WriteString(", ")
			}
			g.writePattern(sb, el)
		}
		if pat.RestVar != "" {
			if len(pat.Elements) > 0 {
				sb.WriteString(", ")
			}
			fmt.Fprintf(sb, "*%s", pat.RestVar)
		}
		sb.WriteString("]")
	}
}

func (g *PythonGenerator) writeCronJobs(sb *strings.Builder, service *ir.ServiceIR) {
	if len(service.CronJobs) == 0 {
		return
	}
	sb.WriteString("\n# --- Cron Jobs ---\n")
	sb.WriteString("scheduler = AsyncIOScheduler()\n\n")
	for _, cron := range service.CronJobs {
		name := cron.Name
		if name == "" {
			name = "cron_task"
		}
		fmt.Fprintf(sb, "\n@scheduler.scheduled_job(CronTrigger.from_crontab(\"%s\"))\n", cron.Schedule)
		fmt.Fprintf(sb, "async def %s():\n", name)
		// Inject providers
		for _, prov := range cron.Providers {
			writeIndent(sb, 1)
			depFunc := providerToDependsFunc(prov.ProviderType)
			fmt.Fprintf(sb, "%s = %s()\n", prov.Name, depFunc)
		}
		g.writeStatements(sb, cron.Body, 1)
		sb.WriteString("\n")
	}
}

func (g *PythonGenerator) writeEventHandlers(sb *strings.Builder, service *ir.ServiceIR) {
	if len(service.Events) == 0 {
		return
	}
	sb.WriteString("\n# --- Event Handlers ---\n")
	sb.WriteString("# Implement event dispatch with your preferred event system\n\n")
	for _, ev := range service.Events {
		funcName := "handle_" + strings.ReplaceAll(ev.EventType, ".", "_")
		fmt.Fprintf(sb, "\nasync def %s(event):\n", funcName)
		for _, prov := range ev.Providers {
			writeIndent(sb, 1)
			depFunc := providerToDependsFunc(prov.ProviderType)
			fmt.Fprintf(sb, "%s = %s()\n", prov.Name, depFunc)
		}
		g.writeStatements(sb, ev.Body, 1)
		sb.WriteString("\n")
	}
}

func (g *PythonGenerator) writeQueueWorkers(sb *strings.Builder, service *ir.ServiceIR) {
	if len(service.Queues) == 0 {
		return
	}
	sb.WriteString("\n# --- Queue Workers ---\n")
	sb.WriteString("# Implement with Celery, RQ, or your preferred task queue\n\n")
	for _, q := range service.Queues {
		funcName := "worker_" + strings.ReplaceAll(q.QueueName, ".", "_")
		fmt.Fprintf(sb, "\nasync def %s(message):\n", funcName)
		for _, prov := range q.Providers {
			writeIndent(sb, 1)
			depFunc := providerToDependsFunc(prov.ProviderType)
			fmt.Fprintf(sb, "%s = %s()\n", prov.Name, depFunc)
		}
		g.writeStatements(sb, q.Body, 1)
		sb.WriteString("\n")
	}
}

func (g *PythonGenerator) writeMain(sb *strings.Builder) {
	sb.WriteString("\nif __name__ == \"__main__\":\n")
	sb.WriteString("    import uvicorn\n")
	fmt.Fprintf(sb, "    uvicorn.run(app, host=%q, port=%d)\n", g.host, g.port)
}

// --- Helper functions ---

func irTypeToPython(t ir.TypeRef) string {
	switch t.Kind {
	case ir.TypeInt:
		return "int"
	case ir.TypeFloat:
		return "float"
	case ir.TypeString:
		return "str"
	case ir.TypeBool:
		return "bool"
	case ir.TypeArray:
		if t.Inner != nil {
			return fmt.Sprintf("List[%s]", irTypeToPython(*t.Inner))
		}
		return "List[Any]"
	case ir.TypeOptional:
		if t.Inner != nil {
			return fmt.Sprintf("Optional[%s]", irTypeToPython(*t.Inner))
		}
		return "Optional[Any]"
	case ir.TypeNamed:
		return t.Name
	case ir.TypeProvider:
		return t.Name
	case ir.TypeUnion:
		if len(t.Elements) > 0 {
			var parts []string
			for _, elem := range t.Elements {
				parts = append(parts, irTypeToPython(elem))
			}
			return strings.Join(parts, " | ")
		}
		return "Any"
	case ir.TypeAny:
		return "Any"
	default:
		return "Any"
	}
}

func glyphPathToFastAPI(path string) string {
	var parts []string
	for _, seg := range strings.Split(path, "/") {
		if strings.HasPrefix(seg, ":") {
			parts = append(parts, fmt.Sprintf("{%s}", seg[1:]))
		} else {
			parts = append(parts, seg)
		}
	}
	return strings.Join(parts, "/")
}

func providerToDependsFunc(providerType string) string {
	switch providerType {
	case "Database":
		return "get_db"
	case "Redis":
		return "get_redis"
	default:
		return "get_" + strings.ToLower(providerType)
	}
}

func irTypeNameForInput(t *ir.TypeRef) string {
	if t == nil {
		return "dict"
	}
	if t.Kind == ir.TypeNamed {
		return t.Name
	}
	return "dict"
}

func routeToFuncName(route ir.RouteHandler) string {
	method := strings.ToLower(route.Method.String())
	segments := strings.Split(route.Path, "/")
	var nameParts []string
	for _, seg := range segments {
		if seg == "" {
			continue
		}
		if strings.HasPrefix(seg, ":") {
			nameParts = append(nameParts, seg[1:])
		} else if strings.HasPrefix(seg, "{") {
			nameParts = append(nameParts, strings.Trim(seg, "{}"))
		} else {
			nameParts = append(nameParts, seg)
		}
	}
	if len(nameParts) == 0 {
		return method + "_root"
	}
	return method + "_" + strings.Join(nameParts, "_")
}

// isObjectReceiver returns true if the expression looks like a method receiver
// (a variable or field access chain, not a literal or complex expression).
func isObjectReceiver(expr ir.ExprIR) bool {
	return expr.Kind == ir.ExprFieldAccess || expr.Kind == ir.ExprVar
}

func binOpToPython(op ir.BinOp) string {
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
		return "=="
	case ir.OpNe:
		return "!="
	case ir.OpLt:
		return "<"
	case ir.OpLe:
		return "<="
	case ir.OpGt:
		return ">"
	case ir.OpGe:
		return ">="
	case ir.OpAnd:
		return "and"
	case ir.OpOr:
		return "or"
	default:
		return "+"
	}
}

func unaryOpToPython(op ir.UnOp) string {
	switch op {
	case ir.OpNot:
		return "not "
	case ir.OpNeg:
		return "-"
	default:
		return ""
	}
}

func writeIndent(sb *strings.Builder, level int) {
	for i := 0; i < level; i++ {
		sb.WriteString("    ")
	}
}
