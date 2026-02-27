package models

type StockQuote struct {
	Symbol       string  `json:"symbol"`
	OpenPrice    float64 `json:"open_price"`
	HighPrice    float64 `json:"high_price"`
	LowPrice     float64 `json:"low_price"`
	CurrentPrice float64 `json:"current_price"`
	PrevClose    float64 `json:"prev_close"`
	FetchedAt    int64   `json:"fetched_at"`
}
