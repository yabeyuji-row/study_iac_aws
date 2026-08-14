package todo

import (
	"errors"
	"testing"
	"time"
)

func TestCursorRoundTrip(t *testing.T) {
	createdAt := time.Date(2026, 8, 3, 1, 2, 3, 0, time.FixedZone("JST", 9*3600))
	cursor := Cursor{
		CreatedAt: createdAt,
		ID:        "018f1a54-3c5b-4ef3-9ec4-8ff8e32180b8",
	}

	encoded, err := EncodeCursor(cursor)
	if err != nil {
		t.Fatalf("EncodeCursor() error = %v", err)
	}

	decoded, err := DecodeCursor(encoded)
	if err != nil {
		t.Fatalf("DecodeCursor() error = %v", err)
	}

	if !decoded.CreatedAt.Equal(createdAt.UTC()) {
		t.Fatalf("CreatedAt = %v, want %v", decoded.CreatedAt, createdAt.UTC())
	}
	if decoded.ID != cursor.ID {
		t.Fatalf("ID = %q, want %q", decoded.ID, cursor.ID)
	}
}

func TestDecodeCursorInvalid(t *testing.T) {
	_, err := DecodeCursor("not-base64")
	if !errors.Is(err, ErrInvalidCursor) {
		t.Fatalf("DecodeCursor() error = %v, want ErrInvalidCursor", err)
	}
}
