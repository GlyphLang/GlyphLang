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
type TypeDef struct {
	Name   string
	Fields []Field
}

func (TypeDef) isItem() {}

// Route represents an HTTP route
type Route struct {
	Path       string
	Method     HttpMethod
	ReturnType Type
	Auth       *AuthConfig
	RateLimit  *RateLimit
	Injections []Injection
	Body       []Statement
}

func (Route) isItem() {}

// Function represents a function definition
type Function struct {
	Name       string
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

type UnionType struct {
	Types []Type
}

func (IntType) isType()      {}
func (StringType) isType()   {}
func (BoolType) isType()     {}
func (FloatType) isType()    {}
func (ArrayType) isType()    {}
func (OptionalType) isType() {}
func (NamedType) isType()    {}
func (DatabaseType) isType() {}
func (UnionType) isType()    {}

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

// Statement represents a statement in the AST
type Statement interface {
	isStatement()
}

// AssignStatement represents variable assignment
type AssignStatement struct {
	Target string
	Value  Expr
}

func (AssignStatement) isStatement() {}

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
	Message Expr      // Message to broadcast
	Except  *Expr     // Optional: client ID to exclude
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
type FunctionCallExpr struct {
	Name string
	Args []Expr
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
