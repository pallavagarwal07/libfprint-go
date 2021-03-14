package fprint

func (c *Conn) ListEnrolledFingers() ([]string, error) {
	var names []string
	err := c.object.Call(METHOD_LIST_ENROLLED_FINGERS, 0, c.user).Store(&names)
	return names, err
}

func (c *Conn) DeleteEnrolledFingers() error {
	return c.Call(METHOD_DELETE_ENROLLED_FINGERS)
}

func (c *Conn) DeleteEnrolledFinger(finger string) error {
	return c.Call(METHOD_DELETE_ENROLLED_FINGER, finger)
}
