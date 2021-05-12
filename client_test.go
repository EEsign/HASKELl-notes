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
	cancel, errC, err := c.SubscribePrice(pairs, ch)
	defer cancel()
	assert.NoError(t, err)
	count := 0
	notify := make(chan struct{})
	go func() {
		for price := range ch {
			found := false
			t.Logf("%#v", *price)
			for _, pair := range pairs {
				if price.Pair == pair {
					found = true
					count++
					break
				}
			}
			assert.True(t, found)
			if count >= 10 {
				close(notify)
				break
			}
		}
	}()

	select {
	case <-notify:
	case <-time.After(20 * time.Seco