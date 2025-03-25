package utility

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
)

func RespondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

func RespondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	var buf bytes.Buffer

	// Try encoding to the buffer first
	err := json.NewEncoder(&buf).Encode(data)
	if err != nil {
		http.Error(w, "Encoding error", http.StatusInternalServerError)
		return
	}

	// Check if the encoded result is just "null" (with possible whitespace/newline)
	encoded := buf.String()
	trimmed := strings.TrimSpace(encoded)

	if trimmed == "null" {
		// Handle the null case - for example, return an empty array instead
		json.NewEncoder(w).Encode(map[string]string{"message": "No pending requests for approval"})
	} else {
		// Not null, write the original encoded data
		json.NewEncoder(w).Encode(data)
	}
}
