//go:build integration

package httpcache_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	redisHost string
	redisPort string
	username  = "UserName*"
	password  = "VFFFDsd1&"
)

// TestMain set
func TestMain(m *testing.M) {
	container := setup(username, password)

	code := m.Run()

	teardown(container) // comment out, if you want to keep the docker-instance running for debugging
	os.Exit(code)
}

// setup the redis docker-container for integration tests
func setup(username, password string) testcontainers.Container {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "valkey/valkey:8",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForLog("Ready to accept connections"),
			wait.ForListeningPort("6379/tcp")),
		Cmd: []string{
			"valkey-server",
			"--user default off",
			fmt.Sprintf("--user %s on >%s allcommands allkeys", username, password),
		},
	}

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
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

	options := []redis.DialOption{
		redis.DialUsername(username),
		redis.DialPassword(password),
	}

	conn, err := redis.Dial("tcp", address, options...)
	if err != nil {
		log.Fatal(err)
	}

	_, err = conn.Do("PING")
	if err != nil {
		log.Fatal(err)
	}

	_ = conn.Close()

	return redisContainer
}

// teardown the redis docker-container
func teardown(redisContainer testcontainers.Container) {
	err := redisContainer.Terminate(context.Background())
	if err != nil {
		log.Fatalf("Error purging docker resources: %s", err)
	}
}
