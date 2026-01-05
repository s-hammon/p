//go:build integration
// +build integration

package stream

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/google/uuid"
	"github.com/s-hammon/p"
	"github.com/s-hammon/p/stream/internal/testproto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestPushHandler_BigQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	project := requireEnv(t, "BQ_PROJECT_ID")
	datasetId := requireEnv(t, "BQ_DATASET")

	ctx := context.Background()

	bq, err := bigquery.NewClient(ctx, project)
	require.NoError(t, err)
	defer bq.Close()

	tableId := "push_e2e_" + uuid.NewString()
	table := bq.Dataset(datasetId).Table(tableId)

	err = table.Create(ctx, &bigquery.TableMetadata{
		Schema: bigquery.Schema{
			{Name: "id", Type: bigquery.StringFieldType},
			{Name: "payload", Type: bigquery.StringFieldType},
			{Name: "ts", Type: bigquery.IntegerFieldType},
		},
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = table.Delete(context.Background())
	})

	desc := (&testproto.TestRow{}).ProtoReflect().Descriptor()

	stream, err := NewBigQueryStream(
		ctx,
		project,
		BigQueryStreamConfig{
			BatchSize:     10,
			FlushInterval: 50 * time.Millisecond,
		},
		CommittedStreamOpts(
			p.Format("projects/%s/datasets/%s/tables/%s", project, datasetId, tableId),
			desc,
		)...,
	)
	require.NoError(t, err)

	serializer := func(raw []byte, _ map[string]string) ([]byte, error) {
		msg := &testproto.TestRow{
			Id:      uuid.NewString(),
			Payload: string(raw),
			Ts:      time.Now().Unix(),
		}

		return proto.Marshal(msg)
	}

	h := NewPushHandler(stream, serializer, PushHandlerConfig{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", requestBody(t, []byte("test")))

	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.NoError(t, stream.Shutdown())

	q := bq.Query(p.Format("SELECT COUNT(*) FROM `%s.%s.%s`", project, datasetId, tableId))
	it, err := q.Read(ctx)
	require.NoError(t, err)

	var vals []bigquery.Value
	err = it.Next(&vals)
	require.NoError(t, err)
	require.Equal(t, int64(1), vals[0])
}
