package deps

type Deps struct {
}

type opt func(*Deps)

func New(opts ...opt) *Deps {
	d := &Deps{}
	for _, o := range opts {
		o(d)
	}
	return d
}
