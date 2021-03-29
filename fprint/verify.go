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
	VERIFY_MAX_TRIES
	VERIFY_SEE_OTHER
)

func (c *Conn) StartVerification(max_tries, timeout_sec int) (<-chan VerifyResult, error) {
	if err := c.attemptVerify(); err != nil {
		return nil, err
	}
	go c.VerifyRoutine(max_tries, timeout_sec)
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

func (c *Conn) VerifyRoutine(tries, timeout_sec int) {
	timeout := time.After(time.Duration(timeout_sec) * time.Second)
	for c.attempts <= tries && !c.finished {
		select {
		case sig := <-c.msgs:
			if c.finished {
				return
			}
			c.VerifySignals = append(c.VerifySignals, sig)
			if sig.Name != SIGNAL_VERIFY_STATUS {
				c.verifyResult <- VERIFY_SEE_OTHER
				break
			}
			if sig.Body[0].(string) == "verify-match" {
				c.verifyResult <- VERIFY_SUCCESS
				c.Close()
			}
			if sig.Body[0].(string) == "verify-no-match" {
				c.verifyResult <- VERIFY_FAILED
				if err := c.attemptVerify(); err != nil {
					fmt.Println(err)
					c.verifyResult <- VERIFY_ERROR
					c.Close()
				}
			}
		case <-timeout:
			c.verifyResult <- VERIFY_TIMEOUT
			c.Close()
		}
	}
	if !c.finished {
		c.verifyResult <- VERIFY_MAX_TRIES
	}
	c.Close()
}
