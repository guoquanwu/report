package binance

var Servers = "https://fapi.binance.com"

var ImportantPair = []string{
	"BTCUSDT",
	"ETHUSDT",
	"BCHUSDT",
	"DOTUSDT",
	"KSMUSDT",
	"LTCUSDT",
	"TRXUSDT",
	"UNIUSDT",
	"FILUSDT",
	"YFIUSDT",
	"LINKUSDT",
}

var FundingRate = "/fapi/v1/fundingRate?symbol=%s&limit=1"
