package foundation

type Option func(*Service)

func Name(name string) Option {
	return func(s *Service) {
		s.name = name
	}
}
