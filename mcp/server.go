package mcp

import (
	"flag"
	"fmt"
	"time"

	"github.com/sedyh/ebitengine-mcp/internal/cli"
	"github.com/sedyh/ebitengine-mcp/internal/event"
)

type Server struct {
	responser *event.Client[*event.RecordResponse]
	requester *event.Client[*event.RecordRequest]
	requests  <-chan *event.RecordRequest
	id        string
}

func NewServer() (*Server, error) {
	url := flag.String(cli.FlagURL, "", "url to run the server")
	pub := flag.String(cli.FlagPub, "", "pub to run the server")
	sub := flag.String(cli.FlagSub, "", "sub to run the server")
	id := flag.String(cli.FlagID, "", "id to run the server")
	flag.Parse()

	responser, err := event.NewClient[*event.RecordResponse](*url, *pub, *sub, *id)
	if err != nil {
		return nil, fmt.Errorf("responser: %w", err)
	}

	requester, err := event.NewClient[*event.RecordRequest](*url, *pub, *sub, *id)
	if err != nil {
		return nil, fmt.Errorf("requester: %w", err)
	}

	requests := requester.Start(time.Now().Add(-2 * time.Minute))

	return &Server{
		requests:  requests,
		requester: requester,
		responser: responser,
		id:        *id,
	}, nil
}

func (s *Server) Close() {
	s.requester.Stop()
}

func (s *Server) RecordRequests() <-chan *event.RecordRequest {
	return s.requests
}

func (s *Server) RecordResponce(res *event.RecordResponse) error {
	return s.responser.Publish(s.id, res)
}
