package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestContainer(t *testing.T) (string, func()) {
	ctx := context.Background()

	// Use pre-built image instead of building from Dockerfile
	imageName := "bitcoin-ltp-api:test"

	req := testcontainers.ContainerRequest{
		Image:        imageName,
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForHTTP("/health").WithPort("8080/tcp").WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Get the container host and port
	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "8080")
	require.NoError(t, err)

	baseURL := fmt.Sprintf("http://%s:%s", host, port.Port())

	// Cleanup function
	cleanup := func() {
		err := container.Terminate(ctx)
		assert.NoError(t, err)
	}

	return baseURL, cleanup
}

func TestGetLTP_AllPairs_Container(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping container test in short mode")
	}

	baseURL, cleanup := setupTestContainer(t)
	defer cleanup()

	// Wait a bit for the server to be fully ready
	time.Sleep(2 * time.Second)

	resp, err := http.Get(fmt.Sprintf("%s/api/v1/ltp", baseURL))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response struct {
		LTP []struct {
			Pair   string  `json:"pair"`
			Amount float64 `json:"amount"`
		} `json:"ltp"`
	}

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(response.LTP), 1)

	// Verify response structure
	for _, item := range response.LTP {
		assert.NotEmpty(t, item.Pair)
		assert.Greater(t, item.Amount, 0.0)
	}
}

func TestGetLTP_SinglePair_Container(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping container test in short mode")
	}

	baseURL, cleanup := setupTestContainer(t)
	defer cleanup()

	time.Sleep(2 * time.Second)

	resp, err := http.Get(fmt.Sprintf("%s/api/v1/ltp?pairs=BTC/USD", baseURL))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response struct {
		LTP []struct {
			Pair   string  `json:"pair"`
			Amount float64 `json:"amount"`
		} `json:"ltp"`
	}

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	assert.Len(t, response.LTP, 1)
	assert.Equal(t, "BTC/USD", response.LTP[0].Pair)
	assert.Greater(t, response.LTP[0].Amount, 0.0)
}

func TestGetLTP_MultiplePairs_Container(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping container test in short mode")
	}

	baseURL, cleanup := setupTestContainer(t)
	defer cleanup()

	time.Sleep(2 * time.Second)

	resp, err := http.Get(fmt.Sprintf("%s/api/v1/ltp?pairs=BTC/USD,BTC/EUR", baseURL))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response struct {
		LTP []struct {
			Pair   string  `json:"pair"`
			Amount float64 `json:"amount"`
		} `json:"ltp"`
	}

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	assert.Len(t, response.LTP, 2)

	// Verify both pairs are present
	pairs := make(map[string]bool)
	for _, item := range response.LTP {
		pairs[item.Pair] = true
	}
	assert.True(t, pairs["BTC/USD"])
	assert.True(t, pairs["BTC/EUR"])
}

func TestGetLTP_InvalidPair_Container(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping container test in short mode")
	}

	baseURL, cleanup := setupTestContainer(t)
	defer cleanup()

	time.Sleep(2 * time.Second)

	resp, err := http.Get(fmt.Sprintf("%s/api/v1/ltp?pairs=BTC/INVALID", baseURL))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response struct {
		Error string `json:"error"`
	}

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "invalid pair")
}

func TestHealthEndpoint_Container(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping container test in short mode")
	}

	baseURL, cleanup := setupTestContainer(t)
	defer cleanup()

	time.Sleep(2 * time.Second)

	resp, err := http.Get(fmt.Sprintf("%s/health", baseURL))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]string
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

func TestGetLTP_Cache_Container(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping container test in short mode")
	}

	baseURL, cleanup := setupTestContainer(t)
	defer cleanup()

	time.Sleep(2 * time.Second)

	// First request
	resp1, err := http.Get(fmt.Sprintf("%s/api/v1/ltp?pairs=BTC/USD", baseURL))
	require.NoError(t, err)
	defer resp1.Body.Close()
	assert.Equal(t, http.StatusOK, resp1.StatusCode)

	body1, err := io.ReadAll(resp1.Body)
	require.NoError(t, err)

	var response1 struct {
		LTP []struct {
			Pair   string  `json:"pair"`
			Amount float64 `json:"amount"`
		} `json:"ltp"`
	}
	err = json.Unmarshal(body1, &response1)
	require.NoError(t, err)
	firstAmount := response1.LTP[0].Amount

	// Second request - should use cache (within 1 minute)
	time.Sleep(1 * time.Second)
	resp2, err := http.Get(fmt.Sprintf("%s/api/v1/ltp?pairs=BTC/USD", baseURL))
	require.NoError(t, err)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	body2, err := io.ReadAll(resp2.Body)
	require.NoError(t, err)

	var response2 struct {
		LTP []struct {
			Pair   string  `json:"pair"`
			Amount float64 `json:"amount"`
		} `json:"ltp"`
	}
	err = json.Unmarshal(body2, &response2)
	require.NoError(t, err)
	secondAmount := response2.LTP[0].Amount

	// Amounts should be the same (cached)
	assert.Equal(t, firstAmount, secondAmount)
}
