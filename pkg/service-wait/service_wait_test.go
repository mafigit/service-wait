package service_wait

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/urfave/cli/v3"
	"testing"
)

func TestEmptyFlags(t *testing.T) {
	cmd := &cli.Command{
		Flags: GetFlags(),
	}
	_, err := New(cmd)
	require.ErrorContains(t, err, "at least one of --url, --psql-host, --mongo-host, --urls must be provided")
}

func TestProbePSQLFlags(t *testing.T) {
	cmd := &cli.Command{
		Flags: GetFlags(),
	}
	// Only set one required flag for a postgres probe, which should result in an error
	require.NoError(t, cmd.Set("psql-host", "test"))
	_, err := New(cmd)
	require.ErrorContains(t, err, "--psql-user is required")
}

func TestURLFlags(t *testing.T) {
	cmd := &cli.Command{
		Flags: GetFlags(),
	}
	require.NoError(t, cmd.Set("urls", "http://test:8080,http://test2:8080"))
	_, err := New(cmd)
	require.NoError(t, err)

	require.NoError(t, cmd.Set("url", "http://test:8080"))
	_, err = New(cmd)
	require.NoError(t, err)

	require.NoError(t, cmd.Set("url", "://test:8080"))
	_, err = New(cmd)
	require.ErrorContains(t, err, "missing protocol scheme")
}

func TestProbeHttpEndpoint(t *testing.T) {
	// Currently we only test successful cases, since you cannot know a testcontainers host and port in advance.
	// Working around this constraint requires a lot of logic which is currently exceeding the scope of this tool.
	ctx := context.Background()
	httpHttpsEchoContainer1, err := testcontainers.Run(
		ctx, "mendhak/http-https-echo",
		testcontainers.WithExposedPorts("8080/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("8080/tcp"),
		),
	)
	httpHttpsEchoContainer2, err := testcontainers.Run(
		ctx, "mendhak/http-https-echo",
		testcontainers.WithExposedPorts("8080/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("8080/tcp"),
		),
	)
	port1, err := httpHttpsEchoContainer1.MappedPort(ctx, "8080/tcp")
	if err != nil {
		t.Error(err)
	}
	host1, err := httpHttpsEchoContainer1.Host(ctx)
	if err != nil {
		t.Error(err)
	}

	port2, err := httpHttpsEchoContainer2.MappedPort(ctx, "8080/tcp")
	if err != nil {
		t.Error(err)
	}
	host2, err := httpHttpsEchoContainer2.Host(ctx)
	if err != nil {
		t.Error(err)
	}

	ctxProbeHttpEndpoint, cancel := context.WithCancel(context.Background())
	defer cancel()

	url1 := fmt.Sprintf("http://%s:%d/", host1, port1.Int())
	config := &Config{
		URL:      url1,
		Timeout:  "1s",
		Interval: "5s",
	}
	err = ProbeHttpEndpoint(ctxProbeHttpEndpoint, config)
	require.NoError(t, err)

	url2 := fmt.Sprintf("http://%s:%d/", host2, port2.Int())
	// test multiple urls
	config2 := &Config{
		URLS:     []string{url1, url2},
		Timeout:  "1s",
		Interval: "5s",
	}
	err = ProbeHttpEndpoint(ctxProbeHttpEndpoint, config2)
	require.NoError(t, err)
}

func TestProbePSQLEndpoint(t *testing.T) {
	// Currently we only test successful cases, since you cannot know a testcontainers host and port in advance.
	// Working around this constraint requires a lot of logic which is currently exceeding the scope of this tool.
	const user = "test"
	const password = "test"
	const database = "test"
	ctx := context.Background()
	psqlContainer, err := testcontainers.Run(
		ctx, "postgres:17",
		testcontainers.WithExposedPorts("5432/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp"),
		),
		testcontainers.WithEnv(map[string]string{
			"POSTGRES_USER":     user,
			"POSTGRES_PASSWORD": password,
			"POSTGRES_DB":       database,
		}),
	)

	port, err := psqlContainer.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Error(err)
	}
	host, err := psqlContainer.Host(ctx)
	if err != nil {
		t.Error(err)
	}

	ctxProbePSQLEndpoint, cancel := context.WithCancel(context.Background())
	defer cancel()
	config := &Config{
		PsqlHost:     host,
		PsqlPort:     port.Int(),
		PsqlUser:     user,
		PsqlPassword: password,
		PsqlDatabase: database,
		Timeout:      "1s",
		Interval:     "5s",
	}
	err = ProbePSQLEndpoint(ctxProbePSQLEndpoint, config)
	require.NoError(t, err)

	configWithDsn := &Config{
		PsqlDsn:  fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable", user, password, host, port.Int(), database),
		Timeout:  "1s",
		Interval: "5s",
	}
	fmt.Println(configWithDsn.PsqlDsn)
	err = ProbePSQLEndpoint(ctxProbePSQLEndpoint, configWithDsn)
	require.NoError(t, err)
}

func TestProbeMongoDBEndpoint(t *testing.T) {
	// Currently we only test successful cases, since you cannot know a testcontainers host and port in advance.
	// Working around this constraint requires a lot of logic which is currently exceeding the scope of this tool.
	ctx := context.Background()
	mongoContainer, err := testcontainers.Run(
		ctx, "mongo:6.0",
		testcontainers.WithExposedPorts("27017/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("27017/tcp"),
		),
	)
	port, err := mongoContainer.MappedPort(ctx, "27017/tcp")
	if err != nil {
		t.Error(err)
	}
	host, err := mongoContainer.Host(ctx)
	if err != nil {
		t.Error(err)
	}

	ctxProbeMongoEndpoint, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := &Config{
		MongoHost: host,
		MongoPort: port.Int(),
		Timeout:   "1s",
		Interval:  "5s",
	}
	err = ProbeMongoEndpoint(ctxProbeMongoEndpoint, config)
	require.NoError(t, err)
}

func TestMongoUriBuilder(t *testing.T) {
	tests := []struct {
		name       string
		host       string
		port       int
		user       string
		password   string
		database   string
		authSource string
		want       string
	}{
		{
			name: "default port, no auth, no db, no authSource",
			host: "localhost",
			port: 0,
			want: "mongodb://localhost:27017",
		},
		{
			name: "custom port",
			host: "localhost",
			port: 12345,
			want: "mongodb://localhost:12345",
		},
		{
			name:     "with user and password",
			host:     "host",
			port:     27017,
			user:     "user",
			password: "pass",
			want:     "mongodb://user:pass@host:27017",
		},
		{
			name:     "with database",
			host:     "host",
			port:     27017,
			database: "db",
			want:     "mongodb://host:27017/db",
		},
		{
			name:       "with authSource",
			host:       "host",
			port:       27017,
			authSource: "admin",
			want:       "mongodb://host:27017?authSource=admin",
		},
		{
			name:       "with all fields",
			host:       "host",
			port:       27017,
			user:       "user",
			password:   "pass",
			database:   "db",
			authSource: "admin",
			want:       "mongodb://user:pass@host:27017/db?authSource=admin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mongoUriBuilder(tt.host, tt.port, tt.user, tt.password, tt.database, tt.authSource)
			require.Equal(t, tt.want, got)
		})
	}
}
