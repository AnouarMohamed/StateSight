package render

import "encoding/json"

// JSON returns a stable JSON payload string for database storage.
func JSON(v any) (string, error) {
	body, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
