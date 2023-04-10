// Package errors handles unexpected panics by logging them to the logserver, if configured.
package errs

import (
	"log"
	"runtime"

	"github.com/sc-js/pour"
)

var Signal = make(chan PanicInformation)

// PanicInformation sends a panic message and its stack trace up the Signal
// channel to be safely handled.
type PanicInformation struct {
	RecoveredPanic interface{}
	Stack          string
}

// Bubble sends panic information up the Signal channel. If the traceback is
// empty, this function will collect the stack information.
func Bubble(err interface{}, traceback ...string) {
	if len(traceback) == 0 {
		stack := make([]byte, 1024*8)
		stack = stack[:runtime.Stack(stack, false)]
		traceback = []string{string(stack)}
	}

	pi := PanicInformation{
		RecoveredPanic: err,
		Stack:          traceback[0],
	}

	pour.Log(pi)
}

// Defer is a deferred function that recovers from a panic and Bubble's it
// through the Signal channel.
func Defer() {
	if err := recover(); err != nil {
		log.Println(err)
		Bubble(err)
	}
}
