package pdf

type str struct {
	d []string
}

func (s *str) pop() string {
	var v string
	v, s.d = s.d[0], s.d[1:]
	return v
}

func (s *str) popat(i int) string {
	var v = s.d[i]
	s.d = append(s.d[:i], s.d[i+1:]...)
	return v
}

func (s *str) len() int {
	return len(s.d)
}
