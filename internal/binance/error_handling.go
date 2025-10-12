package binance

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// BinanceError represents a Binance API error
type BinanceError struct {
	Code    int
	Message string
	Retry   bool
}

// Error implements the error interface
func (e *BinanceError) Error() string {
	return fmt.Sprintf("Binance Error %d: %s", e.Code, e.Message)
}

// Error codes from Binance API
const (
	ErrCodeTimestampOutOfSync    = -1021
	ErrCodeInvalidSignature      = -1022
	ErrCodeUnauthorized          = -2015
	ErrCodeInsufficientBalance   = -2010
	ErrCodeMarginInsufficient    = -2019
	ErrCodePositionSideInvalid   = -4164
	ErrCodeRateLimitExceeded     = -1003
	ErrCodeIPBanned              = -1003
	ErrCodeOrderWouldTrigger     = -2021
	ErrCodeReduceOnlyReject      = -2022
)

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	BackoffFactor  float64
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		BackoffFactor:  2.0,
	}
}

// ExecuteWithRetry executes a function with retry logic
func ExecuteWithRetry(fn func() error, config *RetryConfig) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	var lastErr error
	backoff := config.InitialBackoff

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryableError(err) {
			return err
		}

		// Last attempt, don't sleep
		if attempt == config.MaxRetries {
			break
		}

		// Log retry attempt
		log.Printf("⚠️ Retry %d/%d after %v: %v", attempt+1, config.MaxRetries, backoff, err)

		// Sleep with exponential backoff
		time.Sleep(backoff)

		// Increase backoff
		backoff = time.Duration(float64(backoff) * config.BackoffFactor)
		if backoff > config.MaxBackoff {
			backoff = config.MaxBackoff
		}
	}

	return fmt.Errorf("max retries (%d) exceeded: %v", config.MaxRetries, lastErr)
}

// isRetryableError determines if an error should be retried
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	// Rate limit errors (429)
	if strings.Contains(errStr, "429") || strings.Contains(errStr, "too many requests") {
		log.Println("🚦 Rate limit hit, backing off...")
		return true
	}

	// Timeout errors
	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded") {
		log.Println("⏱️ Timeout error, retrying...")
		return true
	}

	// Connection errors
	if strings.Contains(errStr, "connection") || strings.Contains(errStr, "eof") {
		log.Println("🔌 Connection error, retrying...")
		return true
	}

	// Temporary server errors (5xx)
	if strings.Contains(errStr, "500") || strings.Contains(errStr, "502") ||
		strings.Contains(errStr, "503") || strings.Contains(errStr, "504") {
		log.Println("🔧 Server error, retrying...")
		return true
	}

	// Binance specific retryable errors
	if strings.Contains(errStr, "-1003") { // Rate limit
		log.Println("🚦 Binance rate limit, backing off...")
		return true
	}

	return false
}

// HandleBinanceError handles specific Binance error codes
func HandleBinanceError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Timestamp sync error
	if strings.Contains(errStr, "-1021") || strings.Contains(errStr, "timestamp") {
		return &BinanceError{
			Code:    ErrCodeTimestampOutOfSync,
			Message: "Timestamp out of sync with server. Please sync your system clock or use NTP.",
			Retry:   false,
		}
	}

	// Invalid signature
	if strings.Contains(errStr, "-1022") || strings.Contains(errStr, "signature") {
		return &BinanceError{
			Code:    ErrCodeInvalidSignature,
			Message: "Invalid API signature. Check your API secret key.",
			Retry:   false,
		}
	}

	// Insufficient balance
	if strings.Contains(errStr, "-2010") || strings.Contains(errStr, "insufficient balance") {
		return &BinanceError{
			Code:    ErrCodeInsufficientBalance,
			Message: "Insufficient balance to execute this order.",
			Retry:   false,
		}
	}

	// Margin insufficient
	if strings.Contains(errStr, "-2019") || strings.Contains(errStr, "margin") {
		return &BinanceError{
			Code:    ErrCodeMarginInsufficient,
			Message: "Insufficient margin. Reduce position size or add more margin.",
			Retry:   false,
		}
	}

	// Position side invalid
	if strings.Contains(errStr, "-4164") {
		return &BinanceError{
			Code:    ErrCodePositionSideInvalid,
			Message: "Position side does not match. Check your position mode (One-way/Hedge).",
			Retry:   false,
		}
	}

	// Rate limit
	if strings.Contains(errStr, "-1003") || strings.Contains(errStr, "429") {
		return &BinanceError{
			Code:    ErrCodeRateLimitExceeded,
			Message: "Rate limit exceeded. Backing off...",
			Retry:   true,
		}
	}

	// IP banned (418)
	if strings.Contains(errStr, "418") {
		return &BinanceError{
			Code:    ErrCodeIPBanned,
			Message: "IP has been auto-banned for continuing to send requests after 429. Stop all trading immediately.",
			Retry:   false,
		}
	}

	// Order would trigger immediately
	if strings.Contains(errStr, "-2021") {
		return &BinanceError{
			Code:    ErrCodeOrderWouldTrigger,
			Message: "Order would trigger immediately. Adjust your stop price.",
			Retry:   false,
		}
	}

	// Reduce-only rejected
	if strings.Contains(errStr, "-2022") {
		return &BinanceError{
			Code:    ErrCodeReduceOnlyReject,
			Message: "Reduce-only order rejected. This order would increase your position.",
			Retry:   false,
		}
	}

	return err
}

// LogBinanceError logs a user-friendly error message
func LogBinanceError(err error) {
	if err == nil {
		return
	}

	binanceErr, ok := err.(*BinanceError)
	if !ok {
		log.Printf("❌ Error: %v", err)
		return
	}

	emoji := "❌"
	if binanceErr.Code == ErrCodeRateLimitExceeded {
		emoji = "🚦"
	} else if binanceErr.Code == ErrCodeTimestampOutOfSync {
		emoji = "⏰"
	} else if binanceErr.Code == ErrCodeInsufficientBalance {
		emoji = "💰"
	}

	log.Printf("%s Binance Error [%d]: %s", emoji, binanceErr.Code, binanceErr.Message)
}

// GetErrorSuggestion provides a suggestion for fixing the error
func GetErrorSuggestion(err error) string {
	if err == nil {
		return ""
	}

	errStr := strings.ToLower(err.Error())

	suggestions := map[string]string{
		"-1021": "Sync your system clock using NTP: ntpdate pool.ntp.org",
		"-1022": "Verify your BINANCE_SECRET_KEY environment variable is correct",
		"-2010": "Check your account balance and reduce position size",
		"-2019": "Add more margin to your account or reduce leverage",
		"-4164": "Switch between One-way and Hedge mode in Binance settings",
		"429":   "Wait before sending more requests. Consider using WebSocket for real-time data",
		"418":   "CRITICAL: Stop all trading. Your IP is banned. Contact Binance support",
	}

	for code, suggestion := range suggestions {
		if strings.Contains(errStr, code) {
			return suggestion
		}
	}

	return "Check Binance API documentation for more details"
}

// CircuitBreaker implements circuit breaker pattern
type CircuitBreaker struct {
	maxFailures     int
	resetTimeout    time.Duration
	failures        int
	lastFailureTime time.Time
	state           string // "closed", "open", "half-open"
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        "closed",
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	// Check if circuit should be reset
	if cb.state == "open" && time.Since(cb.lastFailureTime) > cb.resetTimeout {
		cb.state = "half-open"
		cb.failures = 0
		log.Println("🔄 Circuit breaker: half-open (testing)")
	}

	// Block if circuit is open
	if cb.state == "open" {
		return fmt.Errorf("circuit breaker is open, rejecting request")
	}

	// Execute function
	err := fn()

	if err != nil {
		cb.failures++
		cb.lastFailureTime = time.Now()

		if cb.failures >= cb.maxFailures {
			cb.state = "open"
			log.Printf("⚠️ Circuit breaker: OPEN (too many failures: %d)", cb.failures)
		}

		return err
	}

	// Success - reset circuit
	if cb.state == "half-open" {
		cb.state = "closed"
		log.Println("✅ Circuit breaker: closed (recovered)")
	}
	cb.failures = 0

	return nil
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() string {
	return cb.state
}

// Reset manually resets the circuit breaker
func (cb *CircuitBreaker) Reset() {
	cb.state = "closed"
	cb.failures = 0
	log.Println("🔄 Circuit breaker manually reset")
}
