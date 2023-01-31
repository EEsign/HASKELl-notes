
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
