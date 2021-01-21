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
	pairs     = []string{"WBNB-BUSD@PANCAKESWAP", "WBNB-USDT@PANCA