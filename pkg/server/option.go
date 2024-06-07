package server

import "net/http"

type Option interface {
	Apply(s *server) error
}

type hostOption string

func (o hostOption) Apply(s *server) error {
	s.host = string(o)

	return nil
}

func WithHost(host string) hostOption {
	return hostOption(host)
}

type portOption string

func (o portOption) Apply(s *server) error {
	s.port = string(o)

	return nil
}

func WithPort(port string) portOption {
	return portOption(port)
}

type handlerOption struct {
	handler http.Handler
}

func (o handlerOption) Apply(s *server) error {
	s.handler = http.Handler(o.handler)

	return nil
}

func WithHandler(handler http.Handler) handlerOption {
	return handlerOption{
		handler: handler,
	}
}
