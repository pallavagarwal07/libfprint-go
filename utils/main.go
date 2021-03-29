package main

import (
	"fmt"
	"log"

	"github.com/pallavagarwal07/libfprint-go/fprint"
)

func app() error {
	conn, err := fprint.NewConn( /* empty indicates current user */ "")
	if err != nil {
		return err
	}
	defer conn.Close()

	fingers, err := conn.ListEnrolledFingers()
	if err != nil {
		return err
	}
	fmt.Println(fingers)

	ch, err := conn.StartEnroll("right-index-finger")
	if err != nil {
		return err
	}

	return nil
}

func main() {
	if err := app(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
