package handlers

import (
	"net/http"
	"stock-monitor/internal/together/finnhub"
	"stock-monitor/internal/together/scheduler"
	"stock-monitor/internal/together/store"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	store     *store.Store
	scheduler *scheduler.Scheduler
	client    *finnhub.Client
}

func New(s *store.Store, sh *scheduler.Scheduler, c *finnhub.Client) *Handler {
	return &Handler{
		store:     s,
		scheduler: sh,
		client:    c,
	}
}

// POST operation start-monitoring
type startMonitoringRequest struct {
	Symbol  string `json:"symbol" binding:"required"`
	Minutes int    `json:"minutes"`
	Seconds int    `json:"seconds"`
}

func (h *Handler) StartMonitoring(c *gin.Context) {
	var req startMonitoringRequest

	//receive request body through context and validate mapping to json
	err := c.ShouldBindJSON(&req)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//check for valid input symbol
	req.Symbol = strings.ToUpper(strings.TrimSpace(req.Symbol))
	if req.Symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol must not be empty"})
		return
	}

	//check for valid non-negative integers for time values
	if req.Minutes < 0 || req.Seconds < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "time can not be negative"})
		return
	}

	totalSeconds := req.Minutes*60 + req.Seconds
	if totalSeconds == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "total interval must be greater than 0"})
	}

	interval := time.Duration(totalSeconds) * time.Second

	//start background job
	h.scheduler.Start(req.Symbol, interval)

	c.JSON(http.StatusOK, gin.H{
		"message":  "monitoring started",
		"symbol":   req.Symbol,
		"interval": interval.String(),
	})
}

// GET /history?symbol=<stockSymbol>
func (h *Handler) GetHistory(c *gin.Context) {
	symbol := strings.ToUpper(strings.TrimSpace(c.Query("symbol")))

	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol must be provided"})
		return
	}
	history := h.store.GetHistory(symbol)
	c.JSON(http.StatusOK, history)
}

// POST /refresh
type refreshRequest struct {
	Symbol string `json:"symbol" binding:"required"`
}

func (h *Handler) Refresh(c *gin.Context) {
	var req refreshRequest

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Symbol = strings.ToUpper(strings.TrimSpace(req.Symbol))

	//get from finnhub
	quote, err := h.client.FetchQuote(req.Symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//store in history
	h.store.Append(quote)

	c.JSON(http.StatusOK, quote)
}

// DELETE /history?symbol=<stockSymbol>
func (h *Handler) Remove(c *gin.Context) {
	symbol := strings.ToUpper(strings.TrimSpace(c.Query("symbol")))

	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol must be provided"})
		return
	}

	respBool := h.store.Clear(symbol)
	if !respBool {
		c.JSON(http.StatusNotFound, gin.H{"error": "no history exists for symbol " + symbol})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "history cleared for " + symbol})
}
