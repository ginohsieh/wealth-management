// Package poller implements a configurable background goroutine that periodically
// fetches the latest market prices for all tracked symbols and updates the store.
//
// Supported sources:
//   - "yahoo"        – Yahoo Finance chart API (no API key required)
//   - "alphavantage" – Alpha Vantage GLOBAL_QUOTE endpoint (API key required)
package poller

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ginohsieh/wealth-management/backend/store"
)

var (
	mu      sync.Mutex
	stopCh  chan struct{}
	running bool
)

// Start launches the background polling goroutine if enabled.
// Safe to call multiple times; a running poller is stopped first.
func Start() {
	cfg, err := store.GetPriceConfig()
	if err != nil {
		log.Printf("[poller] could not read config: %v", err)
		return
	}
	if !cfg.Enabled {
		return
	}

	mu.Lock()
	defer mu.Unlock()

	stopRunning()

	stopCh = make(chan struct{})
	running = true
	go run(stopCh, cfg)
}

// Restart stops any running poller and starts a fresh one using the current config.
func Restart() {
	mu.Lock()
	stopRunning()
	mu.Unlock()
	Start()
}

// Stop halts the background poller.
func Stop() {
	mu.Lock()
	defer mu.Unlock()
	stopRunning()
}

// Refresh fetches prices for all tracked symbols immediately, once.
// It returns a map of symbol → new price and any errors encountered.
func Refresh() map[string]float64 {
	cfg, err := store.GetPriceConfig()
	if err != nil {
		log.Printf("[poller] could not read config: %v", err)
		return nil
	}
	symbols, err := store.GetTrackedSymbols()
	if err != nil {
		log.Printf("[poller] could not read symbols: %v", err)
		return nil
	}
	results := make(map[string]float64, len(symbols))
	for _, sym := range symbols {
		price, err := fetchPrice(sym, cfg)
		if err != nil {
			log.Printf("[poller] %s: %v", sym, err)
			continue
		}
		if err := store.SetSymbolPrice(sym, price); err != nil {
			log.Printf("[poller] %s: save price: %v", sym, err)
			continue
		}
		results[sym] = price
	}
	return results
}

// stopRunning must be called with mu held.
func stopRunning() {
	if running && stopCh != nil {
		close(stopCh)
		stopCh = nil
		running = false
	}
}

func run(stop <-chan struct{}, cfg store.PriceConfig) {
	interval := time.Duration(cfg.IntervalSeconds) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("[poller] started (source=%s interval=%ds)", cfg.Source, cfg.IntervalSeconds)

	for {
		select {
		case <-stop:
			log.Printf("[poller] stopped")
			return
		case <-ticker.C:
			// Re-read config on each tick to pick up interval/source changes
			// (a Restart() will be triggered for interval changes; this is belt-and-braces).
			latestCfg, err := store.GetPriceConfig()
			if err != nil {
				log.Printf("[poller] could not read config: %v", err)
				continue
			}
			if !latestCfg.Enabled {
				log.Printf("[poller] disabled, stopping ticker loop")
				return
			}
			symbols, err := store.GetTrackedSymbols()
			if err != nil {
				log.Printf("[poller] could not read symbols: %v", err)
				continue
			}
			for _, sym := range symbols {
				price, err := fetchPrice(sym, latestCfg)
				if err != nil {
					log.Printf("[poller] %s: %v", sym, err)
					continue
				}
				if err := store.SetSymbolPrice(sym, price); err != nil {
					log.Printf("[poller] %s: save price: %v", sym, err)
					continue
				}
				log.Printf("[poller] %s = %.4f", sym, price)
			}
		}
	}
}

// fetchPrice retrieves the current price for a single symbol from the configured source.
func fetchPrice(symbol string, cfg store.PriceConfig) (float64, error) {
	switch strings.ToLower(cfg.Source) {
	case "alphavantage":
		return fetchAlphaVantage(symbol, cfg.APIKey)
	default: // "yahoo" and anything else
		return fetchYahoo(symbol)
	}
}

// fetchYahoo uses the Yahoo Finance v8 chart API (no API key needed).
// URL: https://query1.finance.yahoo.com/v8/finance/chart/{symbol}?interval=1d&range=1d
func fetchYahoo(symbol string) (float64, error) {
	payload, err := fetchYahooPayload(symbol)
	if err != nil {
		return 0, err
	}
	price := payload.Chart.Result[0].Meta.RegularMarketPrice
	if price <= 0 {
		return 0, fmt.Errorf("yahoo: invalid price %.4f for %s", price, symbol)
	}
	return price, nil
}

// FetchSymbolName queries the Yahoo Finance chart API to get the company name for a symbol.
// It returns longName when available, then shortName, falling back to an error.
func FetchSymbolName(symbol string) (string, error) {
	payload, err := fetchYahooPayload(symbol)
	if err != nil {
		return "", err
	}
	meta := payload.Chart.Result[0].Meta
	if meta.LongName != "" {
		return meta.LongName, nil
	}
	if meta.ShortName != "" {
		return meta.ShortName, nil
	}
	return "", fmt.Errorf("yahoo: no name returned for %s", symbol)
}

// fetchYahooPayload is the shared HTTP + parse layer used by fetchYahoo and FetchSymbolName.
func fetchYahooPayload(symbol string) (*yahooChartPayload, error) {
	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v8/finance/chart/%s?interval=1d&range=1d",
		symbol,
	)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; WealthMgmt/1.0)")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("yahoo request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("yahoo HTTP %d for %s", resp.StatusCode, symbol)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxYahooBody))
	if err != nil {
		return nil, fmt.Errorf("yahoo read body: %w", err)
	}

	var payload yahooChartPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("yahoo parse: %w", err)
	}
	if payload.Chart.Error != nil {
		return nil, fmt.Errorf("yahoo API error %s: %s", payload.Chart.Error.Code, payload.Chart.Error.Description)
	}
	if len(payload.Chart.Result) == 0 {
		return nil, fmt.Errorf("yahoo: no data for %s", symbol)
	}
	return &payload, nil
}

// yahooChartPayload is the JSON structure returned by the Yahoo Finance chart API.
type yahooChartPayload struct {
	Chart struct {
		Result []struct {
			Meta struct {
				RegularMarketPrice float64 `json:"regularMarketPrice"`
				LongName           string  `json:"longName"`
				ShortName          string  `json:"shortName"`
			} `json:"meta"`
		} `json:"result"`
		Error *struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"error"`
	} `json:"chart"`
}

// fetchAlphaVantage uses the Alpha Vantage GLOBAL_QUOTE endpoint.
// URL: https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol={symbol}&apikey={key}
func fetchAlphaVantage(symbol, apiKey string) (float64, error) {
	if apiKey == "" {
		return 0, fmt.Errorf("alphavantage: API key not configured")
	}
	url := fmt.Sprintf(
		"https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=%s&apikey=%s",
		symbol, apiKey,
	)
	resp, err := httpClient.Get(url)
	if err != nil {
		return 0, fmt.Errorf("alphavantage request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxAlphaVantageBody))
	if err != nil {
		return 0, fmt.Errorf("alphavantage read body: %w", err)
	}

	// Parse: {"Global Quote": {"05. price": "185.0000"}}
	var payload struct {
		GlobalQuote map[string]string `json:"Global Quote"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return 0, fmt.Errorf("alphavantage parse: %w", err)
	}
	priceStr, ok := payload.GlobalQuote["05. price"]
	if !ok || priceStr == "" {
		return 0, fmt.Errorf("alphavantage: no price data for %s (check API key / rate limit)", symbol)
	}
	price, err := strconv.ParseFloat(strings.TrimSpace(priceStr), 64)
	if err != nil || price <= 0 {
		return 0, fmt.Errorf("alphavantage: invalid price %q for %s", priceStr, symbol)
	}
	return price, nil
}

// httpClient is shared across all requests with a reasonable timeout.
var httpClient = &http.Client{Timeout: 10 * time.Second}

// maxYahooBody is the maximum number of bytes read from a Yahoo Finance response.
const maxYahooBody = 64 * 1024

// maxAlphaVantageBody is the maximum number of bytes read from an Alpha Vantage response.
const maxAlphaVantageBody = 32 * 1024
