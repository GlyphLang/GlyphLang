// Package ir defines the Semantic Intermediate Representation for GlyphLang.
//
// The IR is a language-neutral, normalized representation of a GlyphLang program
// that sits between the AST and any target backend (interpreter, code generator, etc.).
// It captures intent and semantics without language-specific implementation details.
package ir

import "encoding/json"

// ServiceIR is the top-level IR node representing a complete service definition.
// It contains all routes, types, providers, background tasks, and metadata
// needed to generate or execute a service in any target language.
type ServiceIR struct {
	Name      string           `json:"name"`
	Types     []TypeSchema     `json:"types,omitempty"`
	Providers []ProviderRef    `json:"providers,omitempty"`
	Routes    []RouteHandler   `json:"routes,omitempty"`
	Events    []EventBinding   `json:"events,omitempty"`
	CronJobs  []CronBinding    `json:"cron_jobs,omitempty"`
	Queues    []QueueBinding   `json:"queues,omitempty"`
	Commands  []CommandDef     `json:"commands,omitempty"`
	Functions []FunctionDef    `json:"functions,omitempty"`
	GRPC      []GRPCServiceDef `json:"grpc,omitempty"`
	GraphQL   []GraphQLDef     `json:"graphql,omitempty"`
	WebSocket []WebSocketDef   `json:"websocket,omitempty"`
	Constants []ConstantDef    `json:"constants,omitempty"`
}

// TypeSchema describes a type definition in the IR.
// It is target-neutral: no Go, Python, or Java-specific semantics.
type TypeSchema struct {
	Name       string         `json:"name"`
	Fields     []FieldSchema  `json:"fields,omitempty"`
	TypeParams []string       `json:"type_params,omitempty"`
	Traits     []string       `json:"traits,omitempty"`
	Methods    []MethodSchema `json:"methods,omitempty"`
}

// FieldSchema describes a single field within a TypeSchema.
type FieldSchema struct {
	Name        string       `json:"name"`
	Type        TypeRef      `json:"type"`
	Required    bool         `json:"required"`
	HasDefault  bool         `json:"has_default"`
	Default     ExprIR       `json:"-"`
	Annotations []Annotation `json:"annotations,omitempty"`
}

// MethodSchema describes a method on a type.
type MethodSchema struct {
	Name       string        `json:"name"`
	Params     []FieldSchema `json:"params,omitempty"`
	ReturnType TypeRef       `json:"return_type"`
	Body       []StmtIR      `json:"-"`
}

// Annotation represents a declarative annotation on a field (e.g., @email, @minLen(2)).
type Annotation struct {
	Name   string        `json:"name"`
	Params []interface{} `json:"params,omitempty"`
}

// TypeRef is a target-neutral type reference.
type TypeRef struct {
	Kind     TypeKind  `json:"kind"`
	Name     string    `json:"name,omitempty"`     // For Named, Provider kinds
	Inner    *TypeRef  `json:"inner,omitempty"`    // For Array, Optional, Future kinds
	Elements []TypeRef `json:"elements,omitempty"` // For Union, Generic type args
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

// String returns the TypeKind as a human-readable string.
func (k TypeKind) String() string {
	switch k {
	case TypeInt:
		return "int"
	case TypeFloat:
		return "float"
	case TypeString:
		return "string"
	case TypeBool:
		return "bool"
	case TypeArray:
		return "array"
	case TypeOptional:
		return "optional"
	case TypeNamed:
		return "named"
	case TypeProvider:
		return "provider"
	case TypeUnion:
		return "union"
	case TypeGeneric:
		return "generic"
	case TypeFunction:
		return "function"
	case TypeFuture:
		return "future"
	case TypeAny:
		return "any"
	default:
		return "unknown"
	}
}

// MarshalJSON serializes TypeKind as a human-readable string.
func (k TypeKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.String())
}

// ProviderRef describes a provider dependency required by the service.
// This is the generalized form of Database, Redis, MongoDB, LLM, etc.
type ProviderRef struct {
	Name         string      `json:"name"`              // Instance name as used in code (e.g., "db")
	ProviderType string      `json:"provider_type"`     // Provider type name (e.g., "Database", "Redis", "ImageProcessor")
	IsStandard   bool        `json:"is_standard"`       // True for built-in providers (Database, Redis, MongoDB, LLM)
	Methods      []MethodSig `json:"methods,omitempty"` // Known methods on this provider (from contract)
}

// MethodSig describes a method signature on a provider contract.
type MethodSig struct {
	Name       string        `json:"name"`
	Params     []FieldSchema `json:"params,omitempty"`
	ReturnType TypeRef       `json:"return_type"`
}

// RouteHandler describes an HTTP route in the IR.
type RouteHandler struct {
	Method      HTTPMethod       `json:"method"`
	Path        string           `json:"path"`
	PathParams  []string         `json:"path_params,omitempty"`
	QueryParams []QueryParam     `json:"query_params,omitempty"`
	InputType   *TypeRef         `json:"input_type,omitempty"`
	ReturnType  *TypeRef         `json:"return_type,omitempty"`
	Auth        *AuthRequirement `json:"auth,omitempty"`
	RateLimit   *RateLimitConfig `json:"rate_limit,omitempty"`
	Middleware  []MiddlewareRef  `json:"middleware,omitempty"`
	Providers   []InjectionRef   `json:"providers,omitempty"`
	Body        []StmtIR         `json:"-"`
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

// MarshalJSON serializes HTTPMethod as a string.
func (m HTTPMethod) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}

// QueryParam describes a declared query parameter.
type QueryParam struct {
	Name     string  `json:"name"`
	Type     TypeRef `json:"type"`
	Required bool    `json:"required"`
	Default  ExprIR  `json:"-"`
	IsArray  bool    `json:"is_array,omitempty"`
}

// AuthRequirement describes the authentication needed for a route.
type AuthRequirement struct {
	AuthType string   `json:"auth_type"` // e.g., "jwt", "apikey", "basic"
	Required bool     `json:"required"`
	Roles    []string `json:"roles,omitempty"`
}

// RateLimitConfig describes rate limiting for a route.
type RateLimitConfig struct {
	Requests uint32 `json:"requests"`
	Window   string `json:"window"`
}

// MiddlewareRef is a reference to a named middleware with optional arguments.
type MiddlewareRef struct {
	Name string   `json:"name"`
	Args []ExprIR `json:"-"`
}

// InjectionRef describes a provider injection into a handler.
type InjectionRef struct {
	Name         string `json:"name"`          // Local variable name (e.g., "db")
	ProviderType string `json:"provider_type"` // Provider type name (e.g., "Database")
}

// EventBinding describes an event handler.
type EventBinding struct {
	EventType string         `json:"event_type"`
	Async     bool           `json:"async"`
	Providers []InjectionRef `json:"providers,omitempty"`
	Body      []StmtIR       `json:"-"`
}

// CronBinding describes a scheduled task.
type CronBinding struct {
	Name      string         `json:"name"`
	Schedule  string         `json:"schedule"`
	Timezone  string         `json:"timezone,omitempty"`
	Retries   int            `json:"retries,omitempty"`
	Providers []InjectionRef `json:"providers,omitempty"`
	Body      []StmtIR       `json:"-"`
}

// QueueBinding describes a queue worker.
type QueueBinding struct {
	QueueName   string         `json:"queue_name"`
	Concurrency int            `json:"concurrency,omitempty"`
	MaxRetries  int            `json:"max_retries,omitempty"`
	Timeout     int            `json:"timeout,omitempty"`
	Providers   []InjectionRef `json:"providers,omitempty"`
	Body        []StmtIR       `json:"-"`
}

// CommandDef describes a CLI command.
type CommandDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Params      []CommandParam `json:"params,omitempty"`
	ReturnType  *TypeRef       `json:"return_type,omitempty"`
	Body        []StmtIR       `json:"-"`
}

// CommandParam describes a CLI command parameter.
type CommandParam struct {
	Name     string  `json:"name"`
	Type     TypeRef `json:"type"`
	Required bool    `json:"required"`
	Default  ExprIR  `json:"-"`
	IsFlag   bool    `json:"is_flag,omitempty"`
}

// FunctionDef describes a standalone function.
type FunctionDef struct {
	Name       string        `json:"name"`
	TypeParams []string      `json:"type_params,omitempty"`
	Params     []FieldSchema `json:"params,omitempty"`
	ReturnType *TypeRef      `json:"return_type,omitempty"`
	Body       []StmtIR      `json:"-"`
}

// GRPCServiceDef describes a gRPC service definition.
type GRPCServiceDef struct {
	Name     string           `json:"name"`
	Methods  []GRPCMethodDef  `json:"methods,omitempty"`
	Handlers []GRPCHandlerDef `json:"handlers,omitempty"`
}

// GRPCMethodDef describes a gRPC method signature.
type GRPCMethodDef struct {
	Name       string         `json:"name"`
	InputType  TypeRef        `json:"input_type"`
	ReturnType TypeRef        `json:"return_type"`
	StreamType GRPCStreamType `json:"stream_type"`
}

// GRPCStreamType indicates gRPC streaming mode.
type GRPCStreamType int

const (
	GRPCUnary GRPCStreamType = iota
	GRPCServerStream
	GRPCClientStream
	GRPCBidirectional
)

// String returns the GRPCStreamType as a string.
func (s GRPCStreamType) String() string {
	switch s {
	case GRPCUnary:
		return "unary"
	case GRPCServerStream:
		return "server_stream"
	case GRPCClientStream:
		return "client_stream"
	case GRPCBidirectional:
		return "bidirectional"
	default:
		return "unary"
	}
}

// MarshalJSON serializes GRPCStreamType as a string.
func (s GRPCStreamType) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// GRPCHandlerDef describes a gRPC handler implementation.
type GRPCHandlerDef struct {
	ServiceName string           `json:"service_name"`
	MethodName  string           `json:"method_name"`
	Params      []FieldSchema    `json:"params,omitempty"`
	ReturnType  *TypeRef         `json:"return_type,omitempty"`
	StreamType  GRPCStreamType   `json:"stream_type"`
	Auth        *AuthRequirement `json:"auth,omitempty"`
	Providers   []InjectionRef   `json:"providers,omitempty"`
	Body        []StmtIR         `json:"-"`
}

// GraphQLDef describes a GraphQL resolver.
type GraphQLDef struct {
	Operation  GraphQLOp        `json:"operation"`
	FieldName  string           `json:"field_name"`
	Params     []FieldSchema    `json:"params,omitempty"`
	ReturnType *TypeRef         `json:"return_type,omitempty"`
	Auth       *AuthRequirement `json:"auth,omitempty"`
	Providers  []InjectionRef   `json:"providers,omitempty"`
	Body       []StmtIR         `json:"-"`
}

// GraphQLOp is the GraphQL operation type.
type GraphQLOp int

const (
	GraphQLQuery GraphQLOp = iota
	GraphQLMutation
	GraphQLSubscription
)

// String returns the GraphQLOp as a string.
func (o GraphQLOp) String() string {
	switch o {
	case GraphQLQuery:
		return "query"
	case GraphQLMutation:
		return "mutation"
	case GraphQLSubscription:
		return "subscription"
	default:
		return "query"
	}
}

// MarshalJSON serializes GraphQLOp as a string.
func (o GraphQLOp) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.String())
}

// WebSocketDef describes a WebSocket route.
type WebSocketDef struct {
	Path   string       `json:"path"`
	Events []WSEventDef `json:"events,omitempty"`
}

// WSEventDef describes a WebSocket event handler.
type WSEventDef struct {
	EventType WSEventType `json:"event_type"`
	Body      []StmtIR    `json:"-"`
}

// WSEventType identifies a WebSocket event.
type WSEventType int

const (
	WSConnect WSEventType = iota
	WSDisconnect
	WSMessage
	WSError
)

// String returns the WSEventType as a string.
func (t WSEventType) String() string {
	switch t {
	case WSConnect:
		return "connect"
	case WSDisconnect:
		return "disconnect"
	case WSMessage:
		return "message"
	case WSError:
		return "error"
	default:
		return "message"
	}
}

// MarshalJSON serializes WSEventType as a string.
func (t WSEventType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// ConstantDef describes a module-level constant.
type ConstantDef struct {
	Name  string   `json:"name"`
	Type  *TypeRef `json:"type,omitempty"`
	Value ExprIR   `json:"-"`
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
