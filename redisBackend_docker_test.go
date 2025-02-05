//go:build integration

package httpcache_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"flamingo.me/flamingo/v3/framework/flamingo"

	"flamingo.me/httpcache"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	redisContainer testcontainers.Container
	redisHost      string
	redisPort      string
)

// TestMain set
func TestMain(m *testing.M) {
	setup()

	code := m.Run()

	teardown() // comment out, if you want to keep the docker-instance running for debugging
	os.Exit(code)
}

// setup the redis docker-container for integration tests
func setup() {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "valkey/valkey:7",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForLog("Ready to accept connections"),
			wait.ForListeningPort("6379/tcp")),
	}

	var err error
	redisContainer, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		log.Fatal(err)
	}

	port, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		log.Fatal(err)
	}

	redisPort = port.Port()

	redisHost, err = redisContainer.Host(ctx)
	if err != nil {
		log.Fatal(err)
	}

	address := fmt.Sprintf("%s:%s", redisHost, port.Port())

	conn, err := redis.Dial("tcp", address)
	if err != nil {
		log.Fatal(err)
	}

	_, err = conn.Do("PING")
	if err != nil {
		log.Fatal(err)
	}

	_ = conn.Close()
}

// teardown the redis docker-container
func teardown() {
	err := redisContainer.Terminate(context.Background())
	if err != nil {
		log.Fatalf("Error purging docker resources: %s", err)
	}
}

func Test_RunDefaultBackendTestCase_RedisBackend(t *testing.T) {
	t.Parallel()

	config := httpcache.RedisBackendConfig{
		MaxIdle:            8,
		IdleTimeOutSeconds: 30,
		Host:               redisHost,
		Port:               redisPort,
	}
	factory := httpcache.RedisBackendFactory{}
	backend, err := factory.Inject(flamingo.NullLogger{}).SetConfig(config).SetFrontendName("testfrontend").Build()
	assert.NoError(t, err)
	testcase := NewBackendTestCase(t, backend, false)
	testcase.RunTests()
}
