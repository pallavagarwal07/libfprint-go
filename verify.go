package fprint

import (
	"fmt"
	"time"
)

type VerifyResult int

const (
	VERIFY_SUCCESS VerifyResult = iota
	VERIFY_FAILED
	VERIFY_TIMEOUT
	VERIFY_ERROR
	VERIFY_SEE_OTHER
)

func (c *Conn) StartVerification(max_tries, timeout_sec int) (<-chan VerifyResult, error) {
	if err := c.attemptVerify(); err != nil {
		return nil, err
	}
	go c.VerifyRoutine(c.verifyResult, max_tries, timeout_sec)
	return c.verifyResult, nil
}

func (c *Conn) attemptVerify() error {
	if c.attempts > 0 {
		if err := c.Call(METHOD_VERIFY_STOP); err != nil {
			return err
		}
	}
	if err := c.Call(METHOD_VERIFY_START, "any"); err != nil {
		return err
	}
	c.attempts += 1
	return nil
}

func (c *Conn) VerifyRoutine(ch chan VerifyResult, tries, timeout_sec int) {
	timeout := time.After(time.Duration(timeout_sec) * time.Second)
	for c.attempts <= tries && !c.finished {
		select {
		case sig := <-c.msgs:
			c.VerifySignals = append(c.VerifySignals, sig)
			if sig.Name != SIGNAL_VERIFY_STATUS {
				ch <- VERIFY_SEE_OTHER
				break
			}
			if sig.Body[0].(string) == "verify-match" {
				ch <- VERIFY_SUCCESS
				c.Close()
			}
			if sig.Body[0].(string) == "verify-no-match" {
				ch <- VERIFY_FAILED
				if err := c.attemptVerify(); err != nil {
					fmt.Println(err)
					ch <- VERIFY_ERROR
					c.Close()
				}
			}
		case <-timeout:
			ch <- VERIFY_TIMEOUT
			c.Close()
		}
	}
	c.Close()
}
