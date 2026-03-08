package postgres

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/akaporn-katip/go-project-structure-template/internal/application/repositories"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestWithTx(t *testing.T) {
	// 1. Setup Mock DB
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "postgres")

	// 2. Setup Mock Metrics (using SDK Reader to verify counts)
	reader := metric.NewManualReader()
	provider := metric.NewMeterProvider(metric.WithReader(reader))
	testMeter := provider.Meter("test")

	// 3. Initialize UOW
	uow, err := NewUnitOfWork(sqlxDB, testMeter)
	require.NoError(t, err)

	t.Run("Successful Commit", func(t *testing.T) {
		ctx := context.Background()

		// Expectations
		mock.ExpectBegin()
		mock.ExpectCommit()

		res, err := WithTx(ctx, func(ctx context.Context, repos repositories.Repositories) (*string, error) {
			s := "success"
			return &s, nil
		}, uow)

		assert.NoError(t, err)
		assert.Equal(t, "success", *res)
		assert.NoError(t, mock.ExpectationsWereMet())

		// Verify Metrics: activeTransactions should be 0 (1 - 1)
		var rm metricdata.ResourceMetrics
		_ = reader.Collect(ctx, &rm)
		// You can drill down into rm to verify db.transaction.active is 0
	})

	t.Run("Rollback on Error", func(t *testing.T) {
		ctx := context.Background()

		mock.ExpectBegin()
		mock.ExpectRollback()

		_, err := WithTx(ctx, func(ctx context.Context, repos repositories.Repositories) (*string, error) {
			return nil, errors.New("business logic failed")
		}, uow)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "business logic failed")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Rollback on Panic", func(t *testing.T) {
		ctx := context.Background()

		mock.ExpectBegin()
		mock.ExpectRollback()

		assert.Panics(t, func() {
			_, _ = WithTx(ctx, func(ctx context.Context, repos repositories.Repositories) (*string, error) {
				panic("something went horribly wrong")
			}, uow)
		})

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Transaction Begin Failure", func(t *testing.T) {
		ctx := context.Background()

		mock.ExpectBegin().WillReturnError(errors.New("connection limit reached"))

		_, err := WithTx(ctx, func(ctx context.Context, repos repositories.Repositories) (*string, error) {
			return nil, nil
		}, uow)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to begin transaction")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestWithTx_Concurrency(t *testing.T) {
	mockDB, mock, _ := sqlmock.New()
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "postgres")

	provider := metric.NewMeterProvider()
	uow, _ := NewUnitOfWork(sqlxDB, provider.Meter("test"))

	// 1. IMPORTANT: Allow expectations to be met in any order
	mock.MatchExpectationsInOrder(false)

	const concurrentCount = 10
	var wg sync.WaitGroup
	wg.Add(concurrentCount)

	// 2. Set up expectations for 10 separate transactions
	for i := 0; i < concurrentCount; i++ {
		mock.ExpectBegin()
		mock.ExpectCommit()
	}

	// 3. Run goroutines
	for i := 0; i < concurrentCount; i++ {
		go func(id int) {
			defer wg.Done()

			// Each WithTx now creates its own Transaction struct internally
			_, err := WithTx(context.Background(), func(ctx context.Context, repos repositories.Repositories) (*int, error) {
				// Simulate a tiny bit of processing time
				time.Sleep(time.Millisecond * 10)
				return &id, nil
			}, uow)

			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// 4. Verify all 10 Begins and 10 Commits happened
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithTx_RollbackOnError(t *testing.T) {
	mockDB, mock, _ := sqlmock.New()
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "postgres")
	uow, _ := NewUnitOfWork(sqlxDB, metric.NewMeterProvider().Meter("test"))

	t.Run("Should Rollback when inner function fails", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectRollback()

		_, err := WithTx(context.Background(), func(ctx context.Context, repos repositories.Repositories) (*string, error) {
			return nil, errors.New("database constraint violation")
		}, uow)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database constraint violation")
	})
}

func TestWithTx_PanicRecovery(t *testing.T) {
	mockDB, mock, _ := sqlmock.New()
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "postgres")
	uow, _ := NewUnitOfWork(sqlxDB, metric.NewMeterProvider().Meter("test"))

	t.Run("Should Rollback on Panic", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectRollback()

		assert.Panics(t, func() {
			_, _ = WithTx(context.Background(), func(ctx context.Context, repos repositories.Repositories) (*string, error) {
				panic("unexpected crash")
			}, uow)
		})

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
