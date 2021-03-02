package rapid

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	ctx       = context.TODO()
	apiKey    = "apiKey"
	apiSecret = "apiSecret"
	pairs     = []string{"WBNB-BUSD@PANCAKESWAP", "WBNB-USDT@PANCAKESWAP"}
	c         = NewClient(apiKey, apiSecret)
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
}

func teardown() {
}

func TestNewClient(t *testing.T) {
	ws, err := c.NewStream()
	assert.NoError(t, err)
	assert.False(t, ws.IsClosed())
}

func TestClient_SubscribePrice(t *testing.T) {
	ch := make(chan *PriceData, 20)
	cancel, errC, err := c.SubscribePrice(pairs