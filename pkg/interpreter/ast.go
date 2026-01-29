package interpreter

// Module represents the top-level AST node
type Module struct {
	Items []Item
}

// Item represents a top-level item (TypeDef, Route, or Function)
type Item interface {
	isItem()
}

// TypeDef represents a type definition
// Example: : Result<T, E> { ok: T?, error: E? }
type TypeDef struct {
	Name       string
	TypeParams []TypeParameter // Generic type parameters (e.g., T, E)
	Fields     []Field
}

func (TypeDef) isItem() {}

// Route represents an HTTP route
type Route struct {
	Path        string
	Method      HttpMethod
	InputType   Type // Type for request body validation (from < input: Type syntax)
	ReturnType  Type
	Auth        *AuthConfig
	RateLimit   *RateLimit
	Injections  []Injection
	QueryParams []QueryParamDecl
	Body        []Statement
}

func (Route) isItem() {}

// Function represents a function definition
// Example: ! map<T, U>(arr: [T], fn: (T) -> U): [U]
type Function struct {
	Name       string
	TypeParams []TypeParameter // Generic type parameters (e.g., T, U)
	Params     []Field
	ReturnType Type
	Body       []Statement
}

func (Function) isItem() {}

// Field represents a struct field
type Field struct {
	Name           string
	TypeAnnotation Type
	Required       bool
	Default        Expr // nil if no default value
}

// Type represents a type annotation
type Type interface {
	isType()
}

type IntType struct{}
type StringType struct{}
type BoolType struct{}
type FloatType struct{}

type ArrayType struct {
	ElementType Type
}

type OptionalType struct {
	InnerType Type
}

type NamedType struct {
	Name string
}

type DatabaseType struct{}
type RedisType struct{}

type UnionType struct {
	Types []Type
}

// TypeParameter represents a generic type parameter with optional constraint
// Example: T, T: Comparable, T extends Numeric
type TypeParameter struct {
	Name       string // The type parameter name (e.g., "T", "U")
	Constraint Type   // Optional constraint (trait bound), nil if unconstrained
}

// GenericType represents a generic type instantiation
// Example: List<int>, Map<string, User>, Result<T, E>
type GenericType struct {
	BaseType Type   // The base type (usually NamedType like "List" or "Map")
	TypeArgs []Type // The type arguments (e.g., [int] for List<int>)
}

// TypeParameterType represents a reference to a type parameter
// Used inside generic definitions to refer to type parameters
type TypeParameterType struct {
	Name string // The type parameter name (e.g., "T")
}

// FunctionType represents a function type signature
// Example: (T) -> U, (int, string) -> bool
type FunctionType struct {
	ParamTypes []Type // Parameter types
	ReturnType Type   // Return type
}

func (IntType) isType()           {}
func (StringType) isType()        {}
func (BoolType) isType()          {}
func (FloatType) isType()         {}
func (ArrayType) isType()         {}
func (OptionalType) isType()      {}
func (NamedType) isType()         {}
func (DatabaseType) isType()      {}
func (RedisType) isType()         {}
func (UnionType) isType()         {}
func (GenericType) isType()       {}
func (TypeParameterType) isType() {}
func (FunctionType) isType()      {}

// HttpMethod represents HTTP methods
type HttpMethod int

const (
	Get HttpMethod = iota
	Post
	Put
	Delete
	Patch
	WebSocket // WebSocket upgrade
)

func (m HttpMethod) String() string {
	switch m {
	case Get:
		return "GET"
	case Post:
		return "POST"
	case Put:
		return "PUT"
	case Delete:
		return "DELETE"
	case Patch:
		return "PATCH"
	case WebSocket:
		return "WS"
	default:
		return "UNKNOWN"
	}
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	AuthType string
	Required bool
}

// RateLimit represents rate limiting configuration
type RateLimit struct {
	Requests uint32
	Window   string
}

// Injection represents a dependency injection
type Injection struct {
	Name string
	Type Type
}

// QueryParamDecl represents a declared query parameter
type QueryParamDecl struct {
	Name     string
	Type     Type
	Required bool
	Default  Expr
	IsArray  bool
}

// Statement represents a statement in the AST
type Statement interface {
	isStatement()
	isNode()
}

// AssignStatement represents variable declaration with $ syntax
type AssignStatement struct {
	Target string
	Value  Expr
}

func (AssignStatement) isStatement() {}

// ReassignStatement represents variable reassignment without $ syntax
type ReassignStatement struct {
	Target string
	Value  Expr
}

func (ReassignStatement) isStatement() {}
func (ReassignStatement) isNode()      {}

// DbQueryStatement represents a database query
type DbQueryStatement struct {
	Var    string
	Query  string
	Params []Expr
}

func (DbQueryStatement) isStatement() {}

// ReturnStatement represents a return statement
type ReturnStatement struct {
	Value Expr
}

func (ReturnStatement) isStatement() {}

// IfStatement represents an if statement
type IfStatement struct {
	Condition Expr
	ThenBlock []Statement
	ElseBlock []Statement
}

func (IfStatement) isStatement() {}

// WhileStatement represents a while loop
type WhileStatement struct {
	Condition Expr
	Body      []Statement
}

func (WhileStatement) isStatement() {}

// SwitchStatement represents a switch statement
type SwitchStatement struct {
	Value   Expr
	Cases   []SwitchCase
	Default []Statement
}

func (SwitchStatement) isStatement() {}

// SwitchCase represents a single case in a switch statement
type SwitchCase struct {
	Value Expr
	Body  []Statement
}

// ForStatement represents a for loop
type ForStatement struct {
	KeyVar   string // Optional: variable name for key/index (empty if not used)
	ValueVar string // Variable name for value/element
	Iterable Expr   // Expression that evaluates to array or object
	Body     []Statement
}

func (ForStatement) isStatement() {}

// WebSocket-specific statements

// WsSendStatement represents sending a message to a specific client
type WsSendStatement struct {
	Client  Expr // Client ID or connection
	Message Expr // Message to send
}

func (WsSendStatement) isStatement() {}

// WsBroadcastStatement represents broadcasting a message to all clients
type WsBroadcastStatement struct {
	Message Expr  // Message to broadcast
	Except  *Expr // Optional: client ID to exclude
}

func (WsBroadcastStatement) isStatement() {}

// WsCloseStatement represents closing a WebSocket connection
type WsCloseStatement struct {
	Client Expr // Client ID or connection
	Reason Expr // Optional reason string
}

func (WsCloseStatement) isStatement() {}

// ValidationStatement represents a validation check: ? validate_fn(args)
type ValidationStatement struct {
	Call FunctionCallExpr // The validation function call
}

func (ValidationStatement) isStatement() {}

// ExpressionStatement represents an expression used as a statement (e.g., function call)
type ExpressionStatement struct {
	Expr Expr
}

func (ExpressionStatement) isStatement() {}

// Expr represents an expression in the AST
type Expr interface {
	isExpr()
}

// LiteralExpr represents a literal value
type LiteralExpr struct {
	Value Literal
}

func (LiteralExpr) isExpr() {}

// VariableExpr represents a variable reference
type VariableExpr struct {
	Name string
}

func (VariableExpr) isExpr() {}

// BinaryOpExpr represents a binary operation
type BinaryOpExpr struct {
	Op    BinOp
	Left  Expr
	Right Expr
}

func (BinaryOpExpr) isExpr() {}

// UnaryOpExpr represents a unary operation (e.g., !expr, -expr)
type UnaryOpExpr struct {
	Op    UnOp
	Right Expr
}

func (UnaryOpExpr) isExpr() {}

// FieldAccessExpr represents field access
type FieldAccessExpr struct {
	Object Expr
	Field  string
}

func (FieldAccessExpr) isExpr() {}

// ArrayIndexExpr represents array indexing: array[index]
type ArrayIndexExpr struct {
	Array Expr
	Index Expr
}

func (ArrayIndexExpr) isExpr() {}

// FunctionCallExpr represents a function call
// Example: map<int, string>(arr, fn)
type FunctionCallExpr struct {
	Name     string
	TypeArgs []Type // Type arguments for generic function calls (e.g., <int, string>)
	Args     []Expr
}

func (FunctionCallExpr) isExpr() {}

// ObjectExpr represents an object literal
type ObjectExpr struct {
	Fields []ObjectField
}

func (ObjectExpr) isExpr() {}

// ObjectField represents a field in an object literal
type ObjectField struct {
	Key   string
	Value Expr
}

// ArrayExpr represents an array literal
type ArrayExpr struct {
	Elements []Expr
}

func (ArrayExpr) isExpr() {}

// LambdaExpr represents an anonymous function (lambda/arrow function)
// Example: (x) => x * 2, (a, b) => a + b
type LambdaExpr struct {
	Params []Field     // Parameter list
	Body   Expr        // Single expression body (for short lambdas)
	Block  []Statement // Statement body (for multi-line lambdas), mutually exclusive with Body
}

func (LambdaExpr) isExpr() {}

// PipeExpr represents a pipe expression: left |> right
// The left value is piped as the first argument to the right function
type PipeExpr struct {
	Left  Expr // The value being piped
	Right Expr // The function/call to receive the piped value
}

func (PipeExpr) isExpr() {}

// Literal represents a literal value
type Literal interface {
	isLiteral()
}

type IntLiteral struct {
	Value int64
}

type StringLiteral struct {
	Value string
}

type BoolLiteral struct {
	Value bool
}

type FloatLiteral struct {
	Value float64
}

// NullLiteral represents a null value
type NullLiteral struct{}

func (IntLiteral) isLiteral()    {}
func (StringLiteral) isLiteral() {}
func (BoolLiteral) isLiteral()   {}
func (FloatLiteral) isLiteral()  {}
func (NullLiteral) isLiteral()   {}

// BinOp represents binary operators
type BinOp int

const (
	Add BinOp = iota
	Sub
	Mul
	Div
	Eq
	Ne
	Lt
	Le
	Gt
	Ge
	And
	Or
)

func (op BinOp) String() string {
	switch op {
	case Add:
		return "+"
	case Sub:
		return "-"
	case Mul:
		return "*"
	case Div:
		return "/"
	case Eq:
		return "=="
	case Ne:
		return "!="
	case Lt:
		return "<"
	case Le:
		return "<="
	case Gt:
		return ">"
	case Ge:
		return ">="
	case And:
		return "&&"
	case Or:
		return "||"
	default:
		return "UNKNOWN"
	}
}

// UnOp represents unary operators
type UnOp int

const (
	Not UnOp = iota // Logical NOT (!)
	Neg             // Unary minus (-)
)

func (op UnOp) String() string {
	switch op {
	case Not:
		return "!"
	case Neg:
		return "-"
	default:
		return "UNKNOWN"
	}
}

// WebSocketEventType represents WebSocket event types
type WebSocketEventType int

const (
	WSEventConnect WebSocketEventType = iota
	WSEventDisconnect
	WSEventMessage
	WSEventError
)

// WebSocketEvent represents a WebSocket event handler
type WebSocketEvent struct {
	EventType WebSocketEventType
	Body      []Statement
}

func (WebSocketEvent) isStatement() {}

// WebSocketRoute represents a WebSocket route
type WebSocketRoute struct {
	Path   string
	Events []WebSocketEvent
}

func (WebSocketRoute) isItem() {}

// Command represents a CLI command definition
// Example: @ command hello name: str! greeting: str = "Hello"
type Command struct {
	Name        string
	Description string
	Params      []CommandParam
	ReturnType  Type
	Body        []Statement
}

func (Command) isItem() {}

// CommandParam represents a CLI command parameter with optional default
type CommandParam struct {
	Name     string
	Type     Type
	Required bool
	Default  Expr // nil if no default
	IsFlag   bool // true for --flag style args
}

// CronTask represents a scheduled task
// Example: @ cron "0 0 * * *"
type CronTask struct {
	Name       string // optional name for the task
	Schedule   string // cron expression
	Timezone   string // optional timezone (default UTC)
	Retries    int    // number of retries on failure
	Injections []Injection
	Body       []Statement
}

func (CronTask) isItem() {}

// EventHandler represents an event handler
// Example: @ event "user.created"
type EventHandler struct {
	EventType  string // e.g., "user.created", "order.paid"
	Async      bool   // whether to handle asynchronously
	Injections []Injection
	Body       []Statement
}

func (EventHandler) isItem() {}

// QueueWorker represents a message queue worker
// Example: @ queue "email.send"
type QueueWorker struct {
	QueueName   string
	Concurrency int // number of concurrent workers
	MaxRetries  int
	Timeout     int // timeout in seconds
	Injections  []Injection
	Body        []Statement
}

func (QueueWorker) isItem() {}

// ImportStatement represents an import declaration
// Syntax forms:
//
//	import "path/to/file"
//	import "path/to/file" as alias
//	from "path/to/file" import { name1, name2 }
//	from "path/to/file" import { name1 as alias1, name2 }
type ImportStatement struct {
	Path      string       // The import path (relative or package name)
	Alias     string       // Optional alias for the entire module
	Selective bool         // True if using selective imports (from ... import { ... })
	Names     []ImportName // For selective imports: the names to import
}

func (ImportStatement) isItem() {}

// ImportName represents a single name in a selective import
type ImportName struct {
	Name  string // The original name in the module
	Alias string // Optional alias (empty if not aliased)
}

// ModuleDecl represents a module namespace declaration
// Syntax: module "name"
type ModuleDecl struct {
	Name string // The module namespace name
}

func (ModuleDecl) isItem() {}

// ConstDecl represents a module-level constant declaration.
// Constants are immutable bindings evaluated at module load time.
// Syntax: const NAME = value or const NAME: Type = value
type ConstDecl struct {
	Name  string // The constant name
	Value Expr   // The constant value expression
	Type  Type   // Optional type annotation (nil if type is inferred)
}

func (ConstDecl) isItem() {}

// AsyncExpr represents an async block expression: async { ... }
// The block is executed asynchronously and returns a Future
type AsyncExpr struct {
	Body []Statement
}

func (AsyncExpr) isExpr() {}

// AwaitExpr represents awaiting a Future: await future
type AwaitExpr struct {
	Expr Expr // Expression that evaluates to a Future
}

func (AwaitExpr) isExpr() {}

// FutureType represents the type of a Future value
type FutureType struct {
	ResultType Type // The type of value the future will resolve to
}

func (FutureType) isType() {}

// MatchExpr represents a pattern matching expression
// Example: match value { pattern => result, _ => default }
type MatchExpr struct {
	Value Expr
	Cases []MatchCase
}

func (MatchExpr) isExpr() {}

// MatchCase represents a single case in a match expression
type MatchCase struct {
	Pattern Pattern
	Guard   Expr // Optional guard condition (when clause)
	Body    Expr // The result expression
}

// Pattern represents a pattern for matching
type Pattern interface {
	isPattern()
}

// LiteralPattern matches a literal value (int, string, bool, etc.)
type LiteralPattern struct {
	Value Literal
}

func (LiteralPattern) isPattern() {}

// VariablePattern binds the matched value to a variable
type VariablePattern struct {
	Name string
}

func (VariablePattern) isPattern() {}

// WildcardPattern matches anything (underscore _)
type WildcardPattern struct{}

func (WildcardPattern) isPattern() {}

// ObjectPattern matches and destructures an object
// Example: {name, age} or {name: n, age: a}
type ObjectPattern struct {
	Fields []ObjectPatternField
}

func (ObjectPattern) isPattern() {}

// ObjectPatternField represents a field in an object pattern
type ObjectPatternField struct {
	Key     string  // The key to match in the object
	Pattern Pattern // The pattern to bind (nil means use Key as variable name)
}

// ArrayPattern matches and destructures an array
// Example: [first, second] or [head, ...rest]
type ArrayPattern struct {
	Elements []Pattern
	Rest     *string // Optional rest variable name for [...rest] syntax
}

func (ArrayPattern) isPattern() {}

// MacroDef represents a macro definition
// Example: macro! log(level, msg) { ... }
type MacroDef struct {
	Name   string   // Macro name
	Params []string // Parameter names
	Body   []Node   // Body is a template of AST nodes
}

func (MacroDef) isItem() {}

// MacroInvocation represents a macro call
// Example: log!("INFO", "message")
type MacroInvocation struct {
	Name string // Macro name (without !)
	Args []Expr // Arguments to substitute
}

func (MacroInvocation) isItem()      {}
func (MacroInvocation) isStatement() {}
func (MacroInvocation) isExpr()      {}

// QuoteExpr represents an unevaluated AST fragment
// Example: quote { if x > 0 { return x } }
type QuoteExpr struct {
	Body []Node // The AST nodes being quoted
}

func (QuoteExpr) isExpr() {}

// UnquoteExpr represents a value to be spliced into quoted AST
// Example: $expr (inside a quote block)
type UnquoteExpr struct {
	Expr Expr // Expression to evaluate and splice
}

func (UnquoteExpr) isExpr() {}

// Node is a generic AST node interface that can be Item, Statement, or Expr
type Node interface {
	isNode()
}

// Make existing types implement Node
func (TypeDef) isNode()              {}
func (Route) isNode()                {}
func (Function) isNode()             {}
func (WebSocketRoute) isNode()       {}
func (Command) isNode()              {}
func (CronTask) isNode()             {}
func (EventHandler) isNode()         {}
func (QueueWorker) isNode()          {}
func (ImportStatement) isNode()      {}
func (ModuleDecl) isNode()           {}
func (ConstDecl) isNode()            {}
func (MacroDef) isNode()             {}
func (MacroInvocation) isNode()      {}
func (AssignStatement) isNode()      {}
func (DbQueryStatement) isNode()     {}
func (ReturnStatement) isNode()      {}
func (IfStatement) isNode()          {}
func (WhileStatement) isNode()       {}
func (SwitchStatement) isNode()      {}
func (ForStatement) isNode()         {}
func (WsSendStatement) isNode()      {}
func (WsBroadcastStatement) isNode() {}
func (WsCloseStatement) isNode()     {}
func (ValidationStatement) isNode()  {}
func (ExpressionStatement) isNode()  {}
func (WebSocketEvent) isNode()       {}
func (LiteralExpr) isNode()          {}
func (VariableExpr) isNode()         {}
func (BinaryOpExpr) isNode()         {}
func (UnaryOpExpr) isNode()          {}
func (FieldAccessExpr) isNode()      {}
func (ArrayIndexExpr) isNode()       {}
func (FunctionCallExpr) isNode()     {}
func (ObjectExpr) isNode()           {}
func (ArrayExpr) isNode()            {}
func (QuoteExpr) isNode()            {}
func (UnquoteExpr) isNode()          {}
func (MatchExpr) isNode()            {}
func (LiteralPattern) isNode()       {}
func (VariablePattern) isNode()      {}
func (WildcardPattern) isNode()      {}
func (ObjectPattern) isNode()        {}
func (ArrayPattern) isNode()         {}
func (AsyncExpr) isNode()            {}
func (AwaitExpr) isNode()            {}
func (LambdaExpr) isNode()           {}
