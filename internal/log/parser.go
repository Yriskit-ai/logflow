// internal/log/parser.go
package log

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"
)

// Parser handles parsing of log lines
type Parser struct {
	levelPatterns    []*regexp.Regexp
	timestampPattern *regexp.Regexp
}

// NewParser creates a new log parser
func NewParser() *Parser {
	return &Parser{
		levelPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)\b(ERROR|ERR)\b`),
			regexp.MustCompile(`(?i)\b(WARN|WARNING)\b`),
			regexp.MustCompile(`(?i)\b(INFO|INFORMATION)\b`),
			regexp.MustCompile(`(?i)\b(DEBUG|DBG)\b`),
		},
		timestampPattern: regexp.MustCompile(`\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}`),
	}
}

// ParseLevel extracts the log level from a raw log line
func (p *Parser) ParseLevel(line string) LogLevel {
	upperLine := strings.ToUpper(line)

	// Check for ERROR patterns
	if strings.Contains(upperLine, "ERROR") || strings.Contains(upperLine, "ERR") {
		return LogLevelError
	}

	// Check for WARN patterns
	if strings.Contains(upperLine, "WARN") || strings.Contains(upperLine, "WARNING") {
		return LogLevelWarn
	}

	// Check for INFO patterns
	if strings.Contains(upperLine, "INFO") || strings.Contains(upperLine, "INFORMATION") {
		return LogLevelInfo
	}

	// Check for DEBUG patterns
	if strings.Contains(upperLine, "DEBUG") || strings.Contains(upperLine, "DBG") {
		return LogLevelDebug
	}

	// Default to INFO if no level detected
	return LogLevelInfo
}

// ParseStructured attempts to parse structured log formats (JSON, etc.)
func (p *Parser) ParseStructured(line string) map[string]interface{} {
	// Try to parse as JSON first
	if strings.HasPrefix(strings.TrimSpace(line), "{") {
		var jsonData map[string]interface{}
		if err := json.Unmarshal([]byte(line), &jsonData); err == nil {
			return p.normalizeJSONFields(jsonData)
		}
	}

	// Try to extract timestamp using regex
	result := make(map[string]interface{})
	if match := p.timestampPattern.FindString(line); match != "" {
		if ts, err := time.Parse("2006-01-02 15:04:05", match); err == nil {
			result["timestamp"] = ts
		} else if ts, err := time.Parse("2006-01-02T15:04:05", match); err == nil {
			result["timestamp"] = ts
		}
	}

	return result
}

// normalizeJSONFields normalizes common JSON log field names
func (p *Parser) normalizeJSONFields(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Common timestamp field names
	timestampFields := []string{"timestamp", "ts", "time", "@timestamp", "datetime"}
	for _, field := range timestampFields {
		if val, ok := data[field]; ok {
			if timeStr, ok := val.(string); ok {
				// Try common timestamp formats
				formats := []string{
					time.RFC3339,
					time.RFC3339Nano,
					"2006-01-02 15:04:05",
					"2006-01-02T15:04:05",
					"2006-01-02 15:04:05.000",
				}
				for _, format := range formats {
					if ts, err := time.Parse(format, timeStr); err == nil {
						result["timestamp"] = ts
						break
					}
				}
			}
			break
		}
	}

	// Common message field names
	messageFields := []string{"message", "msg", "text", "content"}
	for _, field := range messageFields {
		if val, ok := data[field]; ok {
			result["message"] = val
			break
		}
	}

	// Common level field names
	levelFields := []string{"level", "severity", "priority"}
	for _, field := range levelFields {
		if val, ok := data[field]; ok {
			result["level"] = val
			break
		}
	}

	// Copy other fields
	for k, v := range data {
		if _, exists := result[k]; !exists {
			result[k] = v
		}
	}

	return result
}
