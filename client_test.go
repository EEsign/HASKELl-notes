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
	case <-time.After(20 * time.Second):
		assert.Errorf(t, assert.AnError, "timeout")
	case err := <-errC:
		t.Log(err)
		panic(err)
	}
	cancel()
	time.Sleep(2 * time.Second)
	select {
	case err = <-errC:
	default:
	}
	assert.ErrorIs(t, err, ErrStreamClosed)
	t.Log(err)
}

func TestClient_SubscribeOrderResult(t *testing.T) {
	ch := make(chan *OrderResultData, 20)
	cancel, _, err := c.SubscribeOrderResult(ch)
	defer cancel()
	assert.NoError(t, err)
}

func TestClient_CreateOrder(t *testing.T) {
	req := CreateOrderReq{
		Pair:              "UNI-BUSD@MDEX",
		Type:              "pga",
		TokenSymbolIn:     "BUSD",
		AmountIn:          "2975000000000000000",
		AmountOutMin:      "0.001",
		GasPriceMax:       "90000000000",
		TraceTime:         strconv.FormatInt(time.Now().UnixNano(), 10),
		TargetBlockNumber: 22534222,
	}
	resp, err := c.CreateOrder(ctx, req)
	assert.NoError(t, err)
	if assert.NotNil(t, resp) {
		assert.Greater(t, resp.Id, uint64(0))
		t.Logf("resp.Id %v", resp.Id)
	}
}

func TestClient_GetPairs(t *testing.T) {
	req := GetPairsReq{
		Name:     "WBNB-BUSD@PANCAKESWAP",
		Exchange: ExchangePancakeSwap,
	}
	pairs, err := c.GetPairs(ctx, req)
	assert.NoError(t, err)
	assert.Len(t, pairs, 1)
	t.Log(pairs[0])
}

func TestClient_CreateOrderByStream(t *testing.T) {
	req := CreateOrderReq{
		Pair:              "WBNB-BUSD@PANCAKESWAP",
		Type:              "pga",
		TokenSymbolIn:     "WBNB",
		AmountIn:          "0.1",
		AmountOutMin:      "10",
		GasPriceMax:       "90000000000",
		TargetBlockNumber: 21675044,
	}
	resp, err := c.CreateOrderByStream(req)
	assert.NoError(t, err)
	if assert.NotNil(t, resp) {
		assert.Greater(t, resp.Id, uint64(0))
		t.Logf("resp.Id %v", resp.Id)
	}
}

func TestClient_GetOrderResult(t *testing.T) {
	result, err := c.GetOrderResult(ct