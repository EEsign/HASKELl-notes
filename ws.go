
package rapid

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
)

const (
	pingPeriod           = 45 * time.Second
	pongWait             = 50 * time.Second
	writeWait            = 20 * time.Second
	maxMessageSize       = 409600
	clientSendChanBuffer = 20

	ChannelPrice Channel = "price" // 订阅交易对最新价格
	ChannelOrder Channel = "order" // 订单

	OpSubscribe   Operation = "subscribe"   // 订阅
	OpUnsubscribe Operation = "unsubscribe" // 取消订阅
	OpOrder       Operation = "order"       // 下单

	MsgTypeSubscribed   MsgType = "subscribed"
	MsgTypeUnsubscribed MsgType = "unsubscribed"
	MsgTypeOrder        MsgType = "ordered"

	MsgTypeUpdate MsgType = "update" // 数据更新
)

type Operation string
type Channel string
type MsgType string

type ReqMessage struct {
	Id      uint64      `json:"id"`      // 请求ID
	Op      Operation   `json:"op"`      // 操作
	Channel Channel     `json:"channel"` // 频道
	Args    interface{} `json:"args"`    // 参数
}

type RespMessage struct {
	Id      uint64          `json:"id,omitempty"`
	Type    MsgType         `json:"type"`    // 消息类型
	Channel Channel         `json:"channel"` // 频道
	Data    json.RawMessage `json:"data"`    // 数据
}

type ConfirmMessage struct {
	RespMessage
	Data ConfirmData `json:"data"`
}

func (confirm ConfirmMessage) IsSuccess() bool {
	return confirm.Data.Code == 0
}

type PairArgs []string

type OrderArgs struct {
	Pair              string          `json:"pair" validate:"required"`
	Type              string          `json:"type" validate:"required"`
	TokenSymbolIn     string          `json:"tokenSymbolIn" validate:"required"`
	AmountIn          decimal.Decimal `json:"amountIn" validate:"required"`
	AmountOutMin      decimal.Decimal `json:"amountOutMin" validate:"required"`
	GasPriceMax       decimal.Decimal `json:"gasPriceMax" validate:"required"`
	TargetBlockNumber uint64          `json:"targetBlockNumber" validate:"required"`
}

type ConfirmData struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type PriceData struct {
	Timestamp   int64  `json:"ts"` // 毫秒时间戳
	BlockNumber uint64 `json:"n"`  // 区块号
	BlockTime   uint64 `json:"bt"` // 区块头时间戳
	Pair        string `json:"p"`  // 交易对名称
	R0          string `json:"r0"`
	R1          string `json:"r1"`
}

type OrderResultData struct {
	Id            uint64          `json:"id"` // 任务id
	Pair          string          `json:"pair"`
	TokenSymbolIn string          `json:"tokenSymbolIn"`
	Success       bool            `json:"success"`
	BlockNumber   uint64          `json:"blockNumber"`
	AmountIn      decimal.Decimal `json:"amountIn"`
	AmountOut     decimal.Decimal `json:"amountOut"`
	GasFee        decimal.Decimal `json:"gasFee"`
	Hash          string          `json:"hash"`
}

type WsClient struct {
	count          uint64
	conn           *websocket.Conn
	send           chan *ReqMessage
	Closed         chan struct{}
	closedOnce     sync.Once
	messageHandler MessageHandler
	callbacks      map[uint64]Callback
	callbackMutex  sync.Mutex
	Logger         Logger
}

func NewWsClient(conn *websocket.Conn, logger Logger) *WsClient {
	client := &WsClient{
		conn:           conn,
		send:           make(chan *ReqMessage, clientSendChanBuffer),
		Closed:         make(chan struct{}),
		messageHandler: nil,
		callbacks:      make(map[uint64]Callback),
		Logger:         logger,
	}

	go client.writePump()
	go client.readPump()
	return client
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *WsClient) readPump() {
	defer func() {
		c.Close()
		c.Logger.Infof("stop readPump of client %v", c.conn.RemoteAddr())
	}()
	c.conn.SetReadLimit(maxMessageSize)
	err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		c.Logger.Errorf(err.Error())
	}
	c.conn.SetPongHandler(func(string) error {
		c.Logger.Infof("received pong message from peer")
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Logger.Infof("websocket read message error: %v", err)
			}
			break
		}