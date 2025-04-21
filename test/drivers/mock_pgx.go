package drivers

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/mock"
)

type MockAdapter struct {
	mock.Mock
}

func (m *MockAdapter) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	args := m.Called(ctx, sql, arguments)
	return args.Get(0).(pgconn.CommandTag), args.Error(1)
}

func (m *MockAdapter) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	allArgs := []interface{}{ctx, sql}
	if len(args) > 0 {
		allArgs = append(allArgs, args)
	}

	arguments := m.Called(allArgs...)
	return arguments.Get(0).(pgx.Rows), arguments.Error(1)
}

func (m *MockAdapter) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	arguments := m.Called(ctx, sql, args)
	return arguments.Get(0).(pgx.Row)
}

func (m *MockAdapter) Begin(ctx context.Context) (pgx.Tx, error) {
	args := m.Called(ctx)
	return args.Get(0).(pgx.Tx), args.Error(1)
}

type MockRows struct {
	mock.Mock
}

func (m *MockRows) Next() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockRows) Scan(dest ...any) error {
	args := m.Called(dest...)
	return args.Error(0)
}

func (m *MockRows) Close() {
	m.Called()
}

func (m *MockRows) Err() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRows) CommandTag() pgconn.CommandTag {
	args := m.Called()
	return args.Get(0).(pgconn.CommandTag)
}

func (m *MockRows) FieldDescriptions() []pgconn.FieldDescription {
	args := m.Called()
	return args.Get(0).([]pgconn.FieldDescription)
}

func (m *MockRows) Values() ([]any, error) {
	args := m.Called()
	return args.Get(0).([]any), args.Error(1)
}

func (m *MockRows) RawValues() [][]byte {
	args := m.Called()
	return args.Get(0).([][]byte)
}

func (m *MockRows) Conn() *pgx.Conn {
	args := m.Called()
	return args.Get(0).(*pgx.Conn)
}

type MockRow struct {
	mock.Mock
}

func (m *MockRow) Scan(dest ...any) error {
	args := m.Called(dest...)
	return args.Error(0)
}

type MockTx struct {
	mock.Mock
}

func (m *MockTx) Begin(ctx context.Context) (pgx.Tx, error) {
	args := m.Called(ctx)
	return args.Get(0).(pgx.Tx), args.Error(1)
}

func (m *MockTx) Commit(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTx) Rollback(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	args := m.Called(ctx, tableName, columnNames, rowSrc)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	args := m.Called(ctx, b)
	return args.Get(0).(pgx.BatchResults)
}

func (m *MockTx) LargeObjects() pgx.LargeObjects {
	args := m.Called()
	return args.Get(0).(pgx.LargeObjects)
}

func (m *MockTx) Prepare(ctx context.Context, name string, sql string) (*pgconn.StatementDescription, error) {
	args := m.Called(ctx, name, sql)
	return args.Get(0).(*pgconn.StatementDescription), args.Error(1)
}

func (m *MockTx) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	args := m.Called(ctx, sql, arguments)
	return args.Get(0).(pgconn.CommandTag), args.Error(1)
}

func (m *MockTx) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	callArgs := m.Called(ctx, sql, args)
	return callArgs.Get(0).(pgx.Rows), callArgs.Error(1)
}

func (m *MockTx) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	callArgs := m.Called(ctx, sql, args)
	return callArgs.Get(0).(pgx.Row)
}

func (m *MockTx) Conn() *pgx.Conn {
	args := m.Called()
	return args.Get(0).(*pgx.Conn)
}
