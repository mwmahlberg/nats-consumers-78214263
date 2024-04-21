package main

import (
	"context"
	"fmt"
	"net/url"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/nats-io/nats.go"
)

var producercfg struct {
	NatsURL   url.URL `kong:"name='nats-url',help='NATS server URL',default='nats://nats:4222'"`
	Producers int     `kong:"name='producers',help='Number of producers to start',default='1'"`
}

func main() {
	ctx := kong.Parse(&producercfg, kong.DefaultEnvars("PRODUCER"))

	// Run the configured number of producers in goroutines
	// Note that all producers share the same NATS connection
	// Each producer sends a messsage every 100ms

	nc, err := nats.Connect(producercfg.NatsURL.String())
	ctx.FatalIfErrorf(err, "Could not connect to NATS server: %s", producercfg.NatsURL.String())
	defer nc.Close()

	// Handle SIGINT and SIGTERM to shut down gracefully
	// We use a context here because that makes it easy for us to shut down
	// all goroutines in one fell swoop, but gracefully so.
	sigs, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var wg sync.WaitGroup
	var sent atomic.Int64

	for i := 0; i < producercfg.Producers; i++ {
		wg.Add(1)
		go func(producerContext context.Context, conn *nats.Conn, id int) {
			ctx.Printf("Starting publisher to %s", fmt.Sprintf("test.%d", id))
			defer wg.Done()

			for {
				// We have...
				select {
				// either received a signal to shut down...
				case <-producerContext.Done():
					ctx.Printf("Producer %d shutting down", id)
					// ... so we return from the goroutine.
					return
				default:
					// or we send a message.
					sent.Add(1)
					err := conn.Publish(fmt.Sprintf("test.%d", id), []byte("Hello, World!"))
					ctx.FatalIfErrorf(err, "Could not publish message: %s", err)
				}
			}
		}(sigs, nc, i)
	}

	tick := time.NewTicker(time.Second)

evt:
	for {
		// Either we receive a signal to shut down...
		select {
		case <-sigs.Done():
			cancel()
			break evt
		// ... or we print out the number of messages sent so far.
		case <-tick.C:
			ctx.Printf("Sent %d messages", sent.Load())
		}
	}
	ctx.Printf("Received signal, shutting down producers...")
	wg.Wait()
	ctx.Printf("All producers shut down. Exiting.")
}
