package interaction

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
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

func (r *Router) OnSubcommandGroup(command, group, subcommand string, h Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := command + ":" + group + ":" + subcommand
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
	case types.InteractionTypeApplicationCommand, types.InteractionTypeApplicationCommandAutocomplete:
		cmdName := in.Data.Name
		for _, opt := range in.Data.Options {
			switch opt.Type {
			case int(types.OptionTypeSubcommand):
				subKey := cmdName + ":" + opt.Name
				if h, ok := r.subcommands[subKey]; ok {
					return h
				}
			case int(types.OptionTypeSubcommandGroup):
				for _, subOpt := range opt.Options {
					if subOpt.Type == int(types.OptionTypeSubcommand) {
						groupKey := cmdName + ":" + opt.Name + ":" + subOpt.Name
						if h, ok := r.subcommands[groupKey]; ok {
							return h
						}
					}
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

func (r *Router) WebhookHandler(restClient *rest.Client, autoDefer ...bool) http.Handler {
	shouldAutoDefer := len(autoDefer) > 0 && autoDefer[0]
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

		handlerCtx, cancel := context.WithTimeout(context.WithoutCancel(req.Context()), 15*time.Minute)
		ctx := NewContext(handlerCtx, &in, restClient)

		errChan := make(chan error, 1)
		go func() {
			defer func() {
				if rec := recover(); rec != nil {
					cancel()
					errChan <- fmt.Errorf("panic in interaction handler: %v", rec)
				}
			}()
			err := r.Handle(ctx)
			if err != nil {
				cancel()
			}
			errChan <- err
		}()

		var timeoutChan <-chan time.Time
		if shouldAutoDefer {
			timer := time.NewTimer(2500 * time.Millisecond)
			defer timer.Stop()
			timeoutChan = timer.C
		}

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
				w.WriteHeader(http.StatusInternalServerError)
			}
		case <-timeoutChan:
			_ = ctx.Defer(false)
			select {
			case res := <-ctx.ResponseChannel():
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(res)
			default:
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(types.InteractionResponse{
					Type: types.CallbackTypeDeferredChannelMessageWithSource,
				})
			}
		case <-req.Context().Done():
			w.WriteHeader(http.StatusGatewayTimeout)
		}
	})
}
