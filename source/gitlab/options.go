package gitlab

type ClientOption func(*Source)

func DefaultToken(token string) (opt ClientOption) {
	return func(s *Source) {
		s.token = token
	}
}
