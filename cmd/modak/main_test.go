package main

import (
	"testing"
	"time"
)

func TestNotificationService(t *testing.T) {
	gateway := &Gateway{}
	service := NewNotificationServiceImpl(gateway)

	mockNow := time.Now()
	service.nowFunc = func() time.Time {
		return mockNow
	}

	tests := []struct {
		notificationType string
		userID           string
		message          string
		expectError      bool
		advanceTime      time.Duration
	}{
		{"news", "user1", "news 1", false, 0},
		{"news", "user1", "news 2", true, 0},
		{"news", "user1", "news 3", false, 24 * time.Hour},
		{"status", "user1", "status update 1", false, 0},
		{"status", "user1", "status update 2", false, 0},
		{"status", "user1", "status update 3", true, 0},
		{"status", "user1", "status update 4", false, time.Minute},
		{"marketing", "user1", "marketing 1", false, 0},
		{"marketing", "user1", "marketing 2", false, 0},
		{"marketing", "user1", "marketing 3", false, 0},
		{"marketing", "user1", "marketing 4", true, 0},               // Expect rate limit exceeded error here
		{"marketing", "user1", "marketing 5", true, 0},               // Expect rate limit exceeded error here
		{"marketing", "user1", "marketing 6", false, 24 * time.Hour}, // Expect works properly
	}

	for _, tt := range tests {
		if tt.advanceTime > 0 {
			mockNow = mockNow.Add(tt.advanceTime)
		}
		err := service.Send(tt.notificationType, tt.userID, tt.message)
		if tt.expectError && err == nil {
			t.Fatalf("expected an error but got none for %v", tt)
		}
		if !tt.expectError && err != nil {
			t.Fatalf("expected no error but got %v for %v", err, tt)
		}
	}
}
