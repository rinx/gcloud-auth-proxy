package idtoken

import "time"

type Option interface {
	Apply(s *server) error
}

type defaultAudienceOption string

func (o defaultAudienceOption) Apply(s *server) error {
	s.defaultAudience = string(o)

	return nil
}

func WithDefaultAudience(audience string) defaultAudienceOption {
	return defaultAudienceOption(audience)
}

type tsCacheDurationOption string

func (o tsCacheDurationOption) Apply(s *server) error {
	dur, err := time.ParseDuration(string(o))
	if err != nil {
		return err
	}

	s.tsCacheDuration = dur

	return nil
}

func WithTokenSourceCacheDuration(dur string) tsCacheDurationOption {
	return tsCacheDurationOption(dur)
}

type debugGoproxyOption bool

func (o debugGoproxyOption) Apply(s *server) error {
	s.proxy.Verbose = bool(o)

	return nil
}

func WithDebugGoproxy(b bool) debugGoproxyOption {
	return debugGoproxyOption(b)
}
