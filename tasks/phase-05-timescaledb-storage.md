# Phase 5: TimescaleDB Integration — Persistent Storage

> In-memory storage disappears when the process restarts. This phase adds a real database. TimescaleDB is PostgreSQL with a time-series extension — it automatically partitions your data by time, making queries like "show me all spans from the last hour" dramatically faster. You'll also learn how Go handles database connections, which is very different from Python's SQLAlchemy.

---

## Learning Objectives

By the end of this phase you will understand:

- How `database/sql` works in Go — connection pools, not single connections
- Why TimescaleDB hypertables are better than plain PostgreSQL for time-series data
- How to write safe parameterized queries (preventing SQL injection)
- How connection pooling works (`MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime`)
- How to use `pgx` — a high-performance PostgreSQL driver for Go
- How bulk inserts (`COPY`) dramatically outperform row-by-row inserts

---

## Tasks

1. **Start TimescaleDB with Docker:**
   ```bash
   docker run -d \
     --name infertrace-db \
     -e POSTGRES_PASSWORD=infertrace \
     -e POSTGRES_DB=infertrace \
     -p 5432:5432 \
     timescale/timescaledb:latest-pg16
   ```

2. **Create the schema** — connect to the DB and run:
   ```sql
   CREATE TABLE spans (
       trace_id        TEXT NOT NULL,
       span_id         TEXT NOT NULL,
       parent_span_id  TEXT,
       service_name    TEXT NOT NULL,
       operation_name  TEXT NOT NULL,
       start_time      TIMESTAMPTZ NOT NULL,
       duration_nanos  BIGINT NOT NULL,
       status          TEXT,
       model_name      TEXT,
       model_version   TEXT,
       batch_size      INT,
       input_tokens    INT,
       output_tokens   INT,
       gpu_utilization DOUBLE PRECISION,
       tags            JSONB
   );

   -- Convert to a TimescaleDB hypertable partitioned by time
   SELECT create_hypertable('spans', 'start_time');

   -- Indexes for common query patterns
   CREATE INDEX ON spans (trace_id, start_time DESC);
   CREATE INDEX ON spans (model_name, start_time DESC);
   ```
   Understand what `create_hypertable` does: it splits the table into chunks by time period automatically. Queries with a time filter only scan relevant chunks.

3. **Install the pgx driver:**
   ```bash
   go get github.com/jackc/pgx/v5
   go get github.com/jackc/pgx/v5/pgxpool
   ```
   Use `pgxpool` — it manages a pool of connections so multiple goroutines can query concurrently without waiting for each other.

4. **Implement `TimescaleStore`** in `internal/storage/timescale.go`:
   ```go
   type TimescaleStore struct {
       pool *pgxpool.Pool
   }

   func NewTimescaleStore(ctx context.Context, connStr string) (*TimescaleStore, error) {
       config, err := pgxpool.ParseConfig(connStr)
       if err != nil {
           return nil, err
       }
       config.MaxConns = 20
       pool, err := pgxpool.NewWithConfig(ctx, config)
       if err != nil {
           return nil, err
       }
       return &TimescaleStore{pool: pool}, nil
   }
   ```

5. **Implement `Write`** using a parameterized query:
   ```go
   func (s *TimescaleStore) Write(span *proto.Span) error {
       _, err := s.pool.Exec(ctx,
           `INSERT INTO spans (trace_id, span_id, service_name, operation_name,
            start_time, duration_nanos, model_name, batch_size)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
           string(span.TraceId),
           string(span.SpanId),
           span.ServiceName,
           span.OperationName,
           time.Unix(0, span.StartTimeUnixNano),
           span.DurationNanos,
           inferenceModelName(span),
           inferenceBatchSize(span),
       )
       return err
   }
   ```
   Notice `$1, $2, ...` — these are parameterized placeholders. Never concatenate user input into SQL strings.

6. **Implement batch writes** using `pgx`'s `CopyFrom` — this is 10-50x faster than individual inserts:
   ```go
   func (s *TimescaleStore) WriteBatch(spans []*proto.Span) error {
       rows := make([][]any, len(spans))
       for i, span := range spans {
           rows[i] = []any{
               string(span.TraceId), string(span.SpanId),
               span.ServiceName, span.OperationName,
               time.Unix(0, span.StartTimeUnixNano), span.DurationNanos,
           }
       }
       _, err := s.pool.CopyFrom(
           ctx,
           pgx.Identifier{"spans"},
           []string{"trace_id", "span_id", "service_name", "operation_name", "start_time", "duration_nanos"},
           pgx.CopyFromRows(rows),
       )
       return err
   }
   ```

7. **Implement `GetTrace`** — fetch all spans for a trace ID, ordered by start time:
   ```go
   rows, err := s.pool.Query(ctx,
       `SELECT trace_id, span_id, service_name, operation_name,
               start_time, duration_nanos, model_name
        FROM spans WHERE trace_id = $1 ORDER BY start_time ASC`,
       traceID,
   )
   ```
   Scan rows into `[]*proto.Span` and return them.

8. **Implement `ListSpans`** using the `SpanFilter` struct from Phase 4:
   - Build a query dynamically using `WHERE` clauses based on which filter fields are set
   - Use `start_time >= $1 AND start_time <= $2` for time range filtering
   - This query will be fast because TimescaleDB only scans the relevant time chunks

9. **Verify `TimescaleStore` satisfies `StorageWriter`** — add this compile-time check to the file:
   ```go
   var _ StorageWriter = (*TimescaleStore)(nil)
   ```
   If `TimescaleStore` is missing any method from the interface, this line causes a compile error. This is a common Go pattern.

10. **Swap the storage backend** in `cmd/collector/main.go` — replace `NewMemoryStore()` with `NewTimescaleStore(...)`. Send some spans with your test client, then query the DB directly to confirm they were written:
    ```bash
    docker exec -it infertrace-db psql -U postgres -d infertrace -c "SELECT * FROM spans;"
    ```
