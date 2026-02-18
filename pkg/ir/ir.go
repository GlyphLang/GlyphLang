// Package ir defines the Semantic Intermediate Representation for GlyphLang.
//
// The IR is a language-neutral, normalized representation of a GlyphLang program
// that sits between the AST and any target backend (interpreter, code generator, etc.).
// It captures intent and semantics without language-specific implementation details.
package ir

// ServiceIR is the top-level IR node representing a complete service definition.
// It contains all routes, types, providers, background tasks, and metadata
// needed to generate or execute a service in any target language.
type ServiceIR struct {
	Name      string
	Types     []TypeSchema
	Providers []ProviderRef
	Routes    []RouteHandler
	Events    []EventBinding
	CronJobs  []CronBinding
	Queues    []QueueBinding
	Commands  []CommandDef
	Functions []FunctionDef
	GRPC      []GRPCServiceDef
	GraphQL   []GraphQLDef
	WebSocket []WebSocketDef
	Constants []ConstantDef
}

// TypeSchema describes a type definition in the IR.
// It is target-neutral: no Go, Python, or Java-specific semantics.
type TypeSchema struct {
	Name       string
	Fields     []FieldSchema
	TypeParams []string
	Traits     []string
	Methods    []MethodSchema
}

// FieldSchema describes a single field within a TypeSchema.
type FieldSchema struct {
	Name        string
	Type        TypeRef
	Required    bool
	HasDefault  bool
	Default     ExprIR
	Annotations []Annotation
}

// MethodSchema describes a method on a type.
type MethodSchema struct {
	Name       string
	Params     []FieldSchema
	ReturnType TypeRef
	Body       []StmtIR
}

// Annotation represents a declarative annotation on a field (e.g., @email, @minLen(2)).
type Annotation struct {
	Name   string
	Params []interface{}
}

// TypeRef is a target-neutral type reference.
type TypeRef struct {
	Kind     TypeKind
	Name     string    // For Named, Provider kinds
	Inner    *TypeRef  // For Array, Optional, Future kinds
	Elements []TypeRef // For Union, Generic type args
}

// TypeKind classifies the shape of a type reference.
type TypeKind int

const (
	TypeInt TypeKind = iota
	TypeFloat
	TypeString
	TypeBool
	TypeArray
	TypeOptional
	TypeNamed
	TypeProvider
	TypeUnion
	TypeGeneric
	TypeFunction
	TypeFuture
	TypeAny
)

// ProviderRef describes a provider dependency required by the service.
// This is the generalized form of Database, Redis, MongoDB, LLM, etc.
type ProviderRef struct {
	Name         string      // Instance name as used in code (e.g., "db")
	ProviderType string      // Provider type name (e.g., "Database", "Redis", "ImageProcessor")
	IsStandard   bool        // True for built-in providers (Database, Redis, MongoDB, LLM)
	Methods      []MethodSig // Known methods on this provider (from contract)
}

// MethodSig describes a method signature on a provider contract.
type MethodSig struct {
	Name       string
	Params     []FieldSchema
	ReturnType TypeRef
}

// RouteHandler describes an HTTP route in the IR.
type RouteHandler struct {
	Method      HTTPMethod
	Path        string
	PathParams  []string
	QueryParams []QueryParam
	InputType   *TypeRef
	ReturnType  *TypeRef
	Auth        *AuthRequirement
	RateLimit   *RateLimitConfig
	Middleware  []MiddlewareRef
	Providers   []InjectionRef
	Body        []StmtIR
}

// HTTPMethod represents an HTTP method.
type HTTPMethod int

const (
	MethodGet HTTPMethod = iota
	MethodPost
	MethodPut
	MethodDelete
	MethodPatch
	MethodWebSocket
	MethodSSE
)

// String returns the HTTP method as an uppercase string.
func (m HTTPMethod) String() string {
	switch m {
	case MethodGet:
		return "GET"
	case MethodPost:
		return "POST"
	case MethodPut:
		return "PUT"
	case MethodDelete:
		return "DELETE"
	case MethodPatch:
		return "PATCH"
	case MethodWebSocket:
		return "WS"
	case MethodSSE:
		return "SSE"
	default:
		return "UNKNOWN"
	}
}

// QueryParam describes a declared query parameter.
type QueryParam struct {
	Name     string
	Type     TypeRef
	Required bool
	Default  ExprIR
	IsArray  bool
}

// AuthRequirement describes the authentication needed for a route.
type AuthRequirement struct {
	AuthType string // e.g., "jwt", "apikey", "basic"
	Required bool
	Roles    []string
}

// RateLimitConfig describes rate limiting for a route.
type RateLimitConfig struct {
	Requests uint32
	Window   string
}

// MiddlewareRef is a reference to a named middleware with optional arguments.
type MiddlewareRef struct {
	Name string
	Args []ExprIR
}

// InjectionRef describes a provider injection into a handler.
type InjectionRef struct {
	Name         string // Local variable name (e.g., "db")
	ProviderType string // Provider type name (e.g., "Database")
}

// EventBinding describes an event handler.
type EventBinding struct {
	EventType string
	Async     bool
	Providers []InjectionRef
	Body      []StmtIR
}

// CronBinding describes a scheduled task.
type CronBinding struct {
	Name      string
	Schedule  string
	Timezone  string
	Retries   int
	Providers []InjectionRef
	Body      []StmtIR
}

// QueueBinding describes a queue worker.
type QueueBinding struct {
	QueueName   string
	Concurrency int
	MaxRetries  int
	Timeout     int
	Providers   []InjectionRef
	Body        []StmtIR
}

// CommandDef describes a CLI command.
type CommandDef struct {
	Name        string
	Description string
	Params      []CommandParam
	ReturnType  *TypeRef
	Body        []StmtIR
}

// CommandParam describes a CLI command parameter.
type CommandParam struct {
	Name     string
	Type     TypeRef
	Required bool
	Default  ExprIR
	IsFlag   bool
}

// FunctionDef describes a standalone function.
type FunctionDef struct {
	Name       string
	TypeParams []string
	Params     []FieldSchema
	ReturnType *TypeRef
	Body       []StmtIR
}

// GRPCServiceDef describes a gRPC service definition.
type GRPCServiceDef struct {
	Name     string
	Methods  []GRPCMethodDef
	Handlers []GRPCHandlerDef
}

// GRPCMethodDef describes a gRPC method signature.
type GRPCMethodDef struct {
	Name       string
	InputType  TypeRef
	ReturnType TypeRef
	StreamType GRPCStreamType
}

// GRPCStreamType indicates gRPC streaming mode.
type GRPCStreamType int

const (
	GRPCUnary GRPCStreamType = iota
	GRPCServerStream
	GRPCClientStream
	GRPCBidirectional
)

// GRPCHandlerDef describes a gRPC handler implementation.
type GRPCHandlerDef struct {
	ServiceName string
	MethodName  string
	Params      []FieldSchema
	ReturnType  *TypeRef
	StreamType  GRPCStreamType
	Auth        *AuthRequirement
	Providers   []InjectionRef
	Body        []StmtIR
}

// GraphQLDef describes a GraphQL resolver.
type GraphQLDef struct {
	Operation  GraphQLOp
	FieldName  string
	Params     []FieldSchema
	ReturnType *TypeRef
	Auth       *AuthRequirement
	Providers  []InjectionRef
	Body       []StmtIR
}

// GraphQLOp is the GraphQL operation type.
type GraphQLOp int

const (
	GraphQLQuery GraphQLOp = iota
	GraphQLMutation
	GraphQLSubscription
)

// WebSocketDef describes a WebSocket route.
type WebSocketDef struct {
	Path   string
	Events []WSEventDef
}

// WSEventDef describes a WebSocket event handler.
type WSEventDef struct {
	EventType WSEventType
	Body      []StmtIR
}

// WSEventType identifies a WebSocket event.
type WSEventType int

const (
	WSConnect WSEventType = iota
	WSDisconnect
	WSMessage
	WSError
)

// ConstantDef describes a module-level constant.
type ConstantDef struct {
	Name  string
	Type  *TypeRef
	Value ExprIR
}

// StmtIR represents a statement in the IR.
type StmtIR struct {
	Kind     StmtKind
	Assign   *AssignStmt
	Return   *ReturnStmt
	If       *IfStmt
	For      *ForStmt
	While    *WhileStmt
	Switch   *SwitchStmt
	ExprStmt *ExprIR
	Validate *ValidateStmt
	Break    bool
	Continue bool
}

// StmtKind classifies the type of statement.
type StmtKind int

const (
	StmtAssign StmtKind = iota
	StmtReassign
	StmtReturn
	StmtIf
	StmtFor
	StmtWhile
	StmtSwitch
	StmtExpr
	StmtValidate
	StmtBreak
	StmtContinue
)

// AssignStmt describes a variable assignment.
type AssignStmt struct {
	Target string
	Value  ExprIR
}

// ReturnStmt describes a return statement.
type ReturnStmt struct {
	Value ExprIR
}

// IfStmt describes an if/else statement.
type IfStmt struct {
	Condition ExprIR
	Then      []StmtIR
	Else      []StmtIR
}

// ForStmt describes a for loop.
type ForStmt struct {
	KeyVar   string
	ValueVar string
	Iterable ExprIR
	Body     []StmtIR
}

// WhileStmt describes a while loop.
type WhileStmt struct {
	Condition ExprIR
	Body      []StmtIR
}

// SwitchStmt describes a switch statement.
type SwitchStmt struct {
	Value   ExprIR
	Cases   []SwitchCase
	Default []StmtIR
}

// SwitchCase is a single case in a switch statement.
type SwitchCase struct {
	Value ExprIR
	Body  []StmtIR
}

// ValidateStmt describes a validation check.
type ValidateStmt struct {
	Call ExprIR
}

// ExprIR represents an expression in the IR.
type ExprIR struct {
	Kind        ExprKind
	IntVal      int64
	FloatVal    float64
	StringVal   string
	BoolVal     bool
	IsNull      bool
	VarName     string
	BinOp       *BinaryExpr
	UnaryOp     *UnaryExpr
	FieldAccess *FieldAccessExpr
	IndexAccess *IndexAccessExpr
	Call        *CallExpr
	Object      *ObjectExpr
	Array       *ArrayExpr
	Lambda      *LambdaExpr
	Pipe        *PipeExpr
	Match       *MatchExpr
	Async       *AsyncExprIR
	Await       *AwaitExprIR
}

// ExprKind classifies the type of expression.
type ExprKind int

const (
	ExprInt ExprKind = iota
	ExprFloat
	ExprString
	ExprBool
	ExprNull
	ExprVar
	ExprBinary
	ExprUnary
	ExprFieldAccess
	ExprIndexAccess
	ExprCall
	ExprObject
	ExprArray
	ExprLambda
	ExprPipe
	ExprMatch
	ExprAsync
	ExprAwait
)

// BinaryExpr describes a binary operation.
type BinaryExpr struct {
	Op    BinOp
	Left  ExprIR
	Right ExprIR
}

// BinOp identifies a binary operator.
type BinOp int

const (
	OpAdd BinOp = iota
	OpSub
	OpMul
	OpDiv
	OpMod
	OpEq
	OpNe
	OpLt
	OpLe
	OpGt
	OpGe
	OpAnd
	OpOr
)

// UnaryExpr describes a unary operation.
type UnaryExpr struct {
	Op    UnOp
	Right ExprIR
}

// UnOp identifies a unary operator.
type UnOp int

const (
	OpNot UnOp = iota
	OpNeg
)

// FieldAccessExpr describes field access (obj.field).
type FieldAccessExpr struct {
	Object ExprIR
	Field  string
}

// IndexAccessExpr describes index access (arr[idx]).
type IndexAccessExpr struct {
	Object ExprIR
	Index  ExprIR
}

// CallExpr describes a function or method call.
type CallExpr struct {
	Name     string
	TypeArgs []TypeRef
	Args     []ExprIR
}

// ObjectExpr describes an object literal.
type ObjectExpr struct {
	Fields []ObjectFieldIR
}

// ObjectFieldIR is a field in an object literal.
type ObjectFieldIR struct {
	Key   string
	Value ExprIR
}

// ArrayExpr describes an array literal.
type ArrayExpr struct {
	Elements []ExprIR
}

// LambdaExpr describes a lambda/arrow function.
type LambdaExpr struct {
	Params []FieldSchema
	Body   ExprIR   // Single-expression body
	Block  []StmtIR // Multi-statement body (mutually exclusive with Body)
}

// PipeExpr describes a pipe operation (left |> right).
type PipeExpr struct {
	Left  ExprIR
	Right ExprIR
}

// AsyncExprIR describes an async block.
type AsyncExprIR struct {
	Body []StmtIR
}

// AwaitExprIR describes an await expression.
type AwaitExprIR struct {
	Expr ExprIR
}

// MatchExpr describes a pattern match.
type MatchExpr struct {
	Value ExprIR
	Cases []MatchCase
}

// MatchCase is a single case in a match expression.
type MatchCase struct {
	Pattern PatternIR
	Guard   *ExprIR
	Body    ExprIR
}

// PatternIR describes a match pattern.
type PatternIR struct {
	Kind     PatternKind
	IntVal   int64
	FloatVal float64
	StrVal   string
	BoolVal  bool
	VarName  string
	Fields   []ObjectPatternField
	Elements []PatternIR
	RestVar  string
}

// PatternKind classifies the type of pattern.
type PatternKind int

const (
	PatternLiteral PatternKind = iota
	PatternVariable
	PatternWildcard
	PatternObject
	PatternArray
)

// ObjectPatternField is a field in an object destructuring pattern.
type ObjectPatternField struct {
	Key     string
	Pattern PatternIR
}
