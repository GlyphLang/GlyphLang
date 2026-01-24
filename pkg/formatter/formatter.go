// Package formatter provides bidirectional conversion between compact glyph syntax
// and expanded human-readable syntax.
//
// Compact (Glyph) syntax uses symbols:
//   : -> type definition
//   @ -> route
//   $ -> variable assignment (let)
//   > -> return
//   + -> middleware
//   % -> inject
//   ~ -> event handler
//   * -> cron task
//   ! -> command
//   & -> queue worker
//
// Expanded syntax uses keywords:
//   type, route, let, return, middleware, inject, event, cron, command, queue
package formatter

import (
	"fmt"
	"strings"

	"github.com/glyphlang/glyph/pkg/interpreter"
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
func (f *Formatter) Format(module *interpreter.Module) string {
	f.output.Reset()

	for i, item := range module.Items {
		if i > 0 {
			f.writeln("")
		}
		f.formatItem(item)
	}

	return f.output.String()
}

func (f *Formatter) formatItem(item interpreter.Item) {
	switch v := item.(type) {
	case *interpreter.TypeDef:
		f.formatTypeDef(v)
	case *interpreter.Route:
		f.formatRoute(v)
	case *interpreter.Command:
		f.formatCommand(v)
	case *interpreter.CronTask:
		f.formatCronTask(v)
	case *interpreter.EventHandler:
		f.formatEventHandler(v)
	case *interpreter.QueueWorker:
		f.formatQueueWorker(v)
	case *interpreter.Function:
		f.formatFunction(v)
	case *interpreter.WebSocketRoute:
		f.formatWebSocketRoute(v)
	case *interpreter.ImportStatement:
		f.formatImport(v)
	case *interpreter.ModuleDecl:
		f.formatModuleDecl(v)
	case *interpreter.MacroDef:
		f.formatMacroDef(v)
	}
}

func (f *Formatter) formatTypeDef(td *interpreter.TypeDef) {
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

func (f *Formatter) formatRoute(r *interpreter.Route) {
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

func (f *Formatter) formatCommand(c *interpreter.Command) {
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

func (f *Formatter) formatCronTask(ct *interpreter.CronTask) {
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

func (f *Formatter) formatEventHandler(eh *interpreter.EventHandler) {
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

func (f *Formatter) formatQueueWorker(qw *interpreter.QueueWorker) {
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

func (f *Formatter) formatFunction(fn *interpreter.Function) {
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

func (f *Formatter) formatWebSocketRoute(ws *interpreter.WebSocketRoute) {
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
		case interpreter.WSEventConnect:
			f.writeln("on connect {")
		case interpreter.WSEventDisconnect:
			f.writeln("on disconnect {")
		case interpreter.WSEventMessage:
			f.writeln("on message {")
		case interpreter.WSEventError:
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

func (f *Formatter) formatImport(imp *interpreter.ImportStatement) {
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

func (f *Formatter) formatModuleDecl(m *interpreter.ModuleDecl) {
	f.write("module \"")
	f.write(m.Name)
	f.writeln("\"")
}

func (f *Formatter) formatMacroDef(m *interpreter.MacroDef) {
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

func (f *Formatter) formatStatement(stmt interpreter.Statement) {
	f.writeIndent()

	switch v := stmt.(type) {
	// Handle both pointer and value types for each statement
	case interpreter.AssignStatement:
		f.formatAssign(v.Target, v.Value)
	case *interpreter.AssignStatement:
		f.formatAssign(v.Target, v.Value)

	case interpreter.ReassignStatement:
		f.formatReassign(v.Target, v.Value)
	case *interpreter.ReassignStatement:
		f.formatReassign(v.Target, v.Value)

	case interpreter.ReturnStatement:
		f.formatReturn(v.Value)
	case *interpreter.ReturnStatement:
		f.formatReturn(v.Value)

	case interpreter.IfStatement:
		f.formatIf(v.Condition, v.ThenBlock, v.ElseBlock)
	case *interpreter.IfStatement:
		f.formatIf(v.Condition, v.ThenBlock, v.ElseBlock)

	case interpreter.WhileStatement:
		f.formatWhile(v.Condition, v.Body)
	case *interpreter.WhileStatement:
		f.formatWhile(v.Condition, v.Body)

	case interpreter.ForStatement:
		f.formatFor(v.KeyVar, v.ValueVar, v.Iterable, v.Body)
	case *interpreter.ForStatement:
		f.formatFor(v.KeyVar, v.ValueVar, v.Iterable, v.Body)

	case interpreter.SwitchStatement:
		f.formatSwitch(v.Value, v.Cases, v.Default)
	case *interpreter.SwitchStatement:
		f.formatSwitch(v.Value, v.Cases, v.Default)

	case interpreter.ExpressionStatement:
		f.formatExpr(v.Expr)
		f.writeln("")
	case *interpreter.ExpressionStatement:
		f.formatExpr(v.Expr)
		f.writeln("")

	case interpreter.ValidationStatement:
		if f.mode == Expanded {
			f.write("validate ")
		} else {
			f.write("? ")
		}
		f.formatFunctionCall(v.Call)
		f.writeln("")
	case *interpreter.ValidationStatement:
		if f.mode == Expanded {
			f.write("validate ")
		} else {
			f.write("? ")
		}
		f.formatFunctionCall(v.Call)
		f.writeln("")

	case interpreter.DbQueryStatement:
		f.formatDbQuery(v.Var, v.Query, v.Params)
	case *interpreter.DbQueryStatement:
		f.formatDbQuery(v.Var, v.Query, v.Params)

	case interpreter.WsSendStatement:
		f.write("ws.send(")
		f.formatExpr(v.Client)
		f.write(", ")
		f.formatExpr(v.Message)
		f.writeln(")")
	case *interpreter.WsSendStatement:
		f.write("ws.send(")
		f.formatExpr(v.Client)
		f.write(", ")
		f.formatExpr(v.Message)
		f.writeln(")")

	case interpreter.WsBroadcastStatement:
		f.write("ws.broadcast(")
		f.formatExpr(v.Message)
		if v.Except != nil {
			f.write(", except: ")
			f.formatExpr(*v.Except)
		}
		f.writeln(")")
	case *interpreter.WsBroadcastStatement:
		f.write("ws.broadcast(")
		f.formatExpr(v.Message)
		if v.Except != nil {
			f.write(", except: ")
			f.formatExpr(*v.Except)
		}
		f.writeln(")")

	case interpreter.WsCloseStatement:
		f.write("ws.close(")
		f.formatExpr(v.Client)
		if v.Reason != nil {
			f.write(", ")
			f.formatExpr(v.Reason)
		}
		f.writeln(")")
	case *interpreter.WsCloseStatement:
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

func (f *Formatter) formatAssign(target string, value interpreter.Expr) {
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

func (f *Formatter) formatReassign(target string, value interpreter.Expr) {
	f.write(target)
	f.write(" = ")
	f.formatExpr(value)
	f.writeln("")
}

func (f *Formatter) formatReturn(value interpreter.Expr) {
	if f.mode == Expanded {
		f.write("return ")
	} else {
		f.write("> ")
	}
	f.formatExpr(value)
	f.writeln("")
}

func (f *Formatter) formatIf(condition interpreter.Expr, thenBlock, elseBlock []interpreter.Statement) {
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

func (f *Formatter) formatWhile(condition interpreter.Expr, body []interpreter.Statement) {
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

func (f *Formatter) formatFor(keyVar, valueVar string, iterable interpreter.Expr, body []interpreter.Statement) {
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

func (f *Formatter) formatSwitch(value interpreter.Expr, cases []interpreter.SwitchCase, defaultBlock []interpreter.Statement) {
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

func (f *Formatter) formatDbQuery(varName, query string, params []interpreter.Expr) {
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

func (f *Formatter) formatFunctionCall(call interpreter.FunctionCallExpr) {
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

func (f *Formatter) formatExpr(expr interpreter.Expr) {
	switch v := expr.(type) {
	case interpreter.LiteralExpr:
		f.formatLiteral(v.Value)
	case *interpreter.LiteralExpr:
		f.formatLiteral(v.Value)

	case interpreter.VariableExpr:
		f.write(v.Name)
	case *interpreter.VariableExpr:
		f.write(v.Name)

	case interpreter.BinaryOpExpr:
		f.formatBinaryOp(v.Op, v.Left, v.Right)
	case *interpreter.BinaryOpExpr:
		f.formatBinaryOp(v.Op, v.Left, v.Right)

	case interpreter.UnaryOpExpr:
		f.write(v.Op.String())
		f.formatExpr(v.Right)
	case *interpreter.UnaryOpExpr:
		f.write(v.Op.String())
		f.formatExpr(v.Right)

	case interpreter.FieldAccessExpr:
		f.formatExpr(v.Object)
		f.write(".")
		f.write(v.Field)
	case *interpreter.FieldAccessExpr:
		f.formatExpr(v.Object)
		f.write(".")
		f.write(v.Field)

	case interpreter.ArrayIndexExpr:
		f.formatExpr(v.Array)
		f.write("[")
		f.formatExpr(v.Index)
		f.write("]")
	case *interpreter.ArrayIndexExpr:
		f.formatExpr(v.Array)
		f.write("[")
		f.formatExpr(v.Index)
		f.write("]")

	case interpreter.FunctionCallExpr:
		f.formatFunctionCall(v)
	case *interpreter.FunctionCallExpr:
		f.formatFunctionCall(*v)

	case interpreter.ObjectExpr:
		f.formatObject(v.Fields)
	case *interpreter.ObjectExpr:
		f.formatObject(v.Fields)

	case interpreter.ArrayExpr:
		f.formatArray(v.Elements)
	case *interpreter.ArrayExpr:
		f.formatArray(v.Elements)

	case interpreter.LambdaExpr:
		f.formatLambda(v.Params, v.Body, v.Block)
	case *interpreter.LambdaExpr:
		f.formatLambda(v.Params, v.Body, v.Block)

	case interpreter.MatchExpr:
		f.formatMatch(v.Value, v.Cases)
	case *interpreter.MatchExpr:
		f.formatMatch(v.Value, v.Cases)

	case interpreter.AsyncExpr:
		f.formatAsync(v.Body)
	case *interpreter.AsyncExpr:
		f.formatAsync(v.Body)

	case interpreter.AwaitExpr:
		f.write("await ")
		f.formatExpr(v.Expr)
	case *interpreter.AwaitExpr:
		f.write("await ")
		f.formatExpr(v.Expr)
	}
}

func (f *Formatter) formatBinaryOp(op interpreter.BinOp, left, right interpreter.Expr) {
	f.formatExpr(left)
	f.write(" ")
	f.write(op.String())
	f.write(" ")
	f.formatExpr(right)
}

func (f *Formatter) formatObject(fields []interpreter.ObjectField) {
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

func (f *Formatter) formatArray(elements []interpreter.Expr) {
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

func (f *Formatter) formatLambda(params []interpreter.Field, body interpreter.Expr, block []interpreter.Statement) {
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

func (f *Formatter) formatMatch(value interpreter.Expr, cases []interpreter.MatchCase) {
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

func (f *Formatter) formatAsync(body []interpreter.Statement) {
	f.writeln("async {")
	f.indent++
	for _, s := range body {
		f.formatStatement(s)
	}
	f.indent--
	f.writeIndent()
	f.write("}")
}

func (f *Formatter) formatLiteral(lit interpreter.Literal) {
	switch v := lit.(type) {
	case interpreter.IntLiteral:
		f.write(fmt.Sprintf("%d", v.Value))
	case interpreter.FloatLiteral:
		f.write(fmt.Sprintf("%g", v.Value))
	case interpreter.StringLiteral:
		f.write("\"")
		f.write(escapeString(v.Value))
		f.write("\"")
	case interpreter.BoolLiteral:
		if v.Value {
			f.write("true")
		} else {
			f.write("false")
		}
	case interpreter.NullLiteral:
		f.write("null")
	}
}

func (f *Formatter) formatType(t interpreter.Type) {
	switch v := t.(type) {
	case interpreter.IntType:
		f.write("int")
	case interpreter.StringType:
		f.write("str")
	case interpreter.BoolType:
		f.write("bool")
	case interpreter.FloatType:
		f.write("float")
	case interpreter.DatabaseType:
		f.write("Database")
	case interpreter.NamedType:
		f.write(v.Name)
	case interpreter.ArrayType:
		f.formatType(v.ElementType)
		f.write("[]")
	case interpreter.OptionalType:
		f.formatType(v.InnerType)
		f.write("?")
	case interpreter.GenericType:
		f.formatType(v.BaseType)
		f.write("<")
		for i, arg := range v.TypeArgs {
			if i > 0 {
				f.write(", ")
			}
			f.formatType(arg)
		}
		f.write(">")
	case interpreter.TypeParameterType:
		f.write(v.Name)
	case interpreter.FunctionType:
		f.write("(")
		for i, pt := range v.ParamTypes {
			if i > 0 {
				f.write(", ")
			}
			f.formatType(pt)
		}
		f.write(") -> ")
		f.formatType(v.ReturnType)
	case interpreter.UnionType:
		for i, ut := range v.Types {
			if i > 0 {
				f.write(" | ")
			}
			f.formatType(ut)
		}
	case interpreter.FutureType:
		f.write("Future<")
		f.formatType(v.ResultType)
		f.write(">")
	}
}

func (f *Formatter) formatPattern(p interpreter.Pattern) {
	switch v := p.(type) {
	case interpreter.LiteralPattern:
		f.formatLiteral(v.Value)
	case interpreter.VariablePattern:
		f.write(v.Name)
	case interpreter.WildcardPattern:
		f.write("_")
	case interpreter.ObjectPattern:
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
	case interpreter.ArrayPattern:
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
