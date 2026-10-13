package interaction

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/sfossl/bfrm/rest"
	"github.com/sfossl/bfrm/types"
)

type Handler func(*Context) error

type Middleware func(Handler) Handler

type Router struct {
	mu             sync.RWMutex
	middlewares    []Middleware
	commands       map[string]Handler
	subcommands    map[string]Handler
	components     map[string]Handler
	modals         map[string]Handler
	defaultHandler Handler
}

func NewRouter() *Router {
	return &Router{
		commands:    make(map[string]Handler),
		subcommands: make(map[string]Handler),
		components:  make(map[string]Handler),
		modals:      make(map[string]Handler),
	}
}

func (r *Router) Use(mw ...Middleware) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.middlewares = append(r.middlewares, mw...)
}

func (r *Router) OnCommand(name string, h Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.commands[name] = h
}

func (r *Router) OnSubcommand(command, subcommand string, h Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := command + ":" + subcommand
	r.subcommands[key] = h
}

func (r *Router) OnComponent(customID string, h Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.components[customID] = h
}

func (r *Router) OnModal(customID string, h Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.modals[customID] = h
}

func (r *Router) SetDefault(h Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defaultHandler = h
}

func findCustomIDHandler(m map[string]Handler, cid string) Handler {
	if h, ok := m[cid]; ok {
		return h
	}
	prefix, _ := splitCustomID(cid)
	return m[prefix]
}

func (r *Router) findHandler(in *types.Interaction) Handler {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if in.Data == nil {
		return r.defaultHandler
	}

	switch in.Type {
	case types.InteractionTypeApplicationCommand:
		cmdName := in.Data.Name
		for _, opt := range in.Data.Options {
			if opt.Type == 1 {
				subKey := cmdName + ":" + opt.Name
				if h, ok := r.subcommands[subKey]; ok {
					return h
				}
			}
		}
		if h, ok := r.commands[cmdName]; ok {
			return h
		}

	case types.InteractionTypeMessageComponent:
		if h := findCustomIDHandler(r.components, in.Data.CustomID); h != nil {
			return h
		}

	case types.InteractionTypeModalSubmit:
		if h := findCustomIDHandler(r.modals, in.Data.CustomID); h != nil {
			return h
		}
	}

	return r.defaultHandler
}

func (r *Router) Handle(ctx *Context) error {
	h := r.findHandler(ctx.Interaction)
	if h == nil {
		return nil
	}

	r.mu.RLock()
	mws := make([]Middleware, len(r.middlewares))
	copy(mws, r.middlewares)
	r.mu.RUnlock()

	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}

	return h(ctx)
}

func (r *Router) WebhookHandler(restClient *rest.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var in types.Interaction
		if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if in.Type == types.InteractionTypePing {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"type":1}`))
			return
		}

		ctx := NewContext(context.WithoutCancel(req.Context()), &in, restClient)

		errChan := make(chan error, 1)
		go func() {
			errChan <- r.Handle(ctx)
		}()

		select {
		case res := <-ctx.ResponseChannel():
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(res)
		case err := <-errChan:
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			select {
			case res := <-ctx.ResponseChannel():
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(res)
			default:
				w.WriteHeader(http.StatusOK)
			}
		case <-time.After(2800 * time.Millisecond):
			_ = ctx.Defer(false)
			select {
			case res := <-ctx.ResponseChannel():
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(res)
			default:
				w.WriteHeader(http.StatusOK)
			}
		}
	})
}

