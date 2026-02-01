// Package formatter provides bidirectional conversion between compact glyph syntax
// and expanded human-readable syntax.
//
// Compact (Glyph) syntax uses symbols:
//
//	: -> type definition
//	@ -> route
//	$ -> variable assignment (let)
//	> -> return
//	+ -> middleware
//	% -> inject
//	~ -> event handler
//	* -> cron task
//	! -> command
//	& -> queue worker
//
// Expanded syntax uses keywords:
//
//	type, route, let, return, middleware, inject, event, cron, command, queue
package formatter

import (
	"fmt"
	"github.com/glyphlang/glyph/pkg/ast"
	"strings"
)

// Mode determines the output format
type Mode int

const (
	// Compact uses glyph symbols (@, $, >, etc.)
	Compact Mode = iota
	// Expanded uses keywords (route, let, return, etc.)
	Expanded
)

// Formatter converts AST to source code in either compact or expanded mode
type Formatter struct {
	mode   Mode
	indent int
	output strings.Builder
}

// New creates a new Formatter with the specified mode
func New(mode Mode) *Formatter {
	return &Formatter{
		mode:   mode,
		indent: 0,
	}
}

// Format formats a module to source code
func (f *Formatter) Format(module *ast.Module) string {
	f.output.Reset()

	for i, item := range module.Items {
		if i > 0 {
			f.writeln("")
		}
		f.formatItem(item)
	}

	return f.output.String()
}

func (f *Formatter) formatItem(item ast.Item) {
	switch v := item.(type) {
	case *ast.TypeDef:
		f.formatTypeDef(v)
	case *ast.Route:
		f.formatRoute(v)
	case *ast.Command:
		f.formatCommand(v)
	case *ast.CronTask:
		f.formatCronTask(v)
	case *ast.EventHandler:
		f.formatEventHandler(v)
	case *ast.QueueWorker:
		f.formatQueueWorker(v)
	case *ast.Function:
		f.formatFunction(v)
	case *ast.WebSocketRoute:
		f.formatWebSocketRoute(v)
	case *ast.ImportStatement:
		f.formatImport(v)
	case *ast.ModuleDecl:
		f.formatModuleDecl(v)
	case *ast.MacroDef:
		f.formatMacroDef(v)
	}
}

func (f *Formatter) formatTypeDef(td *ast.TypeDef) {
	if f.mode == Expanded {
		f.write("type ")
	} else {
		f.write(": ")
	}

	f.write(td.Name)

	if len(td.TypeParams) > 0 {
		f.write("<")
		for i, tp := range td.TypeParams {
			if i > 0 {
				f.write(", ")
			}
			f.write(tp.Name)
			if tp.Constraint != nil {
				f.write(": ")
				f.formatType(tp.Constraint)
			}
		}
		f.write(">")
	}

	f.writeln(" {")
	f.indent++

	for _, field := range td.Fields {
		f.writeIndent()
		f.write(field.Name)
		f.write(": ")
		f.formatType(field.TypeAnnotation)
		if field.Required {
			f.write("!")
		}
		f.writeln("")
	}

	f.indent--
	f.writeIndent()
	f.writeln("}")
}

func (f *Formatter) formatRoute(r *ast.Route) {
	if f.mode == Expanded {
		f.write("route ")
	} else {
		f.write("@ ")
	}

	f.write(r.Method.String())
	f.write(" ")
	f.write(r.Path)

	for _, qp := range r.QueryParams {
		f.write(" ?")
		f.write(qp.Name)
		if qp.Type != nil {
			f.write(": ")
			f.formatType(qp.Type)
		}
		if qp.Required {
			f.write("!")
		}
		if qp.Default != nil {
			f.write(" = ")
			f.formatExpr(qp.Default)
		}
	}

	if r.ReturnType != nil {
		f.write(" -> ")
		f.formatType(r.ReturnType)
	}

	f.writeln(" {")
	f.indent++

	if r.Auth != nil {
		f.writeIndent()
		if f.mode == Expanded {
			f.write("middleware ")
		} else {
			f.write("+ ")
		}
		f.write("auth(")
		f.write(r.Auth.AuthType)
		f.writeln(")")
	}

	if r.RateLimit != nil {
		f.writeIndent()
		if f.mode == Expanded {
			f.write("middleware ")
		} else {
			f.write("+ ")
		}
		f.write("ratelimit(")
		f.write(fmt.Sprintf("%d/%s", r.RateLimit.Requests, r.RateLimit.Window))
		f.writeln(")")
	}

	for _, inj := range r.Injections {
		f.writeIndent()
		if f.mode == Expanded {
			f.write("use ")
		} else {
			f.write("% ")
		}
		f.write(inj.Name)
		f.write(": ")
		f.formatType(inj.Type)
		f.writeln("")
	}

	for _, stmt := range r.Body {
		f.formatStatement(stmt)
	}

	f.indent--
	f.writeIndent()
	f.writeln("}")
}

func (f *Formatter) formatCommand(c *ast.Command) {
	if f.mode == Expanded {
		f.write("command ")
	} else {
		f.write("! ")
	}

	f.write(c.Name)

	if c.Description != "" {
		f.write(" \"")
		f.write(c.Description)
		f.write("\"")
	}

	for _, p := range c.Params {
		f.write(" ")
		if p.IsFlag {
			f.write("--")
		}
		f.write(p.Name)
		if p.Type != nil {
			f.write(": ")
			f.formatType(p.Type)
		}
		if p.Required {
			f.write("!")
		}
		if p.Default != nil {
			f.write(" = ")
			f.formatExpr(p.Default)
		}
	}

	f.writeln(" {")
	f.indent++

	for _, stmt := range c.Body {
		f.formatStatement(stmt)
	}

	f.indent--
	f.writeIndent()
	f.writeln("}")
}

func (f *Formatter) formatCronTask(ct *ast.CronTask) {
	if f.mode == Expanded {
		f.write("cron ")
	} else {
		f.write("* ")
	}

	f.write("\"")
	f.write(ct.Schedule)
	f.write("\"")

	if ct.Name != "" {
		f.write(" ")
		f.write(ct.Name)
	}

	f.writeln(" {")
	f.indent++

	for _, inj := range ct.Injections {
		f.writeIndent()
		if f.mode == Expanded {
			f.write("use ")
		} else {
			f.write("% ")
		}
		f.write(inj.Name)
		f.write(": ")
		f.formatType(inj.Type)
		f.writeln("")
	}

	for _, stmt := range ct.Body {
		f.formatStatement(stmt)
	}

	f.indent--
	f.writeIndent()
	f.writeln("}")
}

func (f *Formatter) formatEventHandler(eh *ast.EventHandler) {
	if f.mode == Expanded {
		f.write("handle ")
	} else {
		f.write("~ ")
	}

	f.write("\"")
	f.write(eh.EventType)
	f.write("\"")

	if eh.Async {
		f.write(" async")
	}

	f.writeln(" {")
	f.indent++

	for _, inj := range eh.Injections {
		f.writeIndent()
		if f.mode == Expanded {
			f.write("use ")
		} else {
			f.write("% ")
		}
		f.write(inj.Name)
		f.write(": ")
		f.formatType(inj.Type)
		f.writeln("")
	}

	for _, stmt := range eh.Body {
		f.formatStatement(stmt)
	}

	f.indent--
	f.writeIndent()
	f.writeln("}")
}

func (f *Formatter) formatQueueWorker(qw *ast.QueueWorker) {
	if f.mode == Expanded {
		f.write("queue ")
	} else {
		f.write("& ")
	}

	f.write("\"")
	f.write(qw.QueueName)
	f.write("\"")

	f.writeln(" {")
	f.indent++

	if qw.Concurrency > 0 {
		f.writeIndent()
		if f.mode == Expanded {
			f.write("middleware ")
		} else {
			f.write("+ ")
		}
		f.write(fmt.Sprintf("concurrency(%d)", qw.Concurrency))
		f.writeln("")
	}

	if qw.MaxRetries > 0 {
		f.writeIndent()
		if f.mode == Expanded {
			f.write("middleware ")
		} else {
			f.write("+ ")
		}
		f.write(fmt.Sprintf("retries(%d)", qw.MaxRetries))
		f.writeln("")
	}

	if qw.Timeout > 0 {
		f.writeIndent()
		if f.mode == Expanded {
			f.write("middleware ")
		} else {
			f.write("+ ")
		}
		f.write(fmt.Sprintf("timeout(%d)", qw.Timeout))
		f.writeln("")
	}

	for _, inj := range qw.Injections {
		f.writeIndent()
		if f.mode == Expanded {
			f.write("use ")
		} else {
			f.write("% ")
		}
		f.write(inj.Name)
		f.write(": ")
		f.formatType(inj.Type)
		f.writeln("")
	}

	for _, stmt := range qw.Body {
		f.formatStatement(stmt)
	}

	f.indent--
	f.writeIndent()
	f.writeln("}")
}

func (f *Formatter) formatFunction(fn *ast.Function) {
	if f.mode == Expanded {
		f.write("func ")
	} else {
		f.write("= ")
	}

	f.write(fn.Name)

	if len(fn.TypeParams) > 0 {
		f.write("<")
		for i, tp := range fn.TypeParams {
			if i > 0 {
				f.write(", ")
			}
			f.write(tp.Name)
		}
		f.write(">")
	}

	f.write("(")
	for i, p := range fn.Params {
		if i > 0 {
			f.write(", ")
		}
		f.write(p.Name)
		if p.TypeAnnotation != nil {
			f.write(": ")
			f.formatType(p.TypeAnnotation)
		}
	}
	f.write(")")

	if fn.ReturnType != nil {
		f.write(" -> ")
		f.formatType(fn.ReturnType)
	}

	f.writeln(" {")
	f.indent++

	for _, stmt := range fn.Body {
		f.formatStatement(stmt)
	}

	f.indent--
	f.writeIndent()
	f.writeln("}")
}

func (f *Formatter) formatWebSocketRoute(ws *ast.WebSocketRoute) {
	if f.mode == Expanded {
		f.write("route ")
	} else {
		f.write("@ ")
	}
	f.write("WS ")
	f.write(ws.Path)
	f.writeln(" {")
	f.indent++

	for _, event := range ws.Events {
		f.writeIndent()
		switch event.EventType {
		case ast.WSEventConnect:
			f.writeln("on connect {")
		case ast.WSEventDisconnect:
			f.writeln("on disconnect {")
		case ast.WSEventMessage:
			f.writeln("on message {")
		case ast.WSEventError:
			f.writeln("on error {")
		}
		f.indent++
		for _, stmt := range event.Body {
			f.formatStatement(stmt)
		}
		f.indent--
		f.writeIndent()
		f.writeln("}")
	}

	f.indent--
	f.writeIndent()
	f.writeln("}")
}

func (f *Formatter) formatImport(imp *ast.ImportStatement) {
	if imp.Selective {
		f.write("from \"")
		f.write(imp.Path)
		f.write("\" import { ")
		for i, name := range imp.Names {
			if i > 0 {
				f.write(", ")
			}
			f.write(name.Name)
			if name.Alias != "" {
				f.write(" as ")
				f.write(name.Alias)
			}
		}
		f.writeln(" }")
	} else {
		f.write("import \"")
		f.write(imp.Path)
		f.write("\"")
		if imp.Alias != "" {
			f.write(" as ")
			f.write(imp.Alias)
		}
		f.writeln("")
	}
}

func (f *Formatter) formatModuleDecl(m *ast.ModuleDecl) {
	f.write("module \"")
	f.write(m.Name)
	f.writeln("\"")
}

func (f *Formatter) formatMacroDef(m *ast.MacroDef) {
	f.write("macro! ")
	f.write(m.Name)
	f.write("(")
	for i, p := range m.Params {
		if i > 0 {
			f.write(", ")
		}
		f.write(p)
	}
	f.writeln(") {")
	f.indent++
	f.writeIndent()
	f.writeln("# ... macro body ...")
	f.indent--
	f.writeIndent()
	f.writeln("}")
}

func (f *Formatter) formatStatement(stmt ast.Statement) {
	f.writeIndent()

	switch v := stmt.(type) {
	// Handle both pointer and value types for each statement
	case ast.AssignStatement:
		f.formatAssign(v.Target, v.Value)
	case *ast.AssignStatement:
		f.formatAssign(v.Target, v.Value)

	case ast.ReassignStatement:
		f.formatReassign(v.Target, v.Value)
	case *ast.ReassignStatement:
		f.formatReassign(v.Target, v.Value)

	case ast.ReturnStatement:
		f.formatReturn(v.Value)
	case *ast.ReturnStatement:
		f.formatReturn(v.Value)

	case ast.IfStatement:
		f.formatIf(v.Condition, v.ThenBlock, v.ElseBlock)
	case *ast.IfStatement:
		f.formatIf(v.Condition, v.ThenBlock, v.ElseBlock)

	case ast.WhileStatement:
		f.formatWhile(v.Condition, v.Body)
	case *ast.WhileStatement:
		f.formatWhile(v.Condition, v.Body)

	case ast.ForStatement:
		f.formatFor(v.KeyVar, v.ValueVar, v.Iterable, v.Body)
	case *ast.ForStatement:
		f.formatFor(v.KeyVar, v.ValueVar, v.Iterable, v.Body)

	case ast.SwitchStatement:
		f.formatSwitch(v.Value, v.Cases, v.Default)
	case *ast.SwitchStatement:
		f.formatSwitch(v.Value, v.Cases, v.Default)

	case ast.ExpressionStatement:
		f.formatExpr(v.Expr)
		f.writeln("")
	case *ast.ExpressionStatement:
		f.formatExpr(v.Expr)
		f.writeln("")

	case ast.ValidationStatement:
		if f.mode == Expanded {
			f.write("validate ")
		} else {
			f.write("? ")
		}
		f.formatFunctionCall(v.Call)
		f.writeln("")
	case *ast.ValidationStatement:
		if f.mode == Expanded {
			f.write("validate ")
		} else {
			f.write("? ")
		}
		f.formatFunctionCall(v.Call)
		f.writeln("")

	case ast.DbQueryStatement:
		f.formatDbQuery(v.Var, v.Query, v.Params)
	case *ast.DbQueryStatement:
		f.formatDbQuery(v.Var, v.Query, v.Params)

	case ast.WsSendStatement:
		f.write("ws.send(")
		f.formatExpr(v.Client)
		f.write(", ")
		f.formatExpr(v.Message)
		f.writeln(")")
	case *ast.WsSendStatement:
		f.write("ws.send(")
		f.formatExpr(v.Client)
		f.write(", ")
		f.formatExpr(v.Message)
		f.writeln(")")

	case ast.WsBroadcastStatement:
		f.write("ws.broadcast(")
		f.formatExpr(v.Message)
		if v.Except != nil {
			f.write(", except: ")
			f.formatExpr(*v.Except)
		}
		f.writeln(")")
	case *ast.WsBroadcastStatement:
		f.write("ws.broadcast(")
		f.formatExpr(v.Message)
		if v.Except != nil {
			f.write(", except: ")
			f.formatExpr(*v.Except)
		}
		f.writeln(")")

	case ast.WsCloseStatement:
		f.write("ws.close(")
		f.formatExpr(v.Client)
		if v.Reason != nil {
			f.write(", ")
			f.formatExpr(v.Reason)
		}
		f.writeln(")")
	case *ast.WsCloseStatement:
		f.write("ws.close(")
		f.formatExpr(v.Client)
		if v.Reason != nil {
			f.write(", ")
			f.formatExpr(v.Reason)
		}
		f.writeln(")")
	}
}

// Statement formatting helpers

func (f *Formatter) formatAssign(target string, value ast.Expr) {
	if f.mode == Expanded {
		f.write("let ")
	} else {
		f.write("$ ")
	}
	f.write(target)
	f.write(" = ")
	f.formatExpr(value)
	f.writeln("")
}

func (f *Formatter) formatReassign(target string, value ast.Expr) {
	f.write(target)
	f.write(" = ")
	f.formatExpr(value)
	f.writeln("")
}

func (f *Formatter) formatReturn(value ast.Expr) {
	if f.mode == Expanded {
		f.write("return ")
	} else {
		f.write("> ")
	}
	f.formatExpr(value)
	f.writeln("")
}

func (f *Formatter) formatIf(condition ast.Expr, thenBlock, elseBlock []ast.Statement) {
	f.write("if ")
	f.formatExpr(condition)
	f.writeln(" {")
	f.indent++
	for _, s := range thenBlock {
		f.formatStatement(s)
	}
	f.indent--
	f.writeIndent()
	if len(elseBlock) > 0 {
		f.writeln("} else {")
		f.indent++
		for _, s := range elseBlock {
			f.formatStatement(s)
		}
		f.indent--
		f.writeIndent()
	}
	f.writeln("}")
}

func (f *Formatter) formatWhile(condition ast.Expr, body []ast.Statement) {
	f.write("while ")
	f.formatExpr(condition)
	f.writeln(" {")
	f.indent++
	for _, s := range body {
		f.formatStatement(s)
	}
	f.indent--
	f.writeIndent()
	f.writeln("}")
}

func (f *Formatter) formatFor(keyVar, valueVar string, iterable ast.Expr, body []ast.Statement) {
	f.write("for ")
	if keyVar != "" {
		f.write(keyVar)
		f.write(", ")
	}
	f.write(valueVar)
	f.write(" in ")
	f.formatExpr(iterable)
	f.writeln(" {")
	f.indent++
	for _, s := range body {
		f.formatStatement(s)
	}
	f.indent--
	f.writeIndent()
	f.writeln("}")
}

func (f *Formatter) formatSwitch(value ast.Expr, cases []ast.SwitchCase, defaultBlock []ast.Statement) {
	f.write("switch ")
	f.formatExpr(value)
	f.writeln(" {")
	f.indent++
	for _, c := range cases {
		f.writeIndent()
		f.write("case ")
		f.formatExpr(c.Value)
		f.writeln(" {")
		f.indent++
		for _, s := range c.Body {
			f.formatStatement(s)
		}
		f.indent--
		f.writeIndent()
		f.writeln("}")
	}
	if len(defaultBlock) > 0 {
		f.writeIndent()
		f.writeln("default {")
		f.indent++
		for _, s := range defaultBlock {
			f.formatStatement(s)
		}
		f.indent--
		f.writeIndent()
		f.writeln("}")
	}
	f.indent--
	f.writeIndent()
	f.writeln("}")
}

func (f *Formatter) formatDbQuery(varName, query string, params []ast.Expr) {
	if f.mode == Expanded {
		f.write("let ")
	} else {
		f.write("$ ")
	}
	f.write(varName)
	f.write(" = db.query(\"")
	f.write(query)
	f.write("\"")
	for _, p := range params {
		f.write(", ")
		f.formatExpr(p)
	}
	f.writeln(")")
}

func (f *Formatter) formatFunctionCall(call ast.FunctionCallExpr) {
	f.write(call.Name)
	if len(call.TypeArgs) > 0 {
		f.write("<")
		for i, t := range call.TypeArgs {
			if i > 0 {
				f.write(", ")
			}
			f.formatType(t)
		}
		f.write(">")
	}
	f.write("(")
	for i, arg := range call.Args {
		if i > 0 {
			f.write(", ")
		}
		f.formatExpr(arg)
	}
	f.write(")")
}

func (f *Formatter) formatExpr(expr ast.Expr) {
	switch v := expr.(type) {
	case ast.LiteralExpr:
		f.formatLiteral(v.Value)
	case *ast.LiteralExpr:
		f.formatLiteral(v.Value)

	case ast.VariableExpr:
		f.write(v.Name)
	case *ast.VariableExpr:
		f.write(v.Name)

	case ast.BinaryOpExpr:
		f.formatBinaryOp(v.Op, v.Left, v.Right)
	case *ast.BinaryOpExpr:
		f.formatBinaryOp(v.Op, v.Left, v.Right)

	case ast.UnaryOpExpr:
		f.write(v.Op.String())
		f.formatExpr(v.Right)
	case *ast.UnaryOpExpr:
		f.write(v.Op.String())
		f.formatExpr(v.Right)

	case ast.FieldAccessExpr:
		f.formatExpr(v.Object)
		f.write(".")
		f.write(v.Field)
	case *ast.FieldAccessExpr:
		f.formatExpr(v.Object)
		f.write(".")
		f.write(v.Field)

	case ast.ArrayIndexExpr:
		f.formatExpr(v.Array)
		f.write("[")
		f.formatExpr(v.Index)
		f.write("]")
	case *ast.ArrayIndexExpr:
		f.formatExpr(v.Array)
		f.write("[")
		f.formatExpr(v.Index)
		f.write("]")

	case ast.FunctionCallExpr:
		f.formatFunctionCall(v)
	case *ast.FunctionCallExpr:
		f.formatFunctionCall(*v)

	case ast.ObjectExpr:
		f.formatObject(v.Fields)
	case *ast.ObjectExpr:
		f.formatObject(v.Fields)

	case ast.ArrayExpr:
		f.formatArray(v.Elements)
	case *ast.ArrayExpr:
		f.formatArray(v.Elements)

	case ast.LambdaExpr:
		f.formatLambda(v.Params, v.Body, v.Block)
	case *ast.LambdaExpr:
		f.formatLambda(v.Params, v.Body, v.Block)

	case ast.MatchExpr:
		f.formatMatch(v.Value, v.Cases)
	case *ast.MatchExpr:
		f.formatMatch(v.Value, v.Cases)

	case ast.AsyncExpr:
		f.formatAsync(v.Body)
	case *ast.AsyncExpr:
		f.formatAsync(v.Body)

	case ast.AwaitExpr:
		f.write("await ")
		f.formatExpr(v.Expr)
	case *ast.AwaitExpr:
		f.write("await ")
		f.formatExpr(v.Expr)
	}
}

func (f *Formatter) formatBinaryOp(op ast.BinOp, left, right ast.Expr) {
	f.formatExpr(left)
	f.write(" ")
	f.write(op.String())
	f.write(" ")
	f.formatExpr(right)
}

func (f *Formatter) formatObject(fields []ast.ObjectField) {
	if len(fields) == 0 {
		f.write("{}")
		return
	}
	if len(fields) <= 3 {
		f.write("{")
		for i, field := range fields {
			if i > 0 {
				f.write(", ")
			}
			f.write(field.Key)
			f.write(": ")
			f.formatExpr(field.Value)
		}
		f.write("}")
	} else {
		f.writeln("{")
		f.indent++
		for i, field := range fields {
			f.writeIndent()
			f.write(field.Key)
			f.write(": ")
			f.formatExpr(field.Value)
			if i < len(fields)-1 {
				f.write(",")
			}
			f.writeln("")
		}
		f.indent--
		f.writeIndent()
		f.write("}")
	}
}

func (f *Formatter) formatArray(elements []ast.Expr) {
	if len(elements) == 0 {
		f.write("[]")
		return
	}
	f.write("[")
	for i, elem := range elements {
		if i > 0 {
			f.write(", ")
		}
		f.formatExpr(elem)
	}
	f.write("]")
}

func (f *Formatter) formatLambda(params []ast.Field, body ast.Expr, block []ast.Statement) {
	f.write("(")
	for i, p := range params {
		if i > 0 {
			f.write(", ")
		}
		f.write(p.Name)
		if p.TypeAnnotation != nil {
			f.write(": ")
			f.formatType(p.TypeAnnotation)
		}
	}
	f.write(") => ")
	if body != nil {
		f.formatExpr(body)
	} else if len(block) > 0 {
		f.writeln("{")
		f.indent++
		for _, s := range block {
			f.formatStatement(s)
		}
		f.indent--
		f.writeIndent()
		f.write("}")
	}
}

func (f *Formatter) formatMatch(value ast.Expr, cases []ast.MatchCase) {
	f.write("match ")
	f.formatExpr(value)
	f.writeln(" {")
	f.indent++
	for _, c := range cases {
		f.writeIndent()
		f.formatPattern(c.Pattern)
		if c.Guard != nil {
			f.write(" when ")
			f.formatExpr(c.Guard)
		}
		f.write(" => ")
		f.formatExpr(c.Body)
		f.writeln("")
	}
	f.indent--
	f.writeIndent()
	f.write("}")
}

func (f *Formatter) formatAsync(body []ast.Statement) {
	f.writeln("async {")
	f.indent++
	for _, s := range body {
		f.formatStatement(s)
	}
	f.indent--
	f.writeIndent()
	f.write("}")
}

func (f *Formatter) formatLiteral(lit ast.Literal) {
	switch v := lit.(type) {
	case ast.IntLiteral:
		f.write(fmt.Sprintf("%d", v.Value))
	case ast.FloatLiteral:
		f.write(fmt.Sprintf("%g", v.Value))
	case ast.StringLiteral:
		f.write("\"")
		f.write(escapeString(v.Value))
		f.write("\"")
	case ast.BoolLiteral:
		if v.Value {
			f.write("true")
		} else {
			f.write("false")
		}
	case ast.NullLiteral:
		f.write("null")
	}
}

func (f *Formatter) formatType(t ast.Type) {
	switch v := t.(type) {
	case ast.IntType:
		f.write("int")
	case ast.StringType:
		f.write("str")
	case ast.BoolType:
		f.write("bool")
	case ast.FloatType:
		f.write("float")
	case ast.DatabaseType:
		f.write("Database")
	case ast.NamedType:
		f.write(v.Name)
	case ast.ArrayType:
		f.formatType(v.ElementType)
		f.write("[]")
	case ast.OptionalType:
		f.formatType(v.InnerType)
		f.write("?")
	case ast.GenericType:
		f.formatType(v.BaseType)
		f.write("<")
		for i, arg := range v.TypeArgs {
			if i > 0 {
				f.write(", ")
			}
			f.formatType(arg)
		}
		f.write(">")
	case ast.TypeParameterType:
		f.write(v.Name)
	case ast.FunctionType:
		f.write("(")
		for i, pt := range v.ParamTypes {
			if i > 0 {
				f.write(", ")
			}
			f.formatType(pt)
		}
		f.write(") -> ")
		f.formatType(v.ReturnType)
	case ast.UnionType:
		for i, ut := range v.Types {
			if i > 0 {
				f.write(" | ")
			}
			f.formatType(ut)
		}
	case ast.FutureType:
		f.write("Future<")
		f.formatType(v.ResultType)
		f.write(">")
	}
}

func (f *Formatter) formatPattern(p ast.Pattern) {
	switch v := p.(type) {
	case ast.LiteralPattern:
		f.formatLiteral(v.Value)
	case ast.VariablePattern:
		f.write(v.Name)
	case ast.WildcardPattern:
		f.write("_")
	case ast.ObjectPattern:
		f.write("{")
		for i, field := range v.Fields {
			if i > 0 {
				f.write(", ")
			}
			f.write(field.Key)
			if field.Pattern != nil {
				f.write(": ")
				f.formatPattern(field.Pattern)
			}
		}
		f.write("}")
	case ast.ArrayPattern:
		f.write("[")
		for i, elem := range v.Elements {
			if i > 0 {
				f.write(", ")
			}
			f.formatPattern(elem)
		}
		if v.Rest != nil {
			if len(v.Elements) > 0 {
				f.write(", ")
			}
			f.write("...")
			f.write(*v.Rest)
		}
		f.write("]")
	}
}

// Helper methods

func (f *Formatter) write(s string) {
	f.output.WriteString(s)
}

func (f *Formatter) writeln(s string) {
	f.output.WriteString(s)
	f.output.WriteString("\n")
}

func (f *Formatter) writeIndent() {
	for i := 0; i < f.indent; i++ {
		f.output.WriteString("  ")
	}
}

func escapeString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}
