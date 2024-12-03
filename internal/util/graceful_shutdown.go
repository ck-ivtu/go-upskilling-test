package util

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Operation func(ctx context.Context) error

func GracefulShutdown(ctx context.Context, timeout time.Duration, ops map[string]Operation) <-chan struct{} {
	wait := make(chan struct{})

	go func() {
		s := make(chan os.Signal, 1)

		signal.Notify(s, syscall.SIGTERM, syscall.SIGINT)
		defer signal.Stop(s)

		<-s

		fmt.Fprintf(os.Stdout, "shutting down")

		timeoutFunc := time.AfterFunc(timeout, func() {
			fmt.Fprintf(os.Stdout, "timeout %d ms has been elapsed, force exit", timeout.Milliseconds())
			os.Exit(0)
		})
		defer timeoutFunc.Stop()

		var wg sync.WaitGroup

		for key, op := range ops {
			wg.Add(1)
			innerKey := key
			innerOp := op

			go func() {
				defer wg.Done()

				fmt.Fprintf(os.Stdout, "cleaning up: %s", innerKey)
				if err := innerOp(ctx); err != nil {
					fmt.Fprintf(os.Stdout, "%s: clean up failed: %s", innerKey, err.Error())
					return
				}
				fmt.Fprintf(os.Stdout, "%s was shutdown gracefully", innerKey)
			}()
		}

		wg.Wait()
		close(wait)
	}()

	return wait
}
