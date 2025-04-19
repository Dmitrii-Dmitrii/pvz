package drivers

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"
	"time"
)

func setupPostgresContainer(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()

	containerCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	pgPort := "5432/tcp"
	dbName := "testdb"
	dbUser := "postgres"
	dbPassword := "postgres"

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{pgPort},
		Env: map[string]string{
			"POSTGRES_DB":       dbName,
			"POSTGRES_USER":     dbUser,
			"POSTGRES_PASSWORD": dbPassword,
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithStartupTimeout(2 * time.Minute),
	}

	postgresContainer, err := testcontainers.GenericContainer(containerCtx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start postgres container: %v", err)
	}

	hostIP, err := postgresContainer.Host(ctx)
	require.NoError(t, err)

	mappedPort, err := postgresContainer.MappedPort(ctx, nat.Port(pgPort))
	require.NoError(t, err)

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		dbUser, dbPassword, hostIP, mappedPort.Port(), dbName)

	time.Sleep(2 * time.Second)

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		postgresContainer.Terminate(ctx)
		t.Fatalf("Could not connect to database: %v", err)
	}

	err = setupSchema(ctx, pool)
	if err != nil {
		pool.Close()
		postgresContainer.Terminate(ctx)
		t.Fatalf("Could not set up schema: %v", err)
	}

	return pool, func() {
		pool.Close()
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %s", err)
		}
	}
}

func setupSchema(ctx context.Context, pool *pgxpool.Pool) error {
	schema := createSchema

	_, err := pool.Exec(ctx, schema)
	return err
}

func createTestData(ctx context.Context, pool *pgxpool.Pool) ([]pgtype.UUID, []pgtype.UUID, []pgtype.UUID, error) {
	pvzIds := make([]pgtype.UUID, 2)
	for i := 0; i < 2; i++ {
		idBytes := uuid.New()
		id := pgtype.UUID{Bytes: idBytes, Valid: true}
		pvzIds[i] = id

		city := "Москва"
		if i == 1 {
			city = "Санкт-Петербург"
		}

		_, err := pool.Exec(ctx, queryCreatePvz, id, time.Now().Add(-time.Hour*24*time.Duration(i+1)), city)

		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to create pvz: %w", err)
		}
	}

	receptionIds := make([]pgtype.UUID, 3)
	for i := 0; i < 3; i++ {
		idBytes := uuid.New()
		id := pgtype.UUID{Bytes: idBytes, Valid: true}
		receptionIds[i] = id

		pvzIndex := i % 2
		status := "in_progress"
		if i == 2 {
			status = "close"
		}

		_, err := pool.Exec(ctx, queryCreateReception, id, time.Now().Add(-time.Hour*12*time.Duration(i+1)), pvzIds[pvzIndex], status)

		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to create reception: %w", err)
		}
	}

	for i := 0; i < 5; i++ {
		idBytes := uuid.New()
		id := pgtype.UUID{Bytes: idBytes, Valid: true}

		receptionIndex := i % 3
		productType := "одежда"
		if i%2 == 1 {
			productType = "электроника"
		}

		_, err := pool.Exec(ctx, queryCreateProduct, id, time.Now().Add(-time.Hour*12*time.Duration(i+1)), productType, receptionIds[receptionIndex])

		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to create product: %w", err)
		}
	}

	userIds := make([]pgtype.UUID, 2)
	for i := 0; i < 2; i++ {
		idBytes := uuid.New()
		id := pgtype.UUID{Bytes: idBytes, Valid: true}
		userIds[i] = id

		role := "employee"
		if i == 1 {
			role = "moderator"
		}

		email := "dummy." + role + "@example.com"

		_, err := pool.Exec(ctx, queryCreateUser, id, email, "", role)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	return pvzIds, receptionIds, userIds, nil
}
