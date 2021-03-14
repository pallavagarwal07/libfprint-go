package fprint

type EnrollResult int

const (
	ENROLL_SUCCESS EnrollResult = iota
	ENROLL_FAILED
	ENROLL_STAGE_PASSED
	ENROLL_RETRY_SCAN
	ENROLL_ERROR
	ENROLL_SEE_OTHER
)

func (c *Conn) StartEnroll(which string) (<-chan EnrollResult, error) {
	if err := c.Call(METHOD_ENROLL_START, which); err != nil {
		return nil, err
	}
	go c.EnrollRoutine(c.enrollResult)
	return c.enrollResult, nil
}

func (c *Conn) EnrollRoutine(ch chan<- EnrollResult) {
	for sig := range c.msgs {
		result := ENROLL_SUCCESS
		c.EnrollSignals = append(c.EnrollSignals, sig)

		switch sig.Name {
		case SIGNAL_ENROLL_STATUS:
			switch sig.Body[0].(string) {
			case "enroll-completed":
				result = ENROLL_SUCCESS
			case "enroll-failed":
				result = ENROLL_FAILED
			case "enroll-stage-passed":
				result = ENROLL_STAGE_PASSED
			case "enroll-remove-and-retry", "enroll-retry-scan",
				"enroll-swipe-too-short", "enroll-finger-not-centered":
				result = ENROLL_RETRY_SCAN
			case "enroll-data-full", "enroll-disconnected", "enroll-unknown-error":
				result = ENROLL_ERROR
			}
		default:
			result = ENROLL_SEE_OTHER
		}
		ch <- result
		if result != ENROLL_STAGE_PASSED && result != ENROLL_RETRY_SCAN {
			c.Call(METHOD_ENROLL_STOP)
			close(ch)
			return
		}
	}
	c.Call(METHOD_ENROLL_STOP)
	close(ch)
}
