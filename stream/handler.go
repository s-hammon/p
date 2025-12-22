package stream

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/s-hammon/p"
)

// Stream is the interface that wraps Append.
// Append will append a record to a stream destination.
// It returns an error if it is not able to do so.
type Stream interface {
	Append(context.Context, []byte) error
}

type PushEnvelope struct {
	Message      Message `json:"message"`
	Subscription string  `json:"subscription"`
}

type Message struct {
	Data        string            `json:"data"`
	Attributes  map[string]string `json:"attributes"`
	MessageId   string            `json:"messageId"`
	PublishTime string            `json:"publishTime"`
}

type PushHandlerConfig struct {
	// MaxConcurrency caps the number of in-flight requests
	MaxConcurrency int
	// AcquireTimeout sets how long a request waits for a slot.
	// If one is not obtained, the handler returns non-2XX to indicate retry.
	AcquireTimeout time.Duration
	// EnqueueTimeout sets how long a request waits for data to be enqueued to the stream.
	// If the stream is backpressured, the handler returns non-2XX to indicate retry.
	EnqueueTimeout time.Duration
}

func (c PushHandlerConfig) withDefaults() PushHandlerConfig {
	if c.MaxConcurrency <= 0 {
		c.MaxConcurrency = 500
	}
	if c.AcquireTimeout <= 0 {
		c.AcquireTimeout = 2 * time.Second
	}
	if c.EnqueueTimeout <= 0 {
		c.EnqueueTimeout = 2 * time.Second
	}

	return c
}

type RowSerializer func(raw []byte, attrs map[string]string) ([]byte, error)

func NewPushHandler(stream Stream, serialize RowSerializer, cfg PushHandlerConfig) http.HandlerFunc {
	cfg = cfg.withDefaults()
	sem := make(chan struct{}, cfg.MaxConcurrency)

	acquire := func(ctx context.Context) error {
		select {
		default:
		case sem <- struct{}{}:
			return nil
		}

		tctx, cancel := context.WithTimeout(ctx, cfg.AcquireTimeout)
		defer cancel()

		select {
		case <-tctx.Done():
			return tctx.Err()
		case sem <- struct{}{}:
			return nil
		}
	}

	release := func() { <-sem }

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if err := acquire(ctx); err != nil {
			http.Error(w, "busy", http.StatusServiceUnavailable)
			return
		}
		defer release()

		var env PushEnvelope
		if err := json.NewDecoder(r.Body).Decode(&env); err != nil {
			http.Error(w, p.Format("invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		raw, err := base64.StdEncoding.DecodeString(env.Message.Data)
		if err != nil {
			log.Printf("poison message (invalid base64) messageId=%s err=%v\n", env.Message.MessageId, err)
			w.WriteHeader(http.StatusOK)
			return
		}

		row, err := serialize(raw, env.Message.Attributes)
		if err != nil {
			log.Printf("poison message (serialization failed) messageId=%s err=%v\n", env.Message.MessageId, err)
			w.WriteHeader(http.StatusOK)
			return
		}

		qctx, cancel := context.WithTimeout(ctx, cfg.EnqueueTimeout)
		defer cancel()

		if err := stream.Append(qctx, row); err != nil {
			http.Error(w, p.Format("enqueue failed; %v", err), http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
