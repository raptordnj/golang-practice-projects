package classes

type Counter struct {
	count int
}

func (c *Counter) Increment() {
	c.count++
}

func (c Counter) GetCount() int {
	return c.count
}

func (c *Counter) Reset() {
	c.count = 0
}
