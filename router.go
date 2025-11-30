package silverlining

import (
	"errors"
	"net"
	"strings"
)

var (
	errNoHandlers       = errors.New("silverlining: at least one handler is required")
	errInvalidWildcard  = errors.New("silverlining: wildcard must be the last path segment")
	errInvalidParameter = errors.New("silverlining: invalid parameter name")
	errRouteExists      = errors.New("silverlining: route already registered")
)

// App provides a Fiber-like router with grouping, middleware, and parameter support.
type App struct {
	tree *routeNode

	globalHandlers []Handler

	notFound         Handler
	methodNotAllowed Handler
}

// Route represents a registered route.
type Route struct {
	Method Method
	Path   string
}

// New creates a new router instance.
func New() *App {
	return &App{
		tree:             newRouteNode(""),
		notFound:         defaultNotFoundHandler,
		methodNotAllowed: defaultMethodNotAllowedHandler,
	}
}

// Handler returns the server-compatible Handler.
func (a *App) Handler() Handler {
	return a.handleRequest
}

// Listen starts the server using the provided address.
func (a *App) Listen(addr string) error {
	return ListenAndServe(addr, a.Handler())
}

// ListenReusePort starts the server with SO_REUSEPORT enabled.
func (a *App) ListenReusePort(addr string) error {
	return ListenAndServeReusePort(addr, a.Handler())
}

// ListenPrefork starts the server using prefork mode.
func (a *App) ListenPrefork(addr string) error {
	return ListenAndServePrefork(addr, a.Handler())
}

// Serve attaches the router to an existing listener.
func (a *App) Serve(l net.Listener) error {
	s := &Server{Listener: l, Handler: a.Handler()}
	return s.Serve(l)
}

// Use registers global middleware executed before every route.
func (a *App) Use(handlers ...Handler) {
	a.globalHandlers = append(a.globalHandlers, handlers...)
}

// NotFound overrides the 404 handler.
func (a *App) NotFound(handler Handler) {
	a.notFound = handler
}

// MethodNotAllowed overrides the 405 handler.
func (a *App) MethodNotAllowed(handler Handler) {
	a.methodNotAllowed = handler
}

// Handle registers a route for the given method and path.
func (a *App) Handle(method Method, path string, handlers ...Handler) *Route {
	route, err := a.addRoute(method, path, nil, handlers...)
	if err != nil {
		panic(err)
	}
	return route
}

// Get registers a handler for HTTP GET.
func (a *App) Get(path string, handlers ...Handler) *Route {
	return a.Handle(MethodGET, path, handlers...)
}

// Head registers a handler for HTTP HEAD.
func (a *App) Head(path string, handlers ...Handler) *Route {
	return a.Handle(MethodHEAD, path, handlers...)
}

// Post registers a handler for HTTP POST.
func (a *App) Post(path string, handlers ...Handler) *Route {
	return a.Handle(MethodPOST, path, handlers...)
}

// Put registers a handler for HTTP PUT.
func (a *App) Put(path string, handlers ...Handler) *Route {
	return a.Handle(MethodPUT, path, handlers...)
}

// Delete registers a handler for HTTP DELETE.
func (a *App) Delete(path string, handlers ...Handler) *Route {
	return a.Handle(MethodDELETE, path, handlers...)
}

// Patch registers a handler for HTTP PATCH.
func (a *App) Patch(path string, handlers ...Handler) *Route {
	return a.Handle(MethodPATCH, path, handlers...)
}

// Options registers a handler for HTTP OPTIONS.
func (a *App) Options(path string, handlers ...Handler) *Route {
	return a.Handle(MethodOPTIONS, path, handlers...)
}

// Connect registers a handler for HTTP CONNECT.
func (a *App) Connect(path string, handlers ...Handler) *Route {
	return a.Handle(MethodCONNECT, path, handlers...)
}

// Trace registers a handler for HTTP TRACE.
func (a *App) Trace(path string, handlers ...Handler) *Route {
	return a.Handle(MethodTRACE, path, handlers...)
}

// Brew registers a handler for HTTP BREW (Easter egg method).
func (a *App) Brew(path string, handlers ...Handler) *Route {
	return a.Handle(MethodBREW, path, handlers...)
}

// All registers the handlers for every supported HTTP method.
func (a *App) All(path string, handlers ...Handler) []*Route {
	methods := []Method{
		MethodGET,
		MethodHEAD,
		MethodPOST,
		MethodPUT,
		MethodDELETE,
		MethodCONNECT,
		MethodOPTIONS,
		MethodTRACE,
		MethodPATCH,
		MethodBREW,
	}

	routes := make([]*Route, 0, len(methods))
	for _, m := range methods {
		routes = append(routes, a.Handle(m, path, handlers...))
	}
	return routes
}

// Group creates a new route group with the provided prefix and middleware.
func (a *App) Group(path string, handlers ...Handler) *RouteGroup {
	return &RouteGroup{
		app:         a,
		basePath:    normalizeGroupPath(path),
		middlewares: cloneHandlers(handlers),
	}
}

// RouteGroup allows grouped routes that share a prefix and middleware stack.
type RouteGroup struct {
	app         *App
	basePath    string
	middlewares []Handler
}

// Use appends middleware to the group.
func (g *RouteGroup) Use(handlers ...Handler) {
	g.middlewares = append(g.middlewares, handlers...)
}

// Group creates a nested route group.
func (g *RouteGroup) Group(path string, handlers ...Handler) *RouteGroup {
	next := cloneHandlers(g.middlewares)
	next = append(next, handlers...)
	return &RouteGroup{
		app:         g.app,
		basePath:    combinePaths(g.basePath, path),
		middlewares: next,
	}
}

// Handle registers a route within the group.
func (g *RouteGroup) Handle(method Method, path string, handlers ...Handler) *Route {
	route, err := g.app.addRoute(method, combinePaths(g.basePath, path), g.middlewares, handlers...)
	if err != nil {
		panic(err)
	}
	return route
}

// Convenience helpers mirroring App methods.
func (g *RouteGroup) Get(path string, handlers ...Handler) *Route {
	return g.Handle(MethodGET, path, handlers...)
}

func (g *RouteGroup) Head(path string, handlers ...Handler) *Route {
	return g.Handle(MethodHEAD, path, handlers...)
}

func (g *RouteGroup) Post(path string, handlers ...Handler) *Route {
	return g.Handle(MethodPOST, path, handlers...)
}

func (g *RouteGroup) Put(path string, handlers ...Handler) *Route {
	return g.Handle(MethodPUT, path, handlers...)
}

func (g *RouteGroup) Delete(path string, handlers ...Handler) *Route {
	return g.Handle(MethodDELETE, path, handlers...)
}

func (g *RouteGroup) Patch(path string, handlers ...Handler) *Route {
	return g.Handle(MethodPATCH, path, handlers...)
}

func (g *RouteGroup) Options(path string, handlers ...Handler) *Route {
	return g.Handle(MethodOPTIONS, path, handlers...)
}

func (g *RouteGroup) Connect(path string, handlers ...Handler) *Route {
	return g.Handle(MethodCONNECT, path, handlers...)
}

func (g *RouteGroup) Trace(path string, handlers ...Handler) *Route {
	return g.Handle(MethodTRACE, path, handlers...)
}

func (g *RouteGroup) Brew(path string, handlers ...Handler) *Route {
	return g.Handle(MethodBREW, path, handlers...)
}

func (a *App) addRoute(method Method, path string, groupHandlers []Handler, handlers ...Handler) (*Route, error) {
	if len(handlers) == 0 {
		return nil, errNoHandlers
	}

	normalized := normalizePath(path)
	segments, err := splitRouteSegments(normalized)
	if err != nil {
		return nil, err
	}

	node, err := a.tree.insert(segments)
	if err != nil {
		return nil, err
	}
	if node.routes == nil {
		node.routes = make(map[Method]*routeEntry)
	}
	if _, exists := node.routes[method]; exists {
		return nil, errRouteExists
	}

	entryHandlers := cloneHandlers(handlers)
	entry := &routeEntry{
		handlers:      entryHandlers,
		groupHandlers: cloneHandlers(groupHandlers),
	}
	node.routes[method] = entry
	return &Route{Method: method, Path: normalized}, nil
}

func (a *App) handleRequest(ctx *Context) {
	path := bytesToString(ctx.Path())
	segments := splitRequestPath(path)
	params := ctx.routeParams[:0]

	entry, matched := a.tree.match(ctx.Method(), segments, &params)
	ctx.routeParams = params

	switch {
	case entry != nil:
		a.runHandlers(ctx, entry)
	case matched:
		a.runFallback(ctx, a.methodNotAllowed)
	default:
		a.runFallback(ctx, a.notFound)
	}
}

func (a *App) runHandlers(ctx *Context, entry *routeEntry) {
	pipeline := ctx.handlerPipeline[:0]
	if len(a.globalHandlers) > 0 {
		pipeline = append(pipeline, a.globalHandlers...)
	}
	if len(entry.groupHandlers) > 0 {
		pipeline = append(pipeline, entry.groupHandlers...)
	}
	pipeline = append(pipeline, entry.handlers...)

	ctx.handlerPipeline = pipeline
	ctx.handlers = pipeline
	ctx.handlerIndex = 0
	ctx.Next()
}

func (a *App) runFallback(ctx *Context, fallback Handler) {
	pipeline := ctx.handlerPipeline[:0]
	if len(a.globalHandlers) > 0 {
		pipeline = append(pipeline, a.globalHandlers...)
	}
	pipeline = append(pipeline, fallback)

	ctx.handlerPipeline = pipeline
	ctx.handlers = pipeline
	ctx.handlerIndex = 0
	ctx.Next()
}

func defaultNotFoundHandler(ctx *Context) {
	ctx.WriteFullBodyString(404, "Not Found")
}

func defaultMethodNotAllowedHandler(ctx *Context) {
	ctx.WriteFullBodyString(405, "Method Not Allowed")
}

type routeEntry struct {
	handlers      []Handler
	groupHandlers []Handler
}

type routeNode struct {
	segment string

	staticChildren map[string]*routeNode

	paramChild *routeNode
	paramName  string

	wildcardChild *routeNode
	wildcardName  string

	routes map[Method]*routeEntry
}

func newRouteNode(segment string) *routeNode {
	return &routeNode{
		segment:        segment,
		staticChildren: make(map[string]*routeNode),
	}
}

func (n *routeNode) insert(segments []string) (*routeNode, error) {
	if len(segments) == 0 {
		return n, nil
	}

	segment := segments[0]

	if len(segment) == 0 {
		return nil, errInvalidParameter
	}

	switch segment[0] {
	case ':':
		if len(segment) == 1 {
			return nil, errInvalidParameter
		}
		if n.paramChild == nil {
			n.paramChild = newRouteNode(segment)
			n.paramChild.paramName = segment[1:]
		}
		return n.paramChild.insert(segments[1:])
	case '*':
		if len(segment) == 1 {
			return nil, errInvalidWildcard
		}
		if len(segments) > 1 {
			return nil, errInvalidWildcard
		}
		if n.wildcardChild == nil {
			n.wildcardChild = newRouteNode(segment)
			n.wildcardChild.wildcardName = segment[1:]
		}
		return n.wildcardChild, nil
	default:
		child, ok := n.staticChildren[segment]
		if !ok {
			child = newRouteNode(segment)
			n.staticChildren[segment] = child
		}
		return child.insert(segments[1:])
	}
}

func (n *routeNode) match(method Method, segments []string, params *[]RouteParam) (*routeEntry, bool) {
	if len(segments) == 0 {
		if entry, ok := n.routes[method]; ok {
			return entry, true
		}
		if len(n.routes) > 0 {
			return nil, true
		}
		if n.wildcardChild != nil {
			*params = append(*params, RouteParam{
				Key:   n.wildcardChild.wildcardName,
				Value: "",
			})
			if entry, matched := n.wildcardChild.match(method, nil, params); matched {
				return entry, true
			}
			*params = (*params)[:len(*params)-1]
		}
		return nil, false
	}

	segment := segments[0]

	if child, ok := n.staticChildren[segment]; ok {
		if entry, matched := child.match(method, segments[1:], params); matched {
			return entry, true
		}
	}

	if n.paramChild != nil {
		*params = append(*params, RouteParam{
			Key:   n.paramChild.paramName,
			Value: segment,
		})
		if entry, matched := n.paramChild.match(method, segments[1:], params); matched {
			return entry, true
		}
		*params = (*params)[:len(*params)-1]
	}

	if n.wildcardChild != nil {
		remaining := strings.Join(segments, "/")
		*params = append(*params, RouteParam{
			Key:   n.wildcardChild.wildcardName,
			Value: remaining,
		})
		if entry, matched := n.wildcardChild.match(method, nil, params); matched {
			return entry, true
		}
		*params = (*params)[:len(*params)-1]
	}

	return nil, false
}

func normalizePath(path string) string {
	if path == "" {
		return "/"
	}
	if path[0] != '/' {
		path = "/" + path
	}
	for len(path) > 1 && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}
	return path
}

func normalizeGroupPath(path string) string {
	p := normalizePath(path)
	if p == "/" {
		return ""
	}
	return p
}

func combinePaths(base, path string) string {
	if base == "" {
		return normalizePath(path)
	}
	if path == "" || path == "/" {
		return normalizePath(base)
	}
	return normalizePath(base + "/" + strings.TrimPrefix(path, "/"))
}

func splitRouteSegments(path string) ([]string, error) {
	if path == "/" {
		return nil, nil
	}
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return nil, nil
	}

	raw := strings.Split(trimmed, "/")
	segments := make([]string, 0, len(raw))
	for _, segment := range raw {
		if segment == "" {
			continue
		}
		if segment == ":" || segment == "*" {
			return nil, errInvalidParameter
		}
		segments = append(segments, segment)
	}
	return segments, nil
}

func splitRequestPath(path string) []string {
	if path == "" || path == "/" {
		return nil
	}
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return nil
	}
	raw := strings.Split(trimmed, "/")
	segments := make([]string, 0, len(raw))
	for _, segment := range raw {
		if segment == "" {
			continue
		}
		segments = append(segments, segment)
	}
	return segments
}

func cloneHandlers(src []Handler) []Handler {
	if len(src) == 0 {
		return nil
	}
	dst := make([]Handler, len(src))
	copy(dst, src)
	return dst
}
