package finnhub

import (
	"encoding/json"
	"fmt"
	"net/http"
	"stock-monitor/internal/together/models"
	"time"
)

// reuse client instead of making new one every request
type Client struct {
	apikey     string
	httpClient *http.Client
}

// constructor
func New(apikey string) *Client {
	return &Client{
		apikey:     apikey,
		httpClient: &http.Client{Timeout: 10 * time.Second}, //fails if taking longer than 10 seconds
	}
}

// JSON object finnhub returns
type finnhubQuoteResponse struct {
	C  float64 `json:"c"`
	H  float64 `json:"h"`
	L  float64 `json:"l"`
	O  float64 `json:"o"`
	PC float64 `json:"pc"`
}

func (c *Client) FetchQuote(symbol string) (models.StockQuote, error) {
	url := fmt.Sprintf(
		"https://finnhub.io/api/v1/quote?symbol=%s&token=%s",
		symbol,
		c.apikey,
	)

	response, err := c.httpClient.Get(url)

	if err != nil {
		return models.StockQuote{}, fmt.Errorf("http GET failed, %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return models.StockQuote{}, fmt.Errorf("finnhub return status %d", response.StatusCode)
	}

	var raw finnhubQuoteResponse
	err = json.NewDecoder(response.Body).Decode(&raw)

	if err != nil {
		return models.StockQuote{}, fmt.Errorf("failed to decode response %w", err)
	}

	return models.StockQuote{
		Symbol:       symbol,
		OpenPrice:    raw.O,
		HighPrice:    raw.H,
		LowPrice:     raw.L,
		CurrentPrice: raw.C,
		PrevClose:    raw.PC,
		FetchedAt:    time.Now().Unix(),
	}, nil
}
