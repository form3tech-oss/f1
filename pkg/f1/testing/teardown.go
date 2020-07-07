package testing

type TeardownRegistry struct {
	t       *T
	fnStack []func()
}

func NewTeardownRegistry(t *T) TeardownRegistry {
	return TeardownRegistry{
		fnStack: []func(){},
		t:       t,
	}
}

func (c *TeardownRegistry) RegisterForTeardown(fn func()) {
	c.fnStack = append(c.fnStack, fn)
}

func (c *TeardownRegistry) RegisterForTeardownT(fn func(t *T)) {
	c.fnStack = append(c.fnStack, func() {
		fn(c.t)
	})
}

func (c *TeardownRegistry) Teardown(t *T) {
	for i := len(c.fnStack) - 1; i >= 0; i-- {
		c.fnStack[i]()
	}
}
