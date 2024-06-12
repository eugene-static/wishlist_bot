package bot

import (
	"context"
)

const DefaultMessage = "default"

type Handler interface {
	ServeBot(ctx context.Context, r *Request)
}

type HandlerFunc func(ctx context.Context, r *Request)

func (hf HandlerFunc) ServeBot(ctx context.Context, r *Request) {
	hf(ctx, r)
}

type Mux struct {
	m map[string]HandlerFunc
}

func NewBotMux() *Mux {
	return &Mux{
		m: make(map[string]HandlerFunc),
	}
}

func (m *Mux) Handle(pattern string, handler HandlerFunc) {
	m.m[pattern] = handler
}

func (m *Mux) Handler(pattern string) Handler {
	if fn, ok := m.m[pattern]; ok {
		return fn
	}
	return nil
}

func (m *Mux) ServeBot(ctx context.Context, r *Request) {
	if f, ok := m.m[r.Data]; ok {
		f(ctx, r)
	} else {
		m.m[DefaultMessage](ctx, r)
	}
}
