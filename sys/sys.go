package sys

import (
	"context"
	"fmt"
	"os"
	"os/signal"
)

// IsPipedIn detects whether data is being piped-in from another command or not.
func IsPipedIn() (bool, error) {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false, err
	}
	piped := info.Mode()&os.ModeCharDevice != os.ModeCharDevice && info.Size() > 0
	return piped, nil
}

// ExitIf prints err and exits with an error code if err is not nil, and is not the same as one of the except values.
func ExitIf(err error, except ...error) {
	if err == nil {
		return
	}
	for _, e := range except {
		if err == e {
			return
		}
	}
	fmt.Println(err)
	os.Exit(1)
}

// HandleInterrupt runs the provided cb func after an os.Interrupt or os.Kill signal is received.
func HandleInterrupt(ctx context.Context, cb func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		select {
		case <-c:
			cb()
		case <-ctx.Done():
		}
	}()
}
