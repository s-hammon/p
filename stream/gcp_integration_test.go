//go:build integration
// +build integration

package stream

import (
	"context"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/google/uuid"
	"github.com/s-hammon/p"
	"github.com/s-hammon/p/stream/internal/testproto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func requireEnv(t *testing.T, key string) string {
	t.Helper()

	v := os.Getenv(key)
	if v == "" {
		t.Skipf("%s not set", key)
	}

	return v
}

func TestBigQueryStream_AppendsRows(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping BigQuery integration test")
	}

	project := requireEnv(t, "BQ_PROJECT_ID")
	datasetId := requireEnv(t, "BQ_DATASET")

	ctx := context.Background()

	bq, err := bigquery.NewClient(ctx, project)
	require.NoError(t, err)
	defer bq.Close()

	tableId := "test_rows_" + uuid.NewString()
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

	const rows = 5
	for i := range rows {
		msg := &testproto.TestRow{
			Id:      p.Format("id-%d", i),
			Payload: "test",
			Ts:      time.Now().Unix(),
		}

		data, err := proto.Marshal(msg)
		require.NoError(t, err)

		err = stream.Append(ctx, data)
		require.NoError(t, err)
	}

	err = stream.Shutdown()
	require.NoError(t, err)

	q := bq.Query(p.Format("SELECT COUNT(*) FROM `%s.%s.%s`", project, datasetId, tableId))
	it, err := q.Read(ctx)
	require.NoError(t, err)

	var vals []bigquery.Value
	require.NoError(t, it.Next(&vals))
	require.Equal(t, int64(rows), vals[0])
}
