package todo

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

type Cursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
}

func EncodeCursor(cursor Cursor) (encoded string, err error) {
	cursor.CreatedAt = cursor.CreatedAt.UTC()

	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", fmt.Errorf("marshal cursor: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func DecodeCursor(encoded string) (cursor Cursor, err error) {
	payload, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return cursor, fmt.Errorf("%w: decode cursor", ErrInvalidCursor)
	}
	if err := json.Unmarshal(payload, &cursor); err != nil {
		return cursor, fmt.Errorf("%w: unmarshal cursor", ErrInvalidCursor)
	}
	if cursor.CreatedAt.IsZero() || cursor.ID == "" {
		return cursor, fmt.Errorf("%w: missing cursor fields", ErrInvalidCursor)
	}

	cursor.CreatedAt = cursor.CreatedAt.UTC()

	return cursor, nil
}
