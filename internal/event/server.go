package event

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/jcuga/golongpoll"
)

type Server struct {
	poller *golongpoll.LongpollManager
	server *http.Server
	cond   *sync.Cond
	host   string
}

func NewServer(url, pub, sub string) (*Server, error) {
	poller, err := golongpoll.StartLongpoll(golongpoll.Options{
		EventTimeToLiveSeconds: 30,
	})
	if err != nil {
		return nil, fmt.Errorf("polling: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc(sub, poller.SubscriptionHandler)
	mux.HandleFunc(pub, poller.PublishHandler)
	server := &http.Server{Addr: url, Handler: mux}
	cond := sync.NewCond(&sync.Mutex{})

	return &Server{poller: poller, server: server, cond: cond}, nil
}

func (s *Server) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp4", s.server.Addr)
	if err != nil {
		return err
	}

	s.context(ctx)
	s.ready(listener)
	if err := s.server.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	err := s.server.Shutdown(ctx)
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

func (s *Server) Host() string {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()

	for s.host == "" {
		s.cond.Wait()
	}

	return s.host
}

func (s *Server) context(ctx context.Context) {
	s.server.BaseContext = func(_ net.Listener) context.Context {
		return ctx
	}
}

func (s *Server) ready(listener net.Listener) {
	s.cond.L.Lock()
	defer s.cond.Broadcast()
	defer s.cond.L.Unlock()

	s.host = listener.Addr().String()
}

func Publish[T Event](s *Server, id string, data T) error {
	return s.poller.Publish(Mark(id, data.Type()), data)
}
