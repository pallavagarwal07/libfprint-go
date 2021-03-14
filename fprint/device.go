package fprint

import (
	"github.com/godbus/dbus/v5"
)

const (
	OBJ_PATH = "/net/reactivated/Fprint/Device/0"
	OBJ_NAME = "net.reactivated.Fprint"

	METHOD_LIST_ENROLLED_FINGERS   = OBJ_NAME + ".Device.ListEnrolledFingers"
	METHOD_DELETE_ENROLLED_FINGERS = OBJ_NAME + ".Device.DeleteEnrolledFingers2"
	METHOD_DELETE_ENROLLED_FINGER  = OBJ_NAME + ".Device.DeleteEnrolledFinger"
	METHOD_CLAIM                   = OBJ_NAME + ".Device.Claim"
	METHOD_RELEASE                 = OBJ_NAME + ".Device.Release"
	METHOD_VERIFY_START            = OBJ_NAME + ".Device.VerifyStart"
	METHOD_VERIFY_STOP             = OBJ_NAME + ".Device.VerifyStop"
	METHOD_ENROLL_START            = OBJ_NAME + ".Device.EnrollStart"
	METHOD_ENROLL_STOP             = OBJ_NAME + ".Device.EnrollStop"

	SIGNAL_VERIFY_FINGER_SELECTED = OBJ_NAME + ".Device.VerifyFingerSelected"
	SIGNAL_VERIFY_STATUS          = OBJ_NAME + ".Device.VerifyStatus"
	SIGNAL_ENROLL_STATUS          = OBJ_NAME + ".Device.EnrollStatus"
)

type Conn struct {
	conn         *dbus.Conn
	msgs         chan *dbus.Signal
	verifyResult chan VerifyResult
	enrollResult chan EnrollResult
	object       dbus.BusObject
	user         string
	attempts     int
	finished     bool

	VerifySignals []*dbus.Signal
	EnrollSignals []*dbus.Signal
}

func NewConn(user string) (*Conn, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}

	match := dbus.WithMatchObjectPath(OBJ_PATH)
	if err := conn.AddMatchSignal(match); err != nil {
		return nil, err
	}
	ch := make(chan *dbus.Signal, 20)
	conn.Signal(ch)

	ret := &Conn{
		conn:         conn,
		msgs:         ch,
		user:         user,
		object:       conn.Object(OBJ_NAME, OBJ_PATH),
		verifyResult: make(chan VerifyResult, 20),
		enrollResult: make(chan EnrollResult, 20),
	}
	if err := ret.Call(METHOD_CLAIM, user); err != nil {
		return nil, err
	}
	return ret, nil
}

func (c *Conn) Close() {
	if c.finished {
		return
	}
	c.finished = true

	close(c.verifyResult)
	c.Call(METHOD_VERIFY_STOP)
	c.Call(METHOD_RELEASE)
}

func (c *Conn) Call(method string, args ...interface{}) error {
	call := c.object.Call(method, 0, args...)
	return call.Err
}
