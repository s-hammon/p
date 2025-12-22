package stream

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mockStream struct {
	mu        sync.Mutex
	calls     int
	blockCh   chan struct{}
	returnErr error
}

func newMockStream() *mockStream {
	return &mockStream{
		blockCh: make(chan struct{}),
	}
}

func (s *mockStream) Append(ctx context.Context, row []byte) error {
	s.mu.Lock()
	s.calls++
	s.mu.Unlock()

	select {
	case <-s.blockCh:
		return s.returnErr
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *mockStream) unblock() {
	close(s.blockCh)
}

func (s *mockStream) callCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

func requestBody(t *testing.T, payload []byte) io.Reader {
	t.Helper()

	env := map[string]any{
		"message": map[string]any{
			"data": base64.StdEncoding.EncodeToString(payload),
		},
	}

	b, err := json.Marshal(env)
	require.NoError(t, err)
	return bytes.NewReader(b)
}

func TestPushHandler_SuccessfulEnqueue(t *testing.T) {
	stream := newMockStream()
	stream.unblock()

	serializer := func(raw []byte, _ map[string]string) ([]byte, error) {
		return []byte("ok"), nil
	}

	h := NewPushHandler(stream, serializer, PushHandlerConfig{MaxConcurrency: 10})
	rec := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodPost, "/", requestBody(t, []byte("test")))
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, 1, stream.callCount())
}

func TestPushHandler_PoisonPayload(t *testing.T) {
	stream := newMockStream()

	serializer := func(raw []byte, _ map[string]string) ([]byte, error) {
		return nil, errors.New("invalid payload")
	}

	h := NewPushHandler(stream, serializer, PushHandlerConfig{})
	rec := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodPost, "/", requestBody(t, []byte("test")))
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, 0, stream.callCount())
}

func TestPushHandler_EnqueueTimeout(t *testing.T) {
	stream := newMockStream()

	serializer := func(raw []byte, _ map[string]string) ([]byte, error) {
		return []byte("ok"), nil
	}

	h := NewPushHandler(stream, serializer, PushHandlerConfig{EnqueueTimeout: 10 * time.Millisecond})
	rec := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodPost, "/", requestBody(t, []byte("test")))
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

func TestPushHandler_MaxConcurrency(t *testing.T) {
	stream := newMockStream()

	serializer := func(raw []byte, _ map[string]string) ([]byte, error) {
		return []byte("ok"), nil
	}

	h := NewPushHandler(stream, serializer, PushHandlerConfig{
		MaxConcurrency: 1,
		AcquireTimeout: 20 * time.Millisecond,
	})

	req1 := httptest.NewRequest(http.MethodPost, "/", requestBody(t, []byte("test1")))
	rec1 := httptest.NewRecorder()
	go h.ServeHTTP(rec1, req1)

	req2 := httptest.NewRequest(http.MethodPost, "/", requestBody(t, []byte("test2")))
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)

	require.Equal(t, http.StatusServiceUnavailable, rec2.Code)
}

func TestPushHandler_InvalidRequest(t *testing.T) {
	stream := newMockStream()

	serializer := func(raw []byte, _ map[string]string) ([]byte, error) {
		return nil, nil
	}

	h := NewPushHandler(stream, serializer, PushHandlerConfig{})
	rec := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("non-json"))
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}
