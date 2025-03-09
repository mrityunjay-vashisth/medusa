package middleware

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// SecurityValidationMiddleware provides protection against common API attacks
func SecurityValidationMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip validation for non-GET requests that don't use query parameters
			if r.Method != http.MethodGet && r.Method != http.MethodDelete && len(r.URL.Query()) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			// Validate all query parameters
			query := r.URL.Query()
			for param, values := range query {
				for _, value := range values {
					// Skip empty values
					if value == "" {
						continue
					}

					// Length check to prevent very large inputs
					if len(value) > 1000 {
						logger.Warn("Query parameter exceeds maximum length",
							zap.String("param", param),
							zap.Int("length", len(value)),
							zap.String("ip", r.RemoteAddr))
						http.Error(w, "Query parameter too long", http.StatusBadRequest)
						return
					}

					// Check for common XSS patterns
					if containsXSSPatterns(value) {
						logger.Warn("Potential XSS attack detected",
							zap.String("param", param),
							zap.String("value", value),
							zap.String("ip", r.RemoteAddr))
						http.Error(w, "Invalid parameter value", http.StatusBadRequest)
						return
					}

					// Check for SQL injection patterns
					if containsSQLInjectionPatterns(value) {
						logger.Warn("Potential SQL injection attack detected",
							zap.String("param", param),
							zap.String("value", value),
							zap.String("ip", r.RemoteAddr))
						http.Error(w, "Invalid parameter value", http.StatusBadRequest)
						return
					}

					// Check for command injection patterns
					if containsCommandInjectionPatterns(value) {
						logger.Warn("Potential command injection attack detected",
							zap.String("param", param),
							zap.String("value", value),
							zap.String("ip", r.RemoteAddr))
						http.Error(w, "Invalid parameter value", http.StatusBadRequest)
						return
					}

					// Numeric parameter validation - if it looks like a number but contains invalid characters
					if looksLikeNumber(param) && !isValidNumber(value) {
						logger.Warn("Invalid numeric parameter",
							zap.String("param", param),
							zap.String("value", value),
							zap.String("ip", r.RemoteAddr))
						http.Error(w, fmt.Sprintf("Parameter '%s' must be a valid number", param), http.StatusBadRequest)
						return
					}

					// Date parameter validation
					if looksLikeDate(param) && !isValidDate(value) {
						logger.Warn("Invalid date parameter",
							zap.String("param", param),
							zap.String("value", value),
							zap.String("ip", r.RemoteAddr))
						http.Error(w, fmt.Sprintf("Parameter '%s' must be a valid date", param), http.StatusBadRequest)
						return
					}

					// ID parameter validation
					if looksLikeID(param) && !isValidID(value) {
						logger.Warn("Invalid ID parameter",
							zap.String("param", param),
							zap.String("value", value),
							zap.String("ip", r.RemoteAddr))
						http.Error(w, fmt.Sprintf("Parameter '%s' contains invalid characters", param), http.StatusBadRequest)
						return
					}
				}
			}

			// If all validations pass, proceed to the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// containsXSSPatterns checks for common XSS attack patterns
func containsXSSPatterns(value string) bool {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)<script`),
		regexp.MustCompile(`(?i)javascript:`),
		regexp.MustCompile(`(?i)on\w+\s*=`),
		regexp.MustCompile(`(?i)data:text/html`),
		regexp.MustCompile(`(?i)data:application/javascript`),
		regexp.MustCompile(`(?i)background-image:\s*url`),
		regexp.MustCompile(`(?i)expression\s*\(`),
		regexp.MustCompile(`(?i)<\s*img[^>]*src\s*=`),
		regexp.MustCompile(`(?i)<\s*iframe`),
	}

	for _, pattern := range patterns {
		if pattern.MatchString(value) {
			return true
		}
	}
	return false
}

// containsSQLInjectionPatterns checks for common SQL injection patterns
func containsSQLInjectionPatterns(value string) bool {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(\%27)|(\')|(--)|(\%23)|(#)`),
		regexp.MustCompile(`(?i)((\%3D)|(=))[^\n]*((\%27)|(\')|(\-\-)|(\%3B)|(\;))`),
		regexp.MustCompile(`(?i)\w*((\%27)|(\'))((\%6F)|o|(\%4F))((\%72)|r|(\%52))`),
		regexp.MustCompile(`(?i)((\%27)|(\'))union`),
		regexp.MustCompile(`(?i)((\%27)|(\'))select`),
		regexp.MustCompile(`(?i)((\%27)|(\'))insert`),
		regexp.MustCompile(`(?i)((\%27)|(\'))update`),
		regexp.MustCompile(`(?i)((\%27)|(\'))drop`),
	}

	for _, pattern := range patterns {
		if pattern.MatchString(value) {
			return true
		}
	}
	return false
}

// containsCommandInjectionPatterns checks for common command injection patterns
func containsCommandInjectionPatterns(value string) bool {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`[&|;$()<>` + "`" + `\\]`),
		regexp.MustCompile(`(?i)cat\s+/etc`),
		regexp.MustCompile(`(?i)ping\s+\-`),
		regexp.MustCompile(`(?i)wget\s+`),
		regexp.MustCompile(`(?i)curl\s+`),
		regexp.MustCompile(`(?i)whoami`),
	}

	for _, pattern := range patterns {
		if pattern.MatchString(value) {
			return true
		}
	}
	return false
}

// looksLikeNumber determines if a parameter likely expects numeric input
func looksLikeNumber(param string) bool {
	numericParams := []string{"id", "limit", "offset", "page", "size", "count", "duration",
		"age", "amount", "price", "quantity", "num", "number"}

	paramLower := strings.ToLower(param)
	for _, numParam := range numericParams {
		if strings.Contains(paramLower, numParam) {
			return true
		}
	}
	return false
}

// isValidNumber validates numeric input
func isValidNumber(value string) bool {
	// Check if it's a valid integer or float
	_, intErr := strconv.Atoi(value)
	_, floatErr := strconv.ParseFloat(value, 64)

	return intErr == nil || floatErr == nil
}

// looksLikeDate determines if a parameter likely expects date input
func looksLikeDate(param string) bool {
	dateParams := []string{"date", "time", "day", "month", "year", "created", "updated",
		"start", "end", "from", "to", "scheduled"}

	paramLower := strings.ToLower(param)
	for _, dateParam := range dateParams {
		if strings.Contains(paramLower, dateParam) {
			return true
		}
	}
	return false
}

// isValidDate validates date input in common formats
func isValidDate(value string) bool {
	formats := []string{
		time.RFC3339,
		"2006-01-02",
		"2006/01/02",
		"01/02/2006",
		"01-02-2006",
		"20060102",
	}

	for _, format := range formats {
		if _, err := time.Parse(format, value); err == nil {
			return true
		}
	}
	return false
}

// looksLikeID determines if a parameter likely expects ID input
func looksLikeID(param string) bool {
	return strings.HasSuffix(strings.ToLower(param), "id") ||
		param == "uuid" || param == "guid" || param == "key"
}

// isValidID validates ID input, allowing alphanumeric and some special characters
func isValidID(value string) bool {
	// UUID pattern (loose check)
	uuidPattern := regexp.MustCompile(`^[a-fA-F0-9\-]+$`)
	if uuidPattern.MatchString(value) {
		return true
	}

	// General ID pattern (alphanumeric, hyphens, underscores)
	idPattern := regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)
	return idPattern.MatchString(value)
}
