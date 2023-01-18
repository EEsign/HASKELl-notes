
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
