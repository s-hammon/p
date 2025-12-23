package stream

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"cloud.google.com/go/bigquery/storage/managedwriter"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/descriptorpb"
)

type BigQueryStream struct {
	client *managedwriter.Client
	ms     *managedwriter.ManagedStream

	ch chan []byte

	wg     sync.WaitGroup
	cancel context.CancelFunc

	batchSize                    int
	flushInterval, appendTimeout time.Duration

	errMu sync.Mutex
	errs  []error
}

type BigQueryStreamConfig struct {
	BatchSize     int
	ChannelSize   int
	FlushInterval time.Duration
	AppendTimeout time.Duration

	clientOpts []option.ClientOption
}

// NOTE: multiplexing with the managed writer is an experimental feature
func (c BigQueryStreamConfig) AsMultiplexer(limit ...int) {
	l := 10
	if len(limit) > 0 {
		l = limit[0]
	}

	c.clientOpts = append(c.clientOpts,
		managedwriter.WithMultiplexing(),
		managedwriter.WithMultiplexPoolLimit(l),
	)
}

func (c BigQueryStreamConfig) withDefaults() BigQueryStreamConfig {
	if c.BatchSize <= 0 {
		c.BatchSize = 500
	}
	if c.ChannelSize <= 0 {
		c.ChannelSize = 10000
	}
	if c.FlushInterval <= 0 {
		c.FlushInterval = 50 * time.Millisecond
	}
	if c.AppendTimeout <= 0 {
		c.AppendTimeout = 15 * time.Second
	}

	return c
}

type Options func() *managedwriter.WriterOption

func CommittedStreamOpts(tableName string, descriptor *descriptorpb.DescriptorProto) (opts []managedwriter.WriterOption) {
	opts = append(opts,
		managedwriter.WithDestinationTable(tableName),
		managedwriter.WithType(managedwriter.CommittedStream),
		managedwriter.WithSchemaDescriptor(descriptor),
	)

	return opts
}

func NewBigQueryStream(ctx context.Context, projectId string, cfg BigQueryStreamConfig, opts ...managedwriter.WriterOption) (*BigQueryStream, error) {
	if len(opts) < 1 {
		return nil, errors.New("please provide options for stream (use CommittedStreamOpts)")
	}
	cfg = cfg.withDefaults()

	client, err := managedwriter.NewClient(ctx, projectId)
	if err != nil {
		return nil, fmt.Errorf("managedwriter.NewClient: %w", err)
	}

	ms, err := client.NewManagedStream(ctx, opts...)
	if err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("client.NewManagedStream: %w", err)
	}

	runCtx, cancel := context.WithCancel(context.Background())

	s := &BigQueryStream{
		client:        client,
		ms:            ms,
		ch:            make(chan []byte, cfg.ChannelSize),
		cancel:        cancel,
		batchSize:     cfg.BatchSize,
		flushInterval: cfg.FlushInterval,
		appendTimeout: cfg.AppendTimeout,
	}

	s.wg.Add(1)
	go s.writerLoop(runCtx)

	return s, nil
}

func (s *BigQueryStream) Append(ctx context.Context, row []byte) error {
	select {
	case s.ch <- row:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *BigQueryStream) Stop() error {
	s.cancel()
	close(s.ch)
	s.wg.Wait()
	return nil
}

func (s *BigQueryStream) Close() error {
	var err1, err2 error
	if s.ms != nil {
		err1 = s.ms.Close()
	}
	if s.client != nil {
		err2 = s.client.Close()
	}

	return errors.Join(err1, err2)
}

func (s *BigQueryStream) Shutdown() error {
	_ = s.Stop()

	closeErr := s.Close()
	s.errMu.Lock()
	defer s.errMu.Unlock()
	if len(s.errs) == 0 {
		return closeErr
	}

	all := make([]error, 0, len(s.errs)+1)
	all = append(all, s.errs...)
	if closeErr != nil {
		all = append(all, closeErr)
	}

	return errors.Join(all...)
}

func (s *BigQueryStream) writerLoop(ctx context.Context) {
	defer s.wg.Done()

	t := time.NewTicker(s.flushInterval)
	defer t.Stop()

	buf := make([][]byte, 0, s.batchSize)

	flush := func() {
		if len(buf) == 0 {
			return
		}

		batch := make([][]byte, len(buf))
		copy(batch, buf)
		buf = buf[:0]

		appendCtx, cancel := context.WithTimeout(context.Background(), s.appendTimeout)
		defer cancel()

		if _, err := s.ms.AppendRows(appendCtx, batch); err != nil {
			switch classifyStreamError(err) {
			default:
			case StreamFatal:
				s.recordErr(err)
				s.cancel()
				return
			case StreamRetryable:
				s.recordErr(err)
				buf = append(batch, buf...)
				return
			}
		}
	}

	handleRow := func(row []byte) {
		buf = append(buf, row)
		if len(buf) >= s.batchSize {
			flush()
		}
	}

	for {
		select {
		case <-t.C:
			flush()
		case <-ctx.Done():
			for {
				select {
				default:
					flush()
					return
				// handle any remaining records
				case row, ok := <-s.ch:
					if !ok {
						flush()
						return
					}

					handleRow(row)
				}
			}
		case row, ok := <-s.ch:
			if !ok {
				flush()
				return
			}

			handleRow(row)
		}
	}
}

func (s *BigQueryStream) recordErr(err error) {
	if err == nil {
		return
	}

	s.errMu.Lock()
	s.errs = append(s.errs, err)
	s.errMu.Unlock()
	log.Printf("bigquery stream error: %v\n", err)
}

type StreamOutcome int

const (
	StreamOK StreamOutcome = iota
	StreamRetryable
	StreamFatal
)

func classifyStreamError(err error) StreamOutcome {
	if err == nil {
		return StreamOK
	}

	st, ok := status.FromError(err)
	if !ok {
		return StreamRetryable
	}

	switch st.Code() {
	default:
		return StreamRetryable
	case codes.Unavailable,
		codes.DeadlineExceeded,
		codes.ResourceExhausted,
		codes.Aborted,
		codes.Internal:
		return StreamRetryable
	case codes.InvalidArgument,
		codes.FailedPrecondition,
		codes.PermissionDenied,
		codes.NotFound,
		codes.Unauthenticated:
		return StreamFatal
	}
}
