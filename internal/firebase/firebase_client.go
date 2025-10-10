package firebase

import (
	"bytes"
	"context"
	"crypto-trading-api/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2/google"
)

type Client struct {
	databaseURL string
	authToken   string
	httpClient  *http.Client
}

func InitClient() (*Client, error) {
	ctx := context.Background()

	// Firebase config
	databaseURL := os.Getenv("FIREBASE_DATABASE_URL")
	credentialsFile := os.Getenv("FIREBASE_CREDENTIALS_FILE")

	if databaseURL == "" {
		return nil, fmt.Errorf("FIREBASE_DATABASE_URL must be set")
	}

	// Remove trailing slash if present
	databaseURL = strings.TrimRight(databaseURL, "/")

	// Get OAuth access token using service account credentials
	var authToken string
	if credentialsFile != "" {
		// Set credentials file as environment variable for Google Default Credentials
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credentialsFile)

		// Get access token using default credentials
		tokenSource, err := google.DefaultTokenSource(ctx, "https://www.googleapis.com/auth/firebase.database")
		if err != nil {
			log.Printf("Warning: Could not get token source: %v", err)
		} else {
			token, err := tokenSource.Token()
			if err != nil {
				log.Printf("Warning: Could not get access token: %v", err)
			} else {
				authToken = token.AccessToken
			}
		}
	}

	httpClient := &http.Client{}

	log.Printf("✅ Firebase client initialized successfully")
	log.Printf("   Database URL: %s", databaseURL)
	if authToken != "" {
		log.Printf("   Auth: ✅ Access token obtained")
	} else {
		log.Printf("   Auth: ⚠️  No access token (using unauthenticated requests)")
	}

	return &Client{
		databaseURL: databaseURL,
		authToken:   authToken,
		httpClient:  httpClient,
	}, nil
}

// makeRequest makes an HTTP request to Firebase REST API
func (f *Client) makeRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s%s.json", f.databaseURL, path)

	// Add auth parameter if we have a token
	if f.authToken != "" {
		if strings.Contains(url, "?") {
			url += "&auth=" + f.authToken
		} else {
			url += "?auth=" + f.authToken
		}
	}

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("firebase request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// SaveTrade - Save trade to Firebase
func (f *Client) SaveTrade(ctx context.Context, trade *models.Trade) error {
	// Save to main trades collection
	path := fmt.Sprintf("/trades/%s", trade.ID)
	_, err := f.makeRequest(ctx, "PUT", path, trade)
	if err != nil {
		return fmt.Errorf("failed to save trade: %v", err)
	}

	// Also save under user's trades for easy querying
	userTradePath := fmt.Sprintf("/users/%s/trades/%s", trade.UserID, trade.ID)
	_, err = f.makeRequest(ctx, "PUT", userTradePath, trade)
	if err != nil {
		log.Printf("Warning: Failed to save trade under user: %v", err)
	}

	return nil
}

// UpdateTrade - Update existing trade
func (f *Client) UpdateTrade(ctx context.Context, trade *models.Trade) error {
	// Update main trade
	path := fmt.Sprintf("/trades/%s", trade.ID)
	_, err := f.makeRequest(ctx, "PUT", path, trade)
	if err != nil {
		return fmt.Errorf("failed to update trade: %v", err)
	}

	// Also update under user's trades
	userTradePath := fmt.Sprintf("/users/%s/trades/%s", trade.UserID, trade.ID)
	_, err = f.makeRequest(ctx, "PUT", userTradePath, trade)
	if err != nil {
		log.Printf("Warning: Failed to update trade under user: %v", err)
	}

	return nil
}

// GetTrade - Get single trade by ID
func (f *Client) GetTrade(ctx context.Context, tradeID string) (*models.Trade, error) {
	path := fmt.Sprintf("/trades/%s", tradeID)
	respBody, err := f.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get trade: %v", err)
	}

	// Check if response is null (trade doesn't exist)
	if string(respBody) == "null" || string(respBody) == "" {
		return nil, fmt.Errorf("trade not found")
	}

	var trade models.Trade
	if err := json.Unmarshal(respBody, &trade); err != nil {
		return nil, fmt.Errorf("failed to unmarshal trade: %v", err)
	}

	return &trade, nil
}

// GetUserTrades - Get all trades for a user
func (f *Client) GetUserTrades(ctx context.Context, userID string) ([]*models.Trade, error) {
	path := fmt.Sprintf("/users/%s/trades", userID)
	respBody, err := f.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get user trades: %v", err)
	}

	// Check if response is null (no trades)
	if string(respBody) == "null" || string(respBody) == "" {
		return []*models.Trade{}, nil
	}

	var tradesMap map[string]*models.Trade
	if err := json.Unmarshal(respBody, &tradesMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal trades: %v", err)
	}

	// Convert map to slice
	trades := make([]*models.Trade, 0, len(tradesMap))
	for _, trade := range tradesMap {
		trades = append(trades, trade)
	}

	return trades, nil
}

// GetActiveTrades - Get all active trades for monitoring
func (f *Client) GetActiveTrades(ctx context.Context) ([]*models.Trade, error) {
	// Get all trades
	path := "/trades"
	respBody, err := f.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get trades: %v", err)
	}

	// Check if response is null
	if string(respBody) == "null" || string(respBody) == "" {
		return []*models.Trade{}, nil
	}

	var tradesMap map[string]*models.Trade
	if err := json.Unmarshal(respBody, &tradesMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal trades: %v", err)
	}

	// Filter active trades
	trades := make([]*models.Trade, 0)
	for _, trade := range tradesMap {
		if trade.Status == "ACTIVE" {
			trades = append(trades, trade)
		}
	}

	return trades, nil
}

// DeleteTrade - Delete a trade
func (f *Client) DeleteTrade(ctx context.Context, tradeID string, userID string) error {
	// Delete from main trades
	path := fmt.Sprintf("/trades/%s", tradeID)
	_, err := f.makeRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete trade: %v", err)
	}

	// Delete from user's trades
	userTradePath := fmt.Sprintf("/users/%s/trades/%s", userID, tradeID)
	_, err = f.makeRequest(ctx, "DELETE", userTradePath, nil)
	if err != nil {
		log.Printf("Warning: Failed to delete trade from user: %v", err)
	}

	return nil
}

// Close - Close Firebase client
func (f *Client) Close() error {
	// HTTP client doesn't require explicit closing
	return nil
}
