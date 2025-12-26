package compiler

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"

	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/glyphlang/glyph/pkg/vm"
)

// Compiler compiles AST to bytecode
type Compiler struct {
	constants     []vm.Value
	code          []byte
	symbolTable   *SymbolTable
	labelCounter  int
	optimizer     *Optimizer
	macroExpander *MacroExpander
}

// NewCompiler creates a new compiler instance
func NewCompiler() *Compiler {
	return &Compiler{
		constants:     make([]vm.Value, 0),
		code:          make([]byte, 0),
		symbolTable:   NewGlobalSymbolTable(),
		labelCounter:  0,
		optimizer:     NewOptimizer(OptBasic), // Default to basic optimization
		macroExpander: NewMacroExpander(),
	}
}

// NewCompilerWithOptLevel creates a new compiler with specified optimization level
func NewCompilerWithOptLevel(level OptimizationLevel) *Compiler {
	return &Compiler{
		constants:     make([]vm.Value, 0),
		code:          make([]byte, 0),
		symbolTable:   NewGlobalSymbolTable(),
		labelCounter:  0,
		optimizer:     NewOptimizer(level),
		macroExpander: NewMacroExpander(),
	}
}

// Reset clears the compiler state for compiling a new route
func (c *Compiler) Reset() {
	c.constants = make([]vm.Value, 0)
	c.code = make([]byte, 0)
	c.symbolTable = NewGlobalSymbolTable()
	c.labelCounter = 0
	// Keep the optimizer with its current settings
}

// Compile compiles an AST module to bytecode
func (c *Compiler) Compile(module *interpreter.Module) ([]byte, error) {
	// First, expand all macros in the module
	expandedModule, err := c.macroExpander.ExpandModule(module)
	if err != nil {
		return nil, fmt.Errorf("macro expansion failed: %w", err)
	}

	// For now, compile the first route we find
	for _, item := range expandedModule.Items {
		if route, ok := item.(*interpreter.Route); ok {
			return c.CompileRoute(route)
		}
	}

	// Check if this is a type-only module (library module)
	// Type-only modules are valid but don't produce executable bytecode
	hasTypeDefs := false
	for _, item := range expandedModule.Items {
		if _, ok := item.(*interpreter.TypeDef); ok {
			hasTypeDefs = true
			break
		}
	}

	if hasTypeDefs {
		// Return minimal bytecode for type-only modules
		// Types are compile-time only, so just return a halt instruction
		c.Reset()
		c.emit(vm.OpHalt)
		return c.buildBytecode(), nil
	}

	// Empty module
	if len(expandedModule.Items) == 0 {
		return nil, fmt.Errorf("empty module: no items to compile")
	}

	return nil, fmt.Errorf("no route found to compile (module contains %d items but no routes)", len(expandedModule.Items))
}

// CompileRoute compiles a route to bytecode
func (c *Compiler) CompileRoute(route *interpreter.Route) ([]byte, error) {
	// Reset compiler state for each route
	c.Reset()

	// Create route scope
	c.symbolTable = c.symbolTable.EnterScope(RouteScope)

	// Add route parameters to symbol table
	// Extract params from route.Path (anything after :)
	params := extractRouteParams(route.Path)
	for _, param := range params {
		nameIdx := c.addConstant(vm.StringValue{Val: param})
		c.symbolTable.Define(param, nameIdx)
	}

	// Add injections to symbol table
	for _, injection := range route.Injections {
		nameIdx := c.addConstant(vm.StringValue{Val: injection.Name})
		c.symbolTable.Define(injection.Name, nameIdx)
	}

	// Add built-in request variables that are auto-injected at runtime
	// query - URL query parameters (always available)
	queryIdx := c.addConstant(vm.StringValue{Val: "query"})
	c.symbolTable.Define("query", queryIdx)

	// input - Request body (always available, may be nil)
	inputIdx := c.addConstant(vm.StringValue{Val: "input"})
	c.symbolTable.Define("input", inputIdx)

	// ws - WebSocket manager (for accessing WebSocket state from REST routes)
	wsIdx := c.addConstant(vm.StringValue{Val: "ws"})
	c.symbolTable.Define("ws", wsIdx)

	// auth - Authenticated user info (available when route has + auth middleware)
	if route.Auth != nil {
		authIdx := c.addConstant(vm.StringValue{Val: "auth"})
		c.symbolTable.Define("auth", authIdx)
	}

	// Optimize route body before compilation
	optimizedBody := c.optimizer.OptimizeStatements(route.Body)

	// Compile optimized route body
	for _, stmt := range optimizedBody {
		if err := c.compileStatement(stmt); err != nil {
			return nil, err
		}
	}

	// If last statement isn't a return, add OpHalt
	if len(optimizedBody) == 0 || !isReturnStatement(optimizedBody[len(optimizedBody)-1]) {
		c.emit(vm.OpHalt)
	}

	// Build final bytecode
	return c.buildBytecode(), nil
}

// CompileCommand compiles a CLI command to bytecode
func (c *Compiler) CompileCommand(cmd *interpreter.Command) ([]byte, error) {
	c.Reset()

	// Create command scope
	c.symbolTable = c.symbolTable.EnterScope(RouteScope)

	// Add command parameters to symbol table
	for _, param := range cmd.Params {
		nameIdx := c.addConstant(vm.StringValue{Val: param.Name})
		c.symbolTable.Define(param.Name, nameIdx)
	}

	// Optimize and compile body
	optimizedBody := c.optimizer.OptimizeStatements(cmd.Body)
	for _, stmt := range optimizedBody {
		if err := c.compileStatement(stmt); err != nil {
			return nil, err
		}
	}

	if len(optimizedBody) == 0 || !isReturnStatement(optimizedBody[len(optimizedBody)-1]) {
		c.emit(vm.OpHalt)
	}

	return c.buildBytecode(), nil
}

// CompileCronTask compiles a cron task to bytecode
func (c *Compiler) CompileCronTask(task *interpreter.CronTask) ([]byte, error) {
	c.Reset()

	// Create task scope
	c.symbolTable = c.symbolTable.EnterScope(RouteScope)

	// Add injections to symbol table
	for _, injection := range task.Injections {
		nameIdx := c.addConstant(vm.StringValue{Val: injection.Name})
		c.symbolTable.Define(injection.Name, nameIdx)
	}

	// Optimize and compile body
	optimizedBody := c.optimizer.OptimizeStatements(task.Body)
	for _, stmt := range optimizedBody {
		if err := c.compileStatement(stmt); err != nil {
			return nil, err
		}
	}

	if len(optimizedBody) == 0 || !isReturnStatement(optimizedBody[len(optimizedBody)-1]) {
		c.emit(vm.OpHalt)
	}

	return c.buildBytecode(), nil
}

// CompileEventHandler compiles an event handler to bytecode
func (c *Compiler) CompileEventHandler(handler *interpreter.EventHandler) ([]byte, error) {
	c.Reset()

	// Create handler scope
	c.symbolTable = c.symbolTable.EnterScope(RouteScope)

	// Add event data variable
	eventIdx := c.addConstant(vm.StringValue{Val: "event"})
	c.symbolTable.Define("event", eventIdx)
	inputIdx := c.addConstant(vm.StringValue{Val: "input"})
	c.symbolTable.Define("input", inputIdx)

	// Add injections to symbol table
	for _, injection := range handler.Injections {
		nameIdx := c.addConstant(vm.StringValue{Val: injection.Name})
		c.symbolTable.Define(injection.Name, nameIdx)
	}

	// Optimize and compile body
	optimizedBody := c.optimizer.OptimizeStatements(handler.Body)
	for _, stmt := range optimizedBody {
		if err := c.compileStatement(stmt); err != nil {
			return nil, err
		}
	}

	if len(optimizedBody) == 0 || !isReturnStatement(optimizedBody[len(optimizedBody)-1]) {
		c.emit(vm.OpHalt)
	}

	return c.buildBytecode(), nil
}

// CompileQueueWorker compiles a queue worker to bytecode
func (c *Compiler) CompileQueueWorker(worker *interpreter.QueueWorker) ([]byte, error) {
	c.Reset()

	// Create worker scope
	c.symbolTable = c.symbolTable.EnterScope(RouteScope)

	// Add message variable
	messageIdx := c.addConstant(vm.StringValue{Val: "message"})
	c.symbolTable.Define("message", messageIdx)
	inputIdx := c.addConstant(vm.StringValue{Val: "input"})
	c.symbolTable.Define("input", inputIdx)

	// Add injections to symbol table
	for _, injection := range worker.Injections {
		nameIdx := c.addConstant(vm.StringValue{Val: injection.Name})
		c.symbolTable.Define(injection.Name, nameIdx)
	}

	// Optimize and compile body
	optimizedBody := c.optimizer.OptimizeStatements(worker.Body)
	for _, stmt := range optimizedBody {
		if err := c.compileStatement(stmt); err != nil {
			return nil, err
		}
	}

	if len(optimizedBody) == 0 || !isReturnStatement(optimizedBody[len(optimizedBody)-1]) {
		c.emit(vm.OpHalt)
	}

	return c.buildBytecode(), nil
}

// compileStatement compiles a statement
func (c *Compiler) compileStatement(stmt interpreter.Statement) error {
	switch s := stmt.(type) {
	case *interpreter.AssignStatement:
		return c.compileAssignStatement(s)
	case interpreter.AssignStatement:
		return c.compileAssignStatement(&s)
	case *interpreter.ReturnStatement:
		return c.compileReturnStatement(s)
	case interpreter.ReturnStatement:
		return c.compileReturnStatement(&s)
	case *interpreter.IfStatement:
		return c.compileIfStatement(s)
	case interpreter.IfStatement:
		return c.compileIfStatement(&s)
	case *interpreter.WhileStatement:
		return c.compileWhileStatement(s)
	case interpreter.WhileStatement:
		return c.compileWhileStatement(&s)
	case *interpreter.ValidationStatement:
		return c.compileValidationStatement(s)
	case interpreter.ValidationStatement:
		return c.compileValidationStatement(&s)
	case *interpreter.ForStatement:
		return c.compileForStatement(s)
	case interpreter.ForStatement:
		return c.compileForStatement(&s)
	case *interpreter.SwitchStatement:
		return c.compileSwitchStatement(s)
	case interpreter.SwitchStatement:
		return c.compileSwitchStatement(&s)
	case *interpreter.ExpressionStatement:
		return c.compileExpressionStatement(s)
	case interpreter.ExpressionStatement:
		return c.compileExpressionStatement(&s)
	default:
		return fmt.Errorf("unsupported statement type: %T", stmt)
	}
}

// compileAssignStatement compiles variable assignment
func (c *Compiler) compileAssignStatement(stmt *interpreter.AssignStatement) error {
	// Compile the value expression
	if err := c.compileExpression(stmt.Value); err != nil {
		return err
	}

	// Add variable name to constants
	nameIdx := c.addConstant(vm.StringValue{Val: stmt.Target})

	// Emit store instruction
	c.emitWithOperand(vm.OpStoreVar, uint32(nameIdx))

	// Define symbol
	c.symbolTable.Define(stmt.Target, nameIdx)

	return nil
}

// compileReturnStatement compiles return statement
func (c *Compiler) compileReturnStatement(stmt *interpreter.ReturnStatement) error {
	// Compile return value
	if err := c.compileExpression(stmt.Value); err != nil {
		return err
	}

	// Emit return instruction
	c.emit(vm.OpReturn)

	return nil
}

// compileIfStatement compiles if statement
func (c *Compiler) compileIfStatement(stmt *interpreter.IfStatement) error {
	// Compile condition
	if err := c.compileExpression(stmt.Condition); err != nil {
		return err
	}

	// Emit jump if false (will patch later)
	jumpToElse := len(c.code)
	c.emitWithOperand(vm.OpJumpIfFalse, 0) // Placeholder

	// Compile then block
	for _, thenStmt := range stmt.ThenBlock {
		if err := c.compileStatement(thenStmt); err != nil {
			return err
		}
	}

	// Emit jump to end (will patch later)
	jumpToEnd := len(c.code)
	c.emitWithOperand(vm.OpJump, 0) // Placeholder

	// Patch jump to else
	elseOffset := len(c.code)
	c.patchJump(jumpToElse, uint32(elseOffset))

	// Compile else block
	if len(stmt.ElseBlock) > 0 {
		for _, elseStmt := range stmt.ElseBlock {
			if err := c.compileStatement(elseStmt); err != nil {
				return err
			}
		}
	}

	// Patch jump to end
	endOffset := len(c.code)
	c.patchJump(jumpToEnd, uint32(endOffset))

	return nil
}

// compileWhileStatement compiles while loop
func (c *Compiler) compileWhileStatement(stmt *interpreter.WhileStatement) error {
	// Remember loop start
	loopStart := len(c.code)

	// Compile condition
	if err := c.compileExpression(stmt.Condition); err != nil {
		return err
	}

	// Emit jump if false to end (will patch later)
	jumpToEnd := len(c.code)
	c.emitWithOperand(vm.OpJumpIfFalse, 0) // Placeholder

	// Compile body
	for _, bodyStmt := range stmt.Body {
		if err := c.compileStatement(bodyStmt); err != nil {
			return err
		}
	}

	// Jump back to loop start
	c.emitWithOperand(vm.OpJump, uint32(loopStart))

	// Patch jump to end
	endOffset := len(c.code)
	c.patchJump(jumpToEnd, uint32(endOffset))

	return nil
}

// compileValidationStatement compiles validation statement (no-op in compiled mode for now)
func (c *Compiler) compileValidationStatement(stmt *interpreter.ValidationStatement) error {
	// For now, validation statements are ignored in compiled mode
	// In production, you might want to compile them as actual validation calls
	return nil
}

// compileExpressionStatement compiles expression statement (e.g., function call)
func (c *Compiler) compileExpressionStatement(stmt *interpreter.ExpressionStatement) error {
	// Compile the expression (side effects like function calls)
	if err := c.compileExpression(stmt.Expr); err != nil {
		return err
	}
	// Pop the result since we're not using it
	c.emit(vm.OpPop)
	return nil
}

// compileForStatement compiles for loop statement
func (c *Compiler) compileForStatement(stmt *interpreter.ForStatement) error {
	// Create scope for loop variables
	parentSymbolTable := c.symbolTable
	c.symbolTable = c.symbolTable.EnterScope(BlockScope)
	defer func() {
		c.symbolTable = parentSymbolTable
	}()

	// Compile the iterable expression (pushes collection onto stack)
	if err := c.compileExpression(stmt.Iterable); err != nil {
		return fmt.Errorf("for loop iterable: %w", err)
	}

	// Emit OpGetIter - creates iterator, pushes iterator ID onto stack
	c.emit(vm.OpGetIter)

	// Store iterator ID in a temp variable
	iterVarName := fmt.Sprintf("__iter_%d", c.labelCounter)
	c.labelCounter++
	iterNameIdx := c.addConstant(vm.StringValue{Val: iterVarName})
	c.symbolTable.Define(iterVarName, iterNameIdx)
	c.emitWithOperand(vm.OpStoreVar, uint32(iterNameIdx))

	// Define loop variables in symbol table
	valueNameIdx := c.addConstant(vm.StringValue{Val: stmt.ValueVar})
	c.symbolTable.Define(stmt.ValueVar, valueNameIdx)

	var keyNameIdx int
	hasKey := stmt.KeyVar != ""
	if hasKey {
		keyNameIdx = c.addConstant(vm.StringValue{Val: stmt.KeyVar})
		c.symbolTable.Define(stmt.KeyVar, keyNameIdx)
	}

	// Loop start - check if iterator has more elements
	loopStart := len(c.code)

	// Load iterator ID for hasNext check
	c.emitWithOperand(vm.OpLoadVar, uint32(iterNameIdx))

	// Emit OpIterHasNext - pops iterator ID, pushes bool
	c.emit(vm.OpIterHasNext)

	// Jump to end if no more elements
	jumpToEnd := len(c.code)
	c.emitWithOperand(vm.OpJumpIfFalse, 0) // Placeholder

	// Load iterator ID for next element
	c.emitWithOperand(vm.OpLoadVar, uint32(iterNameIdx))

	// Emit OpIterNext with operand indicating if we have a key var
	// OpIterNext pops iterator ID, pushes key (if hasKey) and value
	if hasKey {
		c.emitWithOperand(vm.OpIterNext, 1)
	} else {
		c.emitWithOperand(vm.OpIterNext, 0)
	}

	// Store loop variables - value is on top of stack
	c.emitWithOperand(vm.OpStoreVar, uint32(valueNameIdx))

	// If we have a key variable, store it too (it's next on stack)
	if hasKey {
		c.emitWithOperand(vm.OpStoreVar, uint32(keyNameIdx))
	}

	// Compile loop body
	for _, bodyStmt := range stmt.Body {
		if err := c.compileStatement(bodyStmt); err != nil {
			return fmt.Errorf("for loop body: %w", err)
		}
	}

	// Jump back to loop start
	c.emitWithOperand(vm.OpJump, uint32(loopStart))

	// Patch jump to end
	endOffset := len(c.code)
	c.patchJump(jumpToEnd, uint32(endOffset))

	return nil
}

// compileSwitchStatement compiles switch statement
func (c *Compiler) compileSwitchStatement(stmt *interpreter.SwitchStatement) error {
	// Create a temporary variable to store the switch value
	switchVarName := fmt.Sprintf("__switch_%d", c.labelCounter)
	c.labelCounter++
	switchVarIdx := c.addConstant(vm.StringValue{Val: switchVarName})
	c.symbolTable.Define(switchVarName, switchVarIdx)

	// Compile and store the switch value
	if err := c.compileExpression(stmt.Value); err != nil {
		return err
	}
	c.emitWithOperand(vm.OpStoreVar, uint32(switchVarIdx))

	// Track jump locations for each case
	var jumpToEnd []int

	// Compile each case
	for _, switchCase := range stmt.Cases {
		// Load switch value for comparison
		c.emitWithOperand(vm.OpLoadVar, uint32(switchVarIdx))

		// Compile case value
		if err := c.compileExpression(switchCase.Value); err != nil {
			return err
		}

		// Compare
		c.emit(vm.OpEq)

		// Jump to next case if false
		jumpToNextCase := len(c.code)
		c.emitWithOperand(vm.OpJumpIfFalse, 0)

		// Compile case body
		for _, stmt := range switchCase.Body {
			if err := c.compileStatement(stmt); err != nil {
				return err
			}
		}

		// Jump to end after executing case
		jumpToEnd = append(jumpToEnd, len(c.code))
		c.emitWithOperand(vm.OpJump, 0)

		// Patch jump to next case
		c.patchJump(jumpToNextCase, uint32(len(c.code)))
	}

	// Compile default case if present
	if len(stmt.Default) > 0 {
		for _, stmt := range stmt.Default {
			if err := c.compileStatement(stmt); err != nil {
				return err
			}
		}
	}

	// Patch all jumps to end
	endOffset := uint32(len(c.code))
	for _, jumpLoc := range jumpToEnd {
		c.patchJump(jumpLoc, endOffset)
	}

	return nil
}

// compileExpression compiles an expression
func (c *Compiler) compileExpression(expr interpreter.Expr) error {
	switch e := expr.(type) {
	case *interpreter.LiteralExpr:
		return c.compileLiteral(e)
	case interpreter.LiteralExpr:
		return c.compileLiteral(&e)
	case *interpreter.VariableExpr:
		return c.compileVariable(e)
	case interpreter.VariableExpr:
		return c.compileVariable(&e)
	case *interpreter.BinaryOpExpr:
		return c.compileBinaryOp(e)
	case interpreter.BinaryOpExpr:
		return c.compileBinaryOp(&e)
	case *interpreter.ObjectExpr:
		return c.compileObject(e)
	case interpreter.ObjectExpr:
		return c.compileObject(&e)
	case *interpreter.ArrayExpr:
		return c.compileArray(e)
	case interpreter.ArrayExpr:
		return c.compileArray(&e)
	case *interpreter.FieldAccessExpr:
		return c.compileFieldAccess(e)
	case interpreter.FieldAccessExpr:
		return c.compileFieldAccess(&e)
	case *interpreter.FunctionCallExpr:
		return c.compileFunctionCall(e)
	case interpreter.FunctionCallExpr:
		return c.compileFunctionCall(&e)
	case *interpreter.ArrayIndexExpr:
		return c.compileArrayIndex(e)
	case interpreter.ArrayIndexExpr:
		return c.compileArrayIndex(&e)
	case *interpreter.UnaryOpExpr:
		return c.compileUnaryOp(e)
	case interpreter.UnaryOpExpr:
		return c.compileUnaryOp(&e)
	case *interpreter.MatchExpr:
		return c.compileMatchExpr(e)
	case interpreter.MatchExpr:
		return c.compileMatchExpr(&e)
	default:
		return fmt.Errorf("unsupported expression type: %T", expr)
	}
}

// compileLiteral compiles a literal value
func (c *Compiler) compileLiteral(expr *interpreter.LiteralExpr) error {
	var val vm.Value

	switch lit := expr.Value.(type) {
	case interpreter.IntLiteral:
		val = vm.IntValue{Val: lit.Value}
	case interpreter.FloatLiteral:
		val = vm.FloatValue{Val: lit.Value}
	case interpreter.StringLiteral:
		val = vm.StringValue{Val: lit.Value}
	case interpreter.BoolLiteral:
		val = vm.BoolValue{Val: lit.Value}
	case interpreter.NullLiteral:
		val = vm.NullValue{}
	default:
		return fmt.Errorf("unsupported literal type: %T", expr.Value)
	}

	// Add to constants and emit push
	idx := c.addConstant(val)
	c.emitWithOperand(vm.OpPush, uint32(idx))

	return nil
}

// compileVariable compiles variable reference
func (c *Compiler) compileVariable(expr *interpreter.VariableExpr) error {
	// Look up symbol
	symbol, ok := c.symbolTable.Resolve(expr.Name)
	if !ok {
		return fmt.Errorf("undefined variable: %s", expr.Name)
	}

	// Emit load instruction
	c.emitWithOperand(vm.OpLoadVar, uint32(symbol.Index))

	return nil
}

// compileBinaryOp compiles binary operation
func (c *Compiler) compileBinaryOp(expr *interpreter.BinaryOpExpr) error {
	// Compile left operand
	if err := c.compileExpression(expr.Left); err != nil {
		return err
	}

	// Compile right operand
	if err := c.compileExpression(expr.Right); err != nil {
		return err
	}

	// Emit operation
	switch expr.Op {
	case interpreter.Add:
		c.emit(vm.OpAdd)
	case interpreter.Sub:
		c.emit(vm.OpSub)
	case interpreter.Mul:
		c.emit(vm.OpMul)
	case interpreter.Div:
		c.emit(vm.OpDiv)
	case interpreter.Eq:
		c.emit(vm.OpEq)
	case interpreter.Ne:
		c.emit(vm.OpNe)
	case interpreter.Lt:
		c.emit(vm.OpLt)
	case interpreter.Le:
		c.emit(vm.OpLe)
	case interpreter.Gt:
		c.emit(vm.OpGt)
	case interpreter.Ge:
		c.emit(vm.OpGe)
	case interpreter.And:
		c.emit(vm.OpAnd)
	case interpreter.Or:
		c.emit(vm.OpOr)
	default:
		return fmt.Errorf("unsupported binary operator: %v", expr.Op)
	}

	return nil
}

// compileUnaryOp compiles unary operation
func (c *Compiler) compileUnaryOp(expr *interpreter.UnaryOpExpr) error {
	// Compile the operand
	if err := c.compileExpression(expr.Right); err != nil {
		return err
	}

	// Emit operation
	switch expr.Op {
	case interpreter.Not:
		c.emit(vm.OpNot)
	case interpreter.Neg:
		c.emit(vm.OpNeg)
	default:
		return fmt.Errorf("unsupported unary operator: %v", expr.Op)
	}

	return nil
}

// compileObject compiles object literal
func (c *Compiler) compileObject(expr *interpreter.ObjectExpr) error {
	// Compile each field (key-value pairs)
	for _, field := range expr.Fields {
		// Push key (field name)
		keyIdx := c.addConstant(vm.StringValue{Val: field.Key})
		c.emitWithOperand(vm.OpPush, uint32(keyIdx))

		// Push value
		if err := c.compileExpression(field.Value); err != nil {
			return err
		}
	}

	// Emit build object instruction
	c.emitWithOperand(vm.OpBuildObject, uint32(len(expr.Fields)))

	return nil
}

// compileArray compiles array literal
func (c *Compiler) compileArray(expr *interpreter.ArrayExpr) error {
	// Compile each element
	for _, elem := range expr.Elements {
		if err := c.compileExpression(elem); err != nil {
			return err
		}
	}

	// Emit build array instruction
	c.emitWithOperand(vm.OpBuildArray, uint32(len(expr.Elements)))

	return nil
}

// compileFieldAccess compiles field access (obj.field)
func (c *Compiler) compileFieldAccess(expr *interpreter.FieldAccessExpr) error {
	// Compile object expression
	if err := c.compileExpression(expr.Object); err != nil {
		return err
	}

	// Push field name
	fieldIdx := c.addConstant(vm.StringValue{Val: expr.Field})
	c.emitWithOperand(vm.OpPush, uint32(fieldIdx))

	// Emit get field instruction
	c.emit(vm.OpGetField)

	return nil
}

// compileArrayIndex compiles array indexing (array[index])
func (c *Compiler) compileArrayIndex(expr *interpreter.ArrayIndexExpr) error {
	// Compile array expression (pushes array onto stack)
	if err := c.compileExpression(expr.Array); err != nil {
		return err
	}

	// Compile index expression (pushes index onto stack)
	if err := c.compileExpression(expr.Index); err != nil {
		return err
	}

	// Emit get index instruction
	// Stack before: [..., array, index]
	// Stack after: [..., element]
	c.emit(vm.OpGetIndex)

	return nil
}

// compileFunctionCall compiles a function call
func (c *Compiler) compileFunctionCall(expr *interpreter.FunctionCallExpr) error {
	// Check for WebSocket functions first (ws.*)
	if strings.HasPrefix(expr.Name, "ws.") {
		handled, err := c.compileFunctionCallForWs(expr)
		if err != nil {
			return err
		}
		if handled {
			return nil
		}
	}

	// Push function name first (it will be at bottom of stack)
	fnNameIdx := c.addConstant(vm.StringValue{Val: expr.Name})
	c.emitWithOperand(vm.OpPush, uint32(fnNameIdx))

	// Compile arguments in order (they will be on top of function name)
	for _, arg := range expr.Args {
		if err := c.compileExpression(arg); err != nil {
			return fmt.Errorf("failed to compile function argument: %w", err)
		}
	}

	// Emit call instruction with argument count
	c.emitWithOperand(vm.OpCall, uint32(len(expr.Args)))

	return nil
}

// Helper methods

// addConstant adds a constant to the pool and returns its index
func (c *Compiler) addConstant(val vm.Value) int {
	// Check if constant already exists (deduplication)
	for i, existing := range c.constants {
		if valuesEqual(existing, val) {
			return i
		}
	}

	c.constants = append(c.constants, val)
	return len(c.constants) - 1
}

// emit emits a single opcode
func (c *Compiler) emit(opcode vm.Opcode) {
	c.code = append(c.code, byte(opcode))
}

// emitWithOperand emits an opcode with a 4-byte operand
func (c *Compiler) emitWithOperand(opcode vm.Opcode, operand uint32) {
	c.code = append(c.code, byte(opcode))
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, operand)
	c.code = append(c.code, buf...)
}

// patchJump patches a jump instruction at the given offset
func (c *Compiler) patchJump(offset int, target uint32) {
	// Jump instruction format: opcode (1 byte) + operand (4 bytes)
	// We need to patch the operand at offset+1
	binary.LittleEndian.PutUint32(c.code[offset+1:offset+5], target)
}

// buildBytecode constructs the final bytecode with header
func (c *Compiler) buildBytecode() []byte {
	bytecode := []byte{0x47, 0x4C, 0x59, 0x50} // Magic "GLYP"

	// Version (little-endian u32)
	version := make([]byte, 4)
	binary.LittleEndian.PutUint32(version, 1)
	bytecode = append(bytecode, version...)

	// Constant count
	constCount := make([]byte, 4)
	binary.LittleEndian.PutUint32(constCount, uint32(len(c.constants)))
	bytecode = append(bytecode, constCount...)

	// Serialize constants
	for _, constant := range c.constants {
		bytecode = append(bytecode, serializeConstant(constant)...)
	}

	// Instruction count
	instrCount := make([]byte, 4)
	binary.LittleEndian.PutUint32(instrCount, uint32(len(c.code)))
	bytecode = append(bytecode, instrCount...)

	// Calculate header size (offset where instructions start)
	headerSize := len(bytecode)

	// Adjust all jump targets by adding header offset
	// Jump targets are stored as relative offsets to c.code, but VM expects
	// absolute offsets into the full bytecode buffer
	c.adjustJumpTargets(uint32(headerSize))

	// Instructions
	bytecode = append(bytecode, c.code...)

	return bytecode
}

// adjustJumpTargets adjusts all jump instruction operands by adding the header offset
func (c *Compiler) adjustJumpTargets(headerOffset uint32) {
	// Jump opcodes that have a target operand
	jumpOpcodes := map[byte]bool{
		byte(vm.OpJump):        true,
		byte(vm.OpJumpIfFalse): true,
		byte(vm.OpJumpIfTrue):  true,
	}

	i := 0
	for i < len(c.code) {
		opcode := c.code[i]
		i++

		if jumpOpcodes[opcode] {
			// Read the current operand (4 bytes, little-endian)
			if i+4 <= len(c.code) {
				currentTarget := binary.LittleEndian.Uint32(c.code[i : i+4])
				// Add header offset to make it absolute
				newTarget := currentTarget + headerOffset
				binary.LittleEndian.PutUint32(c.code[i:i+4], newTarget)
			}
			i += 4
		} else if hasOperand(opcode) {
			// Skip operand for other instructions with operands
			i += 4
		}
	}
}

// hasOperand returns true if the opcode has a 4-byte operand
func hasOperand(opcode byte) bool {
	// Opcodes with operands (based on vm.go)
	withOperand := map[byte]bool{
		byte(vm.OpPush):            true,
		byte(vm.OpLoadVar):         true,
		byte(vm.OpStoreVar):        true,
		byte(vm.OpJump):            true,
		byte(vm.OpJumpIfFalse):     true,
		byte(vm.OpJumpIfTrue):      true,
		byte(vm.OpIterNext):        true,
		byte(vm.OpCall):            true,
		byte(vm.OpBuildObject):     true,
		byte(vm.OpBuildArray):      true,
	}
	return withOperand[opcode]
}

// serializeConstant serializes a constant value
func serializeConstant(c vm.Value) []byte {
	switch v := c.(type) {
	case vm.NullValue:
		return []byte{0x00}
	case vm.IntValue:
		buf := make([]byte, 9)
		buf[0] = 0x01
		binary.LittleEndian.PutUint64(buf[1:], uint64(v.Val))
		return buf
	case vm.FloatValue:
		buf := make([]byte, 9)
		buf[0] = 0x02
		binary.LittleEndian.PutUint64(buf[1:], math.Float64bits(v.Val))
		return buf
	case vm.BoolValue:
		if v.Val {
			return []byte{0x03, 0x01}
		}
		return []byte{0x03, 0x00}
	case vm.StringValue:
		buf := []byte{0x04}
		length := make([]byte, 4)
		binary.LittleEndian.PutUint32(length, uint32(len(v.Val)))
		buf = append(buf, length...)
		buf = append(buf, []byte(v.Val)...)
		return buf
	default:
		return nil
	}
}

// Utility functions

func isReturnStatement(stmt interpreter.Statement) bool {
	_, ok := stmt.(*interpreter.ReturnStatement)
	return ok
}

// extractRouteParams extracts parameter names from a route path
// e.g., "/users/:id/:action" returns ["id", "action"]
func extractRouteParams(path string) []string {
	params := []string{}
	parts := []rune(path)

	for i := 0; i < len(parts); i++ {
		if parts[i] == ':' {
			// Found a parameter, extract the name
			paramStart := i + 1
			paramEnd := paramStart

			// Find the end of the parameter (next / or end of string)
			for paramEnd < len(parts) && parts[paramEnd] != '/' {
				paramEnd++
			}

			if paramEnd > paramStart {
				paramName := string(parts[paramStart:paramEnd])
				params = append(params, paramName)
			}

			i = paramEnd - 1 // Move to the end of this parameter
		}
	}

	return params
}

func valuesEqual(a, b vm.Value) bool {
	switch av := a.(type) {
	case vm.IntValue:
		if bv, ok := b.(vm.IntValue); ok {
			return av.Val == bv.Val
		}
	case vm.FloatValue:
		if bv, ok := b.(vm.FloatValue); ok {
			return av.Val == bv.Val
		}
	case vm.BoolValue:
		if bv, ok := b.(vm.BoolValue); ok {
			return av.Val == bv.Val
		}
	case vm.StringValue:
		if bv, ok := b.(vm.StringValue); ok {
			return av.Val == bv.Val
		}
	case vm.NullValue:
		_, ok := b.(vm.NullValue)
		return ok
	}
	return false
}

// compileMatchExpr compiles a match expression
// Match expressions are compiled as a series of conditional jumps similar to switch
func (c *Compiler) compileMatchExpr(expr *interpreter.MatchExpr) error {
	// Create a temporary variable to store the match value
	matchVarName := fmt.Sprintf("__match_%d", c.labelCounter)
	c.labelCounter++
	matchVarIdx := c.addConstant(vm.StringValue{Val: matchVarName})
	c.symbolTable.Define(matchVarName, matchVarIdx)

	// Compile and store the match value
	if err := c.compileExpression(expr.Value); err != nil {
		return err
	}
	c.emitWithOperand(vm.OpStoreVar, uint32(matchVarIdx))

	// Track jump locations for each case to jump to end
	var jumpToEnd []int

	// Compile each case
	for _, matchCase := range expr.Cases {
		// Compile pattern matching for this case
		jumpToNextCase, err := c.compilePatternMatch(matchCase.Pattern, matchVarIdx)
		if err != nil {
			return err
		}

		// If there's a guard, compile it as an additional condition
		if matchCase.Guard != nil {
			// Compile guard expression
			if err := c.compileExpression(matchCase.Guard); err != nil {
				return err
			}
			// Jump to next case if guard is false
			guardJump := len(c.code)
			c.emitWithOperand(vm.OpJumpIfFalse, 0) // Placeholder
			jumpToNextCase = append(jumpToNextCase, guardJump)
		}

		// Compile case body (this pushes the result onto the stack)
		if err := c.compileExpression(matchCase.Body); err != nil {
			return err
		}

		// Jump to end after executing case
		jumpToEnd = append(jumpToEnd, len(c.code))
		c.emitWithOperand(vm.OpJump, 0) // Placeholder

		// Patch all jumps to next case
		nextCaseOffset := uint32(len(c.code))
		for _, jumpLoc := range jumpToNextCase {
			c.patchJump(jumpLoc, nextCaseOffset)
		}
	}

	// If no case matched, push null
	nullIdx := c.addConstant(vm.NullValue{})
	c.emitWithOperand(vm.OpPush, uint32(nullIdx))

	// Patch all jumps to end
	endOffset := uint32(len(c.code))
	for _, jumpLoc := range jumpToEnd {
		c.patchJump(jumpLoc, endOffset)
	}

	return nil
}

// compilePatternMatch compiles pattern matching logic
// Returns a slice of jump locations that should jump to the next case
func (c *Compiler) compilePatternMatch(pattern interpreter.Pattern, matchVarIdx int) ([]int, error) {
	var jumpToNextCase []int

	switch p := pattern.(type) {
	case interpreter.LiteralPattern:
		// Load match value
		c.emitWithOperand(vm.OpLoadVar, uint32(matchVarIdx))
		// Push the literal
		if err := c.compileLiteralValue(p.Value); err != nil {
			return nil, err
		}
		// Compare
		c.emit(vm.OpEq)
		// Jump to next case if false
		jumpLoc := len(c.code)
		c.emitWithOperand(vm.OpJumpIfFalse, 0)
		jumpToNextCase = append(jumpToNextCase, jumpLoc)

	case interpreter.VariablePattern:
		// Variable pattern always matches, just bind the value
		c.emitWithOperand(vm.OpLoadVar, uint32(matchVarIdx))
		varIdx := c.addConstant(vm.StringValue{Val: p.Name})
		c.symbolTable.Define(p.Name, varIdx)
		c.emitWithOperand(vm.OpStoreVar, uint32(varIdx))

	case interpreter.WildcardPattern:
		// Wildcard always matches, no code needed

	case interpreter.ObjectPattern:
		// For object patterns, we need to check if the value is an object
		// and if it has all the required fields
		for _, field := range p.Fields {
			// Load match value
			c.emitWithOperand(vm.OpLoadVar, uint32(matchVarIdx))
			// Push field name
			fieldIdx := c.addConstant(vm.StringValue{Val: field.Key})
			c.emitWithOperand(vm.OpPush, uint32(fieldIdx))
			// Get field (this will push null if field doesn't exist)
			c.emit(vm.OpGetField)

			if field.Pattern != nil {
				// Store in temp var and match nested pattern
				tempVarName := fmt.Sprintf("__field_%s_%d", field.Key, c.labelCounter)
				c.labelCounter++
				tempVarIdx := c.addConstant(vm.StringValue{Val: tempVarName})
				c.symbolTable.Define(tempVarName, tempVarIdx)
				c.emitWithOperand(vm.OpStoreVar, uint32(tempVarIdx))

				nestedJumps, err := c.compilePatternMatch(field.Pattern, tempVarIdx)
				if err != nil {
					return nil, err
				}
				jumpToNextCase = append(jumpToNextCase, nestedJumps...)
			} else {
				// Bind field value to field name as variable
				varIdx := c.addConstant(vm.StringValue{Val: field.Key})
				c.symbolTable.Define(field.Key, varIdx)
				c.emitWithOperand(vm.OpStoreVar, uint32(varIdx))
			}
		}

	case interpreter.ArrayPattern:
		// For array patterns, check length and match elements
		// This is a simplified implementation
		for idx, elemPattern := range p.Elements {
			// Load match value
			c.emitWithOperand(vm.OpLoadVar, uint32(matchVarIdx))
			// Push index
			idxConstIdx := c.addConstant(vm.IntValue{Val: int64(idx)})
			c.emitWithOperand(vm.OpPush, uint32(idxConstIdx))
			// Get element
			c.emit(vm.OpGetIndex)

			// Store in temp var and match nested pattern
			tempVarName := fmt.Sprintf("__elem_%d_%d", idx, c.labelCounter)
			c.labelCounter++
			tempVarIdx := c.addConstant(vm.StringValue{Val: tempVarName})
			c.symbolTable.Define(tempVarName, tempVarIdx)
			c.emitWithOperand(vm.OpStoreVar, uint32(tempVarIdx))

			nestedJumps, err := c.compilePatternMatch(elemPattern, tempVarIdx)
			if err != nil {
				return nil, err
			}
			jumpToNextCase = append(jumpToNextCase, nestedJumps...)
		}

		// Handle rest pattern (simplified - just binds remaining elements)
		if p.Rest != nil {
			// For simplicity, we'll just bind the whole array to rest
			// A full implementation would slice the array
			c.emitWithOperand(vm.OpLoadVar, uint32(matchVarIdx))
			restIdx := c.addConstant(vm.StringValue{Val: *p.Rest})
			c.symbolTable.Define(*p.Rest, restIdx)
			c.emitWithOperand(vm.OpStoreVar, uint32(restIdx))
		}

	default:
		return nil, fmt.Errorf("unsupported pattern type: %T", pattern)
	}

	return jumpToNextCase, nil
}

// compileLiteralValue compiles a literal value from a pattern
func (c *Compiler) compileLiteralValue(lit interpreter.Literal) error {
	var val vm.Value

	switch l := lit.(type) {
	case interpreter.IntLiteral:
		val = vm.IntValue{Val: l.Value}
	case interpreter.FloatLiteral:
		val = vm.FloatValue{Val: l.Value}
	case interpreter.StringLiteral:
		val = vm.StringValue{Val: l.Value}
	case interpreter.BoolLiteral:
		val = vm.BoolValue{Val: l.Value}
	case interpreter.NullLiteral:
		val = vm.NullValue{}
	default:
		return fmt.Errorf("unsupported literal type in pattern: %T", lit)
	}

	idx := c.addConstant(val)
	c.emitWithOperand(vm.OpPush, uint32(idx))
	return nil
}
