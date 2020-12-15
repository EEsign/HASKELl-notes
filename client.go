
package rapid

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	"github.com/umbracle/ethgo"
)

const (
	baseHTTPURL = "https://rapidtrading-api.liquiditytech.com"
	baseWsURL   = "wss://rapidtrading-api.liquiditytech.com"
	httpTimeout = 15 * time.Second
)

var (
	ErrStreamClosed = errors.New("ws client was closed")
)

type Logger interface {
	Infof(msg string, args ...interface{})
	Errorf(msg string, args ...interface{})
}

type CommonResp struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type Client struct {
	APIKey      string
	SecretKey   string
	BaseHTTPURL string
	BaseWsURL   string
	HTTPClient  *http.Client
	Debug       bool
	Logger      Logger

	simplexClient *WsClient // 不涉及推送流时使用
	simplexMutex  sync.Mutex
}

func NewClient(apiKey, secretKey string) *Client {
	return &Client{
		APIKey:      apiKey,
		SecretKey:   secretKey,
		BaseHTTPURL: baseHTTPURL,
		BaseWsURL:   baseWsURL,
		HTTPClient: &http.Client{
			Timeout: httpTimeout,
		},
		Logger: logImp{log.New(os.Stderr, "[FlashNet] ", log.LstdFlags|log.Lmicroseconds)},
	}
}

type logImp struct {
	stdLog *log.Logger
}

func (l logImp) Infof(msg string, data ...interface{}) {
	l.stdLog.Printf(msg, data...)
}

func (l logImp) Errorf(msg string, data ...interface{}) {
	l.stdLog.Printf(msg, data...)
}

func (c *Client) signParams(params url.Values) (sign string) {
	if params == nil {
		params = make(url.Values)
	}
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	sb := strings.Builder{}
	for _, k := range keys {
		value := params.Get(k)
		if value == "" {
			continue
		}
		sb.WriteString(k + "=" + value + "&")
	}
	payload := strings.TrimSuffix(sb.String(), "&")
	h := hmac.New(sha256.New, []byte(c.SecretKey))
	h.Write([]byte(payload))
	sign = hex.EncodeToString(h.Sum(nil))
	params.Del("apiKeyParamName")
	return sign
}

func getTimestamp() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05")
}

func (c *Client) callAPI(ctx context.Context, method string, path string, query url.Values, body interface{}, target interface{}) (statusCode int, respBody []byte, err error) {
	u, err := url.Parse(fmt.Sprintf("%v%v", c.BaseHTTPURL, path))
	if err != nil {
		return
	}
	query, bodyBytes, err := c.renderSign(query, body)
	if err != nil {
		return statusCode, respBody, err
	}
	u.RawQuery = query.Encode()
	urlStr := u.String()
	req, err := http.NewRequestWithContext(ctx, method, urlStr, bytes.NewReader(bodyBytes))
	if err != nil {