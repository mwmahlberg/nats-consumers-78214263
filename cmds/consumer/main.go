package main

import (
	"context"
	"net/url"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/nats-io/nats.go"
)

var consumercfg struct {
	NatsURL   url.URL `kong:"name='nats-url',help='NATS server URL',default='nats://nats:4222'"`
	Topic     string  `kong:"name='topic',help='NATS topic to subscribe to',default='test.>'"`
	Consumers int     `kong:"name='consumers',help='Number of consumers to start',default='1'"`
}

func main() {
	ctx := kong.Parse(&consumercfg, kong.DefaultEnvars("CONSUMER"))
	ctx.Printf("Starting consumer on %s, subscribing to %s", consumercfg.NatsURL.String(), consumercfg.Topic)

	nc, err := nats.Connect(consumercfg.NatsURL.String())
	ctx.FatalIfErrorf(err, "Could not connect to NATS server: %s", consumercfg.NatsURL.String())
	// Run the configured number of consumers in goroutines
	// Note that all consumers share the same NATS connection
	// Each consumer subscribes to the configured topic
	// and counts the number of messages received, printing them out every second.
	// The consumers will stop when SIGINT or SIGTERM are received.

	sigs, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	for i := 0; i < consumercfg.Consumers; i++ {
		go func(sigs context.Context, conn *nats.Conn, topic string, id int) {

			count := atomic.Int64{}

			// We use the same connection!
			sub, err := conn.Subscribe(topic, func(msg *nats.Msg) {
				// Callback for processing a new message.
				count.Add(1)
			})
			ctx.FatalIfErrorf(err, "Could not subscribe to topic %s: %s", topic, err)
			defer sub.Unsubscribe()

			tick := time.NewTicker(time.Second)
			for {
				select {
				case <-sigs.Done():
					ctx.Printf("Received shutdown signal.")
					ctx.Printf("Final result: received %d messages", count.Load())
					return
				case <-tick.C:
					ctx.Printf("%6d Received %d messages", id, count.Load())
				}
			}
		}(sigs, nc, consumercfg.Topic, i)
	}

	<-sigs.Done()
}
