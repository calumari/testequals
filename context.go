package testequals

type cmpCtx struct {
	path       []string
	collect    bool
	mismatches []*MismatchError
}

func (c *cmpCtx) report(m *MismatchError) error {
	if c.collect {
		c.mismatches = append(c.mismatches, m)
		return nil
	}
	return m
}

func (c *cmpCtx) reportAt(seg, msg string) error {
	c.push(seg)
	m := mismatch(c.path, msg)
	c.pop()
	return c.report(m)
}

func (c *cmpCtx) push(seg string) {
	c.path = append(c.path, seg)
}

func (c *cmpCtx) pop() {
	c.path = c.path[:len(c.path)-1]
}
