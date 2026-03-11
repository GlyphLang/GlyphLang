package ir

import (
	"fmt"
	"strings"

	"github.com/glyphlang/glyph/pkg/ast"
)

// Analyzer transforms an AST Module into a Semantic IR ServiceIR.
// It resolves types, extracts provider requirements, and normalizes
// the AST into a language-neutral intermediate representation.
type Analyzer struct {
	types             map[string]ast.TypeDef
	providers         map[string]ProviderRef      // all discovered provider dependencies
	providerContracts map[string]*ast.ProviderDef // provider contracts from source
}

// NewAnalyzer creates a new Analyzer.
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		types:             make(map[string]ast.TypeDef),
		providers:         make(map[string]ProviderRef),
		providerContracts: make(map[string]*ast.ProviderDef),
	}
}

// Analyze transforms an AST Module into a ServiceIR.
func (a *Analyzer) Analyze(module *ast.Module) (*ServiceIR, error) {
	ir := &ServiceIR{}

	// First pass: collect type definitions and provider contracts
	for _, item := range module.Items {
		switch it := item.(type) {
		case *ast.TypeDef:
			a.types[it.Name] = *it
		case *ast.ProviderDef:
			a.providerContracts[it.Name] = it
		}
	}

	// Second pass: build IR from all items
	for _, item := range module.Items {
		if err := a.processItem(ir, item); err != nil {
			return nil, err
		}
	}

	// Collect unique providers
	for _, p := range a.providers {
		ir.Providers = append(ir.Providers, p)
	}

	return ir, nil
}

func (a *Analyzer) processItem(ir *ServiceIR, item ast.Item) error {
	switch it := item.(type) {
	case *ast.TypeDef:
		ts, err := a.convertTypeDef(it)
		if err != nil {
			return fmt.Errorf("type %s: %w", it.Name, err)
		}
		ir.Types = append(ir.Types, ts)

	case *ast.Route:
		rh, err := a.convertRoute(it)
		if err != nil {
			return fmt.Errorf("route %s %s: %w", it.Method, it.Path, err)
		}
		ir.Routes = append(ir.Routes, rh)

	case *ast.Function:
		fd, err := a.convertFunction(it)
		if err != nil {
			return fmt.Errorf("function %s: %w", it.Name, err)
		}
		ir.Functions = append(ir.Functions, fd)

	case *ast.CronTask:
		cb, err := a.convertCronTask(it)
		if err != nil {
			return fmt.Errorf("cron %s: %w", it.Name, err)
		}
		ir.CronJobs = append(ir.CronJobs, cb)

	case *ast.EventHandler:
		eb, err := a.convertEventHandler(it)
		if err != nil {
			return fmt.Errorf("event %s: %w", it.EventType, err)
		}
		ir.Events = append(ir.Events, eb)

	case *ast.QueueWorker:
		qb, err := a.convertQueueWorker(it)
		if err != nil {
			return fmt.Errorf("queue %s: %w", it.QueueName, err)
		}
		ir.Queues = append(ir.Queues, qb)

	case *ast.Command:
		cd, err := a.convertCommand(it)
		if err != nil {
			return fmt.Errorf("command %s: %w", it.Name, err)
		}
		ir.Commands = append(ir.Commands, cd)

	case *ast.GRPCService:
		sd := a.convertGRPCService(it)
		ir.GRPC = append(ir.GRPC, sd)

	case *ast.GRPCHandler:
		// GRPCHandlers are folded into their parent GRPCServiceDef during post-processing
		hd, err := a.convertGRPCHandler(it)
		if err != nil {
			return fmt.Errorf("grpc handler %s: %w", it.MethodName, err)
		}
		// Find or create a service entry to attach this handler
		a.attachGRPCHandler(ir, it.ServiceName, hd)

	case *ast.GraphQLResolver:
		gd, err := a.convertGraphQLResolver(it)
		if err != nil {
			return fmt.Errorf("graphql %s.%s: %w", it.Operation, it.FieldName, err)
		}
		ir.GraphQL = append(ir.GraphQL, gd)

	case *ast.WebSocketRoute:
		ws := a.convertWebSocketRoute(it)
		ir.WebSocket = append(ir.WebSocket, ws)

	case *ast.ConstDecl:
		cd, err := a.convertConstDecl(it)
		if err != nil {
			return fmt.Errorf("const %s: %w", it.Name, err)
		}
		ir.Constants = append(ir.Constants, cd)

	case *ast.ProviderDef:
		prov := a.convertProviderDef(it)
		// Register as a known provider so injections can reference it
		a.providers[prov.ProviderType] = prov

	case *ast.TestBlock, *ast.ImportStatement, *ast.ModuleDecl,
		*ast.MacroDef, *ast.MacroInvocation, *ast.ContractDef,
		*ast.TraitDef, *ast.StaticRoute:
		// These are either handled elsewhere or not represented in the service IR
	}
	return nil
}

// --- Type conversion ---

func (a *Analyzer) convertTypeDef(td *ast.TypeDef) (TypeSchema, error) {
	ts := TypeSchema{
		Name:   td.Name,
		Traits: td.Traits,
	}
	for _, tp := range td.TypeParams {
		ts.TypeParams = append(ts.TypeParams, tp.Name)
	}
	for _, f := range td.Fields {
		fs, err := a.convertField(f)
		if err != nil {
			return ts, err
		}
		ts.Fields = append(ts.Fields, fs)
	}
	for _, m := range td.Methods {
		ms, err := a.convertMethodDef(m)
		if err != nil {
			return ts, err
		}
		ts.Methods = append(ts.Methods, ms)
	}
	return ts, nil
}

func (a *Analyzer) convertField(f ast.Field) (FieldSchema, error) {
	tr := a.convertType(f.TypeAnnotation)
	fs := FieldSchema{
		Name:     f.Name,
		Type:     tr,
		Required: f.Required,
	}
	if f.Default != nil {
		fs.HasDefault = true
		fs.Default = a.convertExpr(f.Default)
	}
	for _, ann := range f.Annotations {
		fs.Annotations = append(fs.Annotations, Annotation{
			Name:   ann.Name,
			Params: ann.Params,
		})
	}
	return fs, nil
}

func (a *Analyzer) convertMethodDef(m ast.MethodDef) (MethodSchema, error) {
	ms := MethodSchema{
		Name:       m.Name,
		ReturnType: a.convertType(m.ReturnType),
	}
	for _, p := range m.Params {
		fs, err := a.convertField(p)
		if err != nil {
			return ms, err
		}
		ms.Params = append(ms.Params, fs)
	}
	ms.Body = a.convertStatements(m.Body)
	return ms, nil
}

func (a *Analyzer) convertType(t ast.Type) TypeRef {
	if t == nil {
		return TypeRef{Kind: TypeAny}
	}
	switch v := t.(type) {
	case ast.IntType:
		return TypeRef{Kind: TypeInt}
	case ast.FloatType:
		return TypeRef{Kind: TypeFloat}
	case ast.StringType:
		return TypeRef{Kind: TypeString}
	case ast.BoolType:
		return TypeRef{Kind: TypeBool}
	case ast.ArrayType:
		inner := a.convertType(v.ElementType)
		return TypeRef{Kind: TypeArray, Inner: &inner}
	case ast.OptionalType:
		inner := a.convertType(v.InnerType)
		return TypeRef{Kind: TypeOptional, Inner: &inner}
	case ast.NamedType:
		return TypeRef{Kind: TypeNamed, Name: v.Name}
	case ast.DatabaseType:
		return TypeRef{Kind: TypeProvider, Name: "Database"}
	case ast.RedisType:
		return TypeRef{Kind: TypeProvider, Name: "Redis"}
	case ast.MongoDBType:
		return TypeRef{Kind: TypeProvider, Name: "MongoDB"}
	case ast.LLMType:
		return TypeRef{Kind: TypeProvider, Name: "LLM"}
	case ast.UnionType:
		tr := TypeRef{Kind: TypeUnion}
		for _, ut := range v.Types {
			tr.Elements = append(tr.Elements, a.convertType(ut))
		}
		return tr
	case ast.GenericType:
		tr := TypeRef{Kind: TypeGeneric, Name: ""}
		base := a.convertType(v.BaseType)
		tr.Name = base.Name
		for _, arg := range v.TypeArgs {
			tr.Elements = append(tr.Elements, a.convertType(arg))
		}
		return tr
	case ast.FunctionType:
		tr := TypeRef{Kind: TypeFunction}
		for _, pt := range v.ParamTypes {
			tr.Elements = append(tr.Elements, a.convertType(pt))
		}
		ret := a.convertType(v.ReturnType)
		tr.Inner = &ret
		return tr
	case ast.FutureType:
		inner := a.convertType(v.ResultType)
		return TypeRef{Kind: TypeFuture, Inner: &inner}
	case ast.TypeParameterType:
		return TypeRef{Kind: TypeNamed, Name: v.Name}
	default:
		return TypeRef{Kind: TypeAny}
	}
}

// --- Route conversion ---

func (a *Analyzer) convertRoute(r *ast.Route) (RouteHandler, error) {
	rh := RouteHandler{
		Method:     convertHTTPMethod(r.Method),
		Path:       r.Path,
		PathParams: extractPathParams(r.Path),
	}

	if r.InputType != nil {
		tr := a.convertType(r.InputType)
		rh.InputType = &tr
	}
	if r.ReturnType != nil {
		tr := a.convertType(r.ReturnType)
		rh.ReturnType = &tr
	}
	if r.Auth != nil {
		rh.Auth = &AuthRequirement{
			AuthType: r.Auth.AuthType,
			Required: r.Auth.Required,
		}
	}
	if r.RateLimit != nil {
		rh.RateLimit = &RateLimitConfig{
			Requests: r.RateLimit.Requests,
			Window:   r.RateLimit.Window,
		}
	}
	for _, qp := range r.QueryParams {
		irQP := QueryParam{
			Name:     qp.Name,
			Type:     a.convertType(qp.Type),
			Required: qp.Required,
			IsArray:  qp.IsArray,
		}
		if qp.Default != nil {
			irQP.Default = a.convertExpr(qp.Default)
		}
		rh.QueryParams = append(rh.QueryParams, irQP)
	}
	for _, inj := range r.Injections {
		ref := a.registerInjection(inj)
		rh.Providers = append(rh.Providers, ref)
	}
	rh.Body = a.convertStatements(r.Body)
	return rh, nil
}

func convertHTTPMethod(m ast.HttpMethod) HTTPMethod {
	switch m {
	case ast.Get:
		return MethodGet
	case ast.Post:
		return MethodPost
	case ast.Put:
		return MethodPut
	case ast.Delete:
		return MethodDelete
	case ast.Patch:
		return MethodPatch
	case ast.WebSocket:
		return MethodWebSocket
	case ast.SSE:
		return MethodSSE
	default:
		return MethodGet
	}
}

func extractPathParams(path string) []string {
	var params []string
	for _, seg := range strings.Split(path, "/") {
		if strings.HasPrefix(seg, ":") {
			params = append(params, seg[1:])
		}
	}
	return params
}

// --- Injection / Provider tracking ---

func (a *Analyzer) registerInjection(inj ast.Injection) InjectionRef {
	providerType := resolveProviderType(inj.Type)
	ref := InjectionRef{
		Name:         inj.Name,
		ProviderType: providerType,
	}

	// Track unique providers
	if _, exists := a.providers[providerType]; !exists {
		prov := ProviderRef{
			Name:         strings.ToLower(providerType),
			ProviderType: providerType,
			IsStandard:   isStandardProvider(providerType),
		}
		// If a provider contract was declared, populate its methods
		if contract, ok := a.providerContracts[providerType]; ok {
			prov = a.convertProviderDef(contract)
		}
		a.providers[providerType] = prov
	}
	return ref
}

func resolveProviderType(t ast.Type) string {
	switch t.(type) {
	case ast.DatabaseType:
		return "Database"
	case ast.RedisType:
		return "Redis"
	case ast.MongoDBType:
		return "MongoDB"
	case ast.LLMType:
		return "LLM"
	case ast.NamedType:
		return t.(ast.NamedType).Name
	default:
		return "Unknown"
	}
}

func isStandardProvider(name string) bool {
	switch name {
	case "Database", "Redis", "MongoDB", "LLM":
		return true
	default:
		return false
	}
}

// --- Provider conversion ---

func (a *Analyzer) convertProviderDef(pd *ast.ProviderDef) ProviderRef {
	prov := ProviderRef{
		Name:         strings.ToLower(pd.Name),
		ProviderType: pd.Name,
		IsStandard:   false,
	}
	for _, m := range pd.Methods {
		ms := MethodSig{
			Name: m.Name,
		}
		for _, p := range m.Params {
			fs, _ := a.convertField(p)
			ms.Params = append(ms.Params, fs)
		}
		if m.ReturnType != nil {
			ms.ReturnType = a.convertType(m.ReturnType)
		}
		prov.Methods = append(prov.Methods, ms)
	}
	return prov
}

// --- Function, CronTask, EventHandler, QueueWorker, Command conversion ---

func (a *Analyzer) convertFunction(f *ast.Function) (FunctionDef, error) {
	fd := FunctionDef{
		Name: f.Name,
	}
	for _, tp := range f.TypeParams {
		fd.TypeParams = append(fd.TypeParams, tp.Name)
	}
	for _, p := range f.Params {
		fs, err := a.convertField(p)
		if err != nil {
			return fd, err
		}
		fd.Params = append(fd.Params, fs)
	}
	if f.ReturnType != nil {
		tr := a.convertType(f.ReturnType)
		fd.ReturnType = &tr
	}
	fd.Body = a.convertStatements(f.Body)
	return fd, nil
}

func (a *Analyzer) convertCronTask(ct *ast.CronTask) (CronBinding, error) {
	cb := CronBinding{
		Name:     ct.Name,
		Schedule: ct.Schedule,
		Timezone: ct.Timezone,
		Retries:  ct.Retries,
	}
	for _, inj := range ct.Injections {
		cb.Providers = append(cb.Providers, a.registerInjection(inj))
	}
	cb.Body = a.convertStatements(ct.Body)
	return cb, nil
}

func (a *Analyzer) convertEventHandler(eh *ast.EventHandler) (EventBinding, error) {
	eb := EventBinding{
		EventType: eh.EventType,
		Async:     eh.Async,
	}
	for _, inj := range eh.Injections {
		eb.Providers = append(eb.Providers, a.registerInjection(inj))
	}
	eb.Body = a.convertStatements(eh.Body)
	return eb, nil
}

func (a *Analyzer) convertQueueWorker(qw *ast.QueueWorker) (QueueBinding, error) {
	qb := QueueBinding{
		QueueName:   qw.QueueName,
		Concurrency: qw.Concurrency,
		MaxRetries:  qw.MaxRetries,
		Timeout:     qw.Timeout,
	}
	for _, inj := range qw.Injections {
		qb.Providers = append(qb.Providers, a.registerInjection(inj))
	}
	qb.Body = a.convertStatements(qw.Body)
	return qb, nil
}

func (a *Analyzer) convertCommand(cmd *ast.Command) (CommandDef, error) {
	cd := CommandDef{
		Name:        cmd.Name,
		Description: cmd.Description,
	}
	for _, p := range cmd.Params {
		cp := CommandParam{
			Name:     p.Name,
			Type:     a.convertType(p.Type),
			Required: p.Required,
			IsFlag:   p.IsFlag,
		}
		if p.Default != nil {
			cp.Default = a.convertExpr(p.Default)
		}
		cd.Params = append(cd.Params, cp)
	}
	if cmd.ReturnType != nil {
		tr := a.convertType(cmd.ReturnType)
		cd.ReturnType = &tr
	}
	cd.Body = a.convertStatements(cmd.Body)
	return cd, nil
}

// --- gRPC conversion ---

func (a *Analyzer) convertGRPCService(svc *ast.GRPCService) GRPCServiceDef {
	sd := GRPCServiceDef{Name: svc.Name}
	for _, m := range svc.Methods {
		sd.Methods = append(sd.Methods, GRPCMethodDef{
			Name:       m.Name,
			InputType:  a.convertType(m.InputType),
			ReturnType: a.convertType(m.ReturnType),
			StreamType: GRPCStreamType(m.StreamType),
		})
	}
	return sd
}

func (a *Analyzer) convertGRPCHandler(h *ast.GRPCHandler) (GRPCHandlerDef, error) {
	hd := GRPCHandlerDef{
		ServiceName: h.ServiceName,
		MethodName:  h.MethodName,
		StreamType:  GRPCStreamType(h.StreamType),
	}
	for _, p := range h.Params {
		fs, err := a.convertField(p)
		if err != nil {
			return hd, err
		}
		hd.Params = append(hd.Params, fs)
	}
	if h.ReturnType != nil {
		tr := a.convertType(h.ReturnType)
		hd.ReturnType = &tr
	}
	if h.Auth != nil {
		hd.Auth = &AuthRequirement{
			AuthType: h.Auth.AuthType,
			Required: h.Auth.Required,
		}
	}
	for _, inj := range h.Injections {
		hd.Providers = append(hd.Providers, a.registerInjection(inj))
	}
	hd.Body = a.convertStatements(h.Body)
	return hd, nil
}

func (a *Analyzer) attachGRPCHandler(ir *ServiceIR, serviceName string, hd GRPCHandlerDef) {
	for i := range ir.GRPC {
		if ir.GRPC[i].Name == serviceName {
			ir.GRPC[i].Handlers = append(ir.GRPC[i].Handlers, hd)
			return
		}
	}
	// If no matching service, create an anonymous one
	ir.GRPC = append(ir.GRPC, GRPCServiceDef{
		Name:     serviceName,
		Handlers: []GRPCHandlerDef{hd},
	})
}

// --- GraphQL conversion ---

func (a *Analyzer) convertGraphQLResolver(r *ast.GraphQLResolver) (GraphQLDef, error) {
	gd := GraphQLDef{
		Operation: GraphQLOp(r.Operation),
		FieldName: r.FieldName,
	}
	for _, p := range r.Params {
		fs, err := a.convertField(p)
		if err != nil {
			return gd, err
		}
		gd.Params = append(gd.Params, fs)
	}
	if r.ReturnType != nil {
		tr := a.convertType(r.ReturnType)
		gd.ReturnType = &tr
	}
	if r.Auth != nil {
		gd.Auth = &AuthRequirement{
			AuthType: r.Auth.AuthType,
			Required: r.Auth.Required,
		}
	}
	for _, inj := range r.Injections {
		gd.Providers = append(gd.Providers, a.registerInjection(inj))
	}
	gd.Body = a.convertStatements(r.Body)
	return gd, nil
}

// --- WebSocket conversion ---

func (a *Analyzer) convertWebSocketRoute(ws *ast.WebSocketRoute) WebSocketDef {
	def := WebSocketDef{Path: ws.Path}
	for _, ev := range ws.Events {
		def.Events = append(def.Events, WSEventDef{
			EventType: WSEventType(ev.EventType),
			Body:      a.convertStatements(ev.Body),
		})
	}
	return def
}

// --- Const conversion ---

func (a *Analyzer) convertConstDecl(cd *ast.ConstDecl) (ConstantDef, error) {
	def := ConstantDef{Name: cd.Name}
	if cd.Type != nil {
		tr := a.convertType(cd.Type)
		def.Type = &tr
	}
	if cd.Value != nil {
		def.Value = a.convertExpr(cd.Value)
	}
	return def, nil
}

// --- Statement conversion ---

func (a *Analyzer) convertStatements(stmts []ast.Statement) []StmtIR {
	var result []StmtIR
	for _, stmt := range stmts {
		result = append(result, a.convertStatement(stmt))
	}
	return result
}

func (a *Analyzer) convertStatement(stmt ast.Statement) StmtIR {
	switch s := stmt.(type) {
	case *ast.AssignStatement:
		expr := a.convertExpr(s.Value)
		return StmtIR{
			Kind:   StmtAssign,
			Assign: &AssignStmt{Target: s.Target, Value: expr},
		}
	case ast.AssignStatement:
		expr := a.convertExpr(s.Value)
		return StmtIR{
			Kind:   StmtAssign,
			Assign: &AssignStmt{Target: s.Target, Value: expr},
		}
	case *ast.ReassignStatement:
		expr := a.convertExpr(s.Value)
		return StmtIR{
			Kind:   StmtReassign,
			Assign: &AssignStmt{Target: s.Target, Value: expr},
		}
	case ast.ReassignStatement:
		expr := a.convertExpr(s.Value)
		return StmtIR{
			Kind:   StmtReassign,
			Assign: &AssignStmt{Target: s.Target, Value: expr},
		}
	case *ast.ReturnStatement:
		expr := a.convertExpr(s.Value)
		return StmtIR{
			Kind:   StmtReturn,
			Return: &ReturnStmt{Value: expr},
		}
	case ast.ReturnStatement:
		expr := a.convertExpr(s.Value)
		return StmtIR{
			Kind:   StmtReturn,
			Return: &ReturnStmt{Value: expr},
		}
	case *ast.IfStatement:
		return StmtIR{
			Kind: StmtIf,
			If: &IfStmt{
				Condition: a.convertExpr(s.Condition),
				Then:      a.convertStatements(s.ThenBlock),
				Else:      a.convertStatements(s.ElseBlock),
			},
		}
	case ast.IfStatement:
		return StmtIR{
			Kind: StmtIf,
			If: &IfStmt{
				Condition: a.convertExpr(s.Condition),
				Then:      a.convertStatements(s.ThenBlock),
				Else:      a.convertStatements(s.ElseBlock),
			},
		}
	case *ast.ForStatement:
		return StmtIR{
			Kind: StmtFor,
			For: &ForStmt{
				KeyVar:   s.KeyVar,
				ValueVar: s.ValueVar,
				Iterable: a.convertExpr(s.Iterable),
				Body:     a.convertStatements(s.Body),
			},
		}
	case ast.ForStatement:
		return StmtIR{
			Kind: StmtFor,
			For: &ForStmt{
				KeyVar:   s.KeyVar,
				ValueVar: s.ValueVar,
				Iterable: a.convertExpr(s.Iterable),
				Body:     a.convertStatements(s.Body),
			},
		}
	case *ast.WhileStatement:
		return StmtIR{
			Kind: StmtWhile,
			While: &WhileStmt{
				Condition: a.convertExpr(s.Condition),
				Body:      a.convertStatements(s.Body),
			},
		}
	case ast.WhileStatement:
		return StmtIR{
			Kind: StmtWhile,
			While: &WhileStmt{
				Condition: a.convertExpr(s.Condition),
				Body:      a.convertStatements(s.Body),
			},
		}
	case *ast.SwitchStatement:
		sw := &SwitchStmt{
			Value:   a.convertExpr(s.Value),
			Default: a.convertStatements(s.Default),
		}
		for _, c := range s.Cases {
			sw.Cases = append(sw.Cases, SwitchCase{
				Value: a.convertExpr(c.Value),
				Body:  a.convertStatements(c.Body),
			})
		}
		return StmtIR{Kind: StmtSwitch, Switch: sw}
	case ast.SwitchStatement:
		sw := &SwitchStmt{
			Value:   a.convertExpr(s.Value),
			Default: a.convertStatements(s.Default),
		}
		for _, c := range s.Cases {
			sw.Cases = append(sw.Cases, SwitchCase{
				Value: a.convertExpr(c.Value),
				Body:  a.convertStatements(c.Body),
			})
		}
		return StmtIR{Kind: StmtSwitch, Switch: sw}
	case *ast.ExpressionStatement:
		expr := a.convertExpr(s.Expr)
		return StmtIR{Kind: StmtExpr, ExprStmt: &expr}
	case ast.ExpressionStatement:
		expr := a.convertExpr(s.Expr)
		return StmtIR{Kind: StmtExpr, ExprStmt: &expr}
	case *ast.ValidationStatement:
		expr := a.convertExpr(&s.Call)
		return StmtIR{Kind: StmtValidate, Validate: &ValidateStmt{Call: expr}}
	case ast.ValidationStatement:
		expr := a.convertExpr(&s.Call)
		return StmtIR{Kind: StmtValidate, Validate: &ValidateStmt{Call: expr}}
	case *ast.BreakStatement, ast.BreakStatement:
		return StmtIR{Kind: StmtBreak, Break: true}
	case *ast.ContinueStatement, ast.ContinueStatement:
		return StmtIR{Kind: StmtContinue, Continue: true}
	default:
		// Fallback for unhandled statement types
		return StmtIR{Kind: StmtExpr}
	}
}

// --- Expression conversion ---

func (a *Analyzer) convertExpr(expr ast.Expr) ExprIR {
	if expr == nil {
		return ExprIR{Kind: ExprNull, IsNull: true}
	}
	switch e := expr.(type) {
	case *ast.LiteralExpr:
		return a.convertLiteral(e.Value)
	case ast.LiteralExpr:
		return a.convertLiteral(e.Value)
	case *ast.VariableExpr:
		return ExprIR{Kind: ExprVar, VarName: e.Name}
	case ast.VariableExpr:
		return ExprIR{Kind: ExprVar, VarName: e.Name}
	case *ast.BinaryOpExpr:
		return ExprIR{
			Kind: ExprBinary,
			BinOp: &BinaryExpr{
				Op:    convertBinOp(e.Op),
				Left:  a.convertExpr(e.Left),
				Right: a.convertExpr(e.Right),
			},
		}
	case ast.BinaryOpExpr:
		return ExprIR{
			Kind: ExprBinary,
			BinOp: &BinaryExpr{
				Op:    convertBinOp(e.Op),
				Left:  a.convertExpr(e.Left),
				Right: a.convertExpr(e.Right),
			},
		}
	case *ast.UnaryOpExpr:
		return ExprIR{
			Kind: ExprUnary,
			UnaryOp: &UnaryExpr{
				Op:    convertUnOp(e.Op),
				Right: a.convertExpr(e.Right),
			},
		}
	case ast.UnaryOpExpr:
		return ExprIR{
			Kind: ExprUnary,
			UnaryOp: &UnaryExpr{
				Op:    convertUnOp(e.Op),
				Right: a.convertExpr(e.Right),
			},
		}
	case *ast.FieldAccessExpr:
		return ExprIR{
			Kind: ExprFieldAccess,
			FieldAccess: &FieldAccessExpr{
				Object: a.convertExpr(e.Object),
				Field:  e.Field,
			},
		}
	case ast.FieldAccessExpr:
		return ExprIR{
			Kind: ExprFieldAccess,
			FieldAccess: &FieldAccessExpr{
				Object: a.convertExpr(e.Object),
				Field:  e.Field,
			},
		}
	case *ast.ArrayIndexExpr:
		return ExprIR{
			Kind: ExprIndexAccess,
			IndexAccess: &IndexAccessExpr{
				Object: a.convertExpr(e.Array),
				Index:  a.convertExpr(e.Index),
			},
		}
	case ast.ArrayIndexExpr:
		return ExprIR{
			Kind: ExprIndexAccess,
			IndexAccess: &IndexAccessExpr{
				Object: a.convertExpr(e.Array),
				Index:  a.convertExpr(e.Index),
			},
		}
	case *ast.FunctionCallExpr:
		call := &CallExpr{Name: e.Name}
		for _, ta := range e.TypeArgs {
			call.TypeArgs = append(call.TypeArgs, a.convertType(ta))
		}
		for _, arg := range e.Args {
			call.Args = append(call.Args, a.convertExpr(arg))
		}
		return ExprIR{Kind: ExprCall, Call: call}
	case ast.FunctionCallExpr:
		call := &CallExpr{Name: e.Name}
		for _, ta := range e.TypeArgs {
			call.TypeArgs = append(call.TypeArgs, a.convertType(ta))
		}
		for _, arg := range e.Args {
			call.Args = append(call.Args, a.convertExpr(arg))
		}
		return ExprIR{Kind: ExprCall, Call: call}
	case *ast.ObjectExpr:
		obj := &ObjectExpr{}
		for _, f := range e.Fields {
			obj.Fields = append(obj.Fields, ObjectFieldIR{
				Key:   f.Key,
				Value: a.convertExpr(f.Value),
			})
		}
		return ExprIR{Kind: ExprObject, Object: obj}
	case ast.ObjectExpr:
		obj := &ObjectExpr{}
		for _, f := range e.Fields {
			obj.Fields = append(obj.Fields, ObjectFieldIR{
				Key:   f.Key,
				Value: a.convertExpr(f.Value),
			})
		}
		return ExprIR{Kind: ExprObject, Object: obj}
	case *ast.ArrayExpr:
		arr := &ArrayExpr{}
		for _, el := range e.Elements {
			arr.Elements = append(arr.Elements, a.convertExpr(el))
		}
		return ExprIR{Kind: ExprArray, Array: arr}
	case ast.ArrayExpr:
		arr := &ArrayExpr{}
		for _, el := range e.Elements {
			arr.Elements = append(arr.Elements, a.convertExpr(el))
		}
		return ExprIR{Kind: ExprArray, Array: arr}
	case *ast.LambdaExpr:
		lam := &LambdaExpr{}
		for _, p := range e.Params {
			fs, _ := a.convertField(p)
			lam.Params = append(lam.Params, fs)
		}
		if e.Body != nil {
			lam.Body = a.convertExpr(e.Body)
		}
		if len(e.Block) > 0 {
			lam.Block = a.convertStatements(e.Block)
		}
		return ExprIR{Kind: ExprLambda, Lambda: lam}
	case ast.LambdaExpr:
		lam := &LambdaExpr{}
		for _, p := range e.Params {
			fs, _ := a.convertField(p)
			lam.Params = append(lam.Params, fs)
		}
		if e.Body != nil {
			lam.Body = a.convertExpr(e.Body)
		}
		if len(e.Block) > 0 {
			lam.Block = a.convertStatements(e.Block)
		}
		return ExprIR{Kind: ExprLambda, Lambda: lam}
	case *ast.PipeExpr:
		return ExprIR{
			Kind: ExprPipe,
			Pipe: &PipeExpr{
				Left:  a.convertExpr(e.Left),
				Right: a.convertExpr(e.Right),
			},
		}
	case ast.PipeExpr:
		return ExprIR{
			Kind: ExprPipe,
			Pipe: &PipeExpr{
				Left:  a.convertExpr(e.Left),
				Right: a.convertExpr(e.Right),
			},
		}
	case *ast.MatchExpr:
		return a.convertMatchExpr(e)
	case ast.MatchExpr:
		return a.convertMatchExpr(&e)
	case *ast.AsyncExpr:
		return ExprIR{
			Kind:  ExprAsync,
			Async: &AsyncExprIR{Body: a.convertStatements(e.Body)},
		}
	case ast.AsyncExpr:
		return ExprIR{
			Kind:  ExprAsync,
			Async: &AsyncExprIR{Body: a.convertStatements(e.Body)},
		}
	case *ast.AwaitExpr:
		return ExprIR{
			Kind:  ExprAwait,
			Await: &AwaitExprIR{Expr: a.convertExpr(e.Expr)},
		}
	case ast.AwaitExpr:
		return ExprIR{
			Kind:  ExprAwait,
			Await: &AwaitExprIR{Expr: a.convertExpr(e.Expr)},
		}
	default:
		return ExprIR{Kind: ExprNull, IsNull: true}
	}
}

func (a *Analyzer) convertLiteral(lit ast.Literal) ExprIR {
	switch l := lit.(type) {
	case ast.IntLiteral:
		return ExprIR{Kind: ExprInt, IntVal: l.Value}
	case ast.FloatLiteral:
		return ExprIR{Kind: ExprFloat, FloatVal: l.Value}
	case ast.StringLiteral:
		return ExprIR{Kind: ExprString, StringVal: l.Value}
	case ast.BoolLiteral:
		return ExprIR{Kind: ExprBool, BoolVal: l.Value}
	case ast.NullLiteral:
		return ExprIR{Kind: ExprNull, IsNull: true}
	default:
		return ExprIR{Kind: ExprNull, IsNull: true}
	}
}

func (a *Analyzer) convertMatchExpr(m *ast.MatchExpr) ExprIR {
	me := &MatchExpr{
		Value: a.convertExpr(m.Value),
	}
	for _, c := range m.Cases {
		mc := MatchCase{
			Pattern: a.convertPattern(c.Pattern),
			Body:    a.convertExpr(c.Body),
		}
		if c.Guard != nil {
			guard := a.convertExpr(c.Guard)
			mc.Guard = &guard
		}
		me.Cases = append(me.Cases, mc)
	}
	return ExprIR{Kind: ExprMatch, Match: me}
}

func (a *Analyzer) convertPattern(p ast.Pattern) PatternIR {
	switch pat := p.(type) {
	case ast.LiteralPattern:
		ir := PatternIR{Kind: PatternLiteral}
		switch l := pat.Value.(type) {
		case ast.IntLiteral:
			ir.IntVal = l.Value
		case ast.FloatLiteral:
			ir.FloatVal = l.Value
		case ast.StringLiteral:
			ir.StrVal = l.Value
		case ast.BoolLiteral:
			ir.BoolVal = l.Value
		}
		return ir
	case ast.VariablePattern:
		return PatternIR{Kind: PatternVariable, VarName: pat.Name}
	case ast.WildcardPattern:
		return PatternIR{Kind: PatternWildcard}
	case ast.ObjectPattern:
		ir := PatternIR{Kind: PatternObject}
		for _, f := range pat.Fields {
			opf := ObjectPatternField{Key: f.Key}
			if f.Pattern != nil {
				opf.Pattern = a.convertPattern(f.Pattern)
			}
			ir.Fields = append(ir.Fields, opf)
		}
		return ir
	case ast.ArrayPattern:
		ir := PatternIR{Kind: PatternArray}
		for _, el := range pat.Elements {
			ir.Elements = append(ir.Elements, a.convertPattern(el))
		}
		if pat.Rest != nil {
			ir.RestVar = *pat.Rest
		}
		return ir
	default:
		return PatternIR{Kind: PatternWildcard}
	}
}

func convertBinOp(op ast.BinOp) BinOp {
	switch op {
	case ast.Add:
		return OpAdd
	case ast.Sub:
		return OpSub
	case ast.Mul:
		return OpMul
	case ast.Div:
		return OpDiv
	case ast.Mod:
		return OpMod
	case ast.Eq:
		return OpEq
	case ast.Ne:
		return OpNe
	case ast.Lt:
		return OpLt
	case ast.Le:
		return OpLe
	case ast.Gt:
		return OpGt
	case ast.Ge:
		return OpGe
	case ast.And:
		return OpAnd
	case ast.Or:
		return OpOr
	default:
		return OpAdd
	}
}

func convertUnOp(op ast.UnOp) UnOp {
	switch op {
	case ast.Not:
		return OpNot
	case ast.Neg:
		return OpNeg
	default:
		return OpNot
	}
}
