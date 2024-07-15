package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// NotificationService defines the interface for sending notifications.
type NotificationService interface {
	Send(notificationType, userID, message string) error
}

// NotificationServiceImpl implements NotificationService with rate limiting.
type NotificationServiceImpl struct {
	gateway     *Gateway
	rateLimits  map[string]rateLimit
	mu          sync.Mutex
	userRecords map[string]map[string][]time.Time
	nowFunc     func() time.Time
}

// rateLimit struct to hold the count and duration for rate limiting
type rateLimit struct {
	count    int
	duration time.Duration
}

// Gateway handles the actual sending of notifications.
type Gateway struct{}

// Send sends a message to a user.
func (g *Gateway) Send(userID, message string) {
	fmt.Printf("sending message to user %s: %s\n", userID, message)
}

// NewNotificationServiceImpl creates a new instance of NotificationServiceImpl.
func NewNotificationServiceImpl(gateway *Gateway) *NotificationServiceImpl {
	return &NotificationServiceImpl{
		gateway: gateway,
		rateLimits: map[string]rateLimit{
			"status":    {2, time.Minute},    // 2 per minute
			"news":      {1, 24 * time.Hour}, // 1 per day
			"marketing": {3, time.Hour},      // 3 per hour
		},
		userRecords: make(map[string]map[string][]time.Time),
		nowFunc:     time.Now,
	}
}

// Send sends a notification if it does not exceed the rate limit.
func (n *NotificationServiceImpl) Send(notificationType, userID, message string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	now := n.nowFunc()
	rateLimit, exists := n.rateLimits[notificationType]
	if !exists {
		return errors.New("unsupported notification type")
	}

	if _, ok := n.userRecords[userID]; !ok {
		n.userRecords[userID] = make(map[string][]time.Time)
	}

	records := n.userRecords[userID][notificationType]
	validRecords := []time.Time{}
	for _, record := range records {
		if record.After(now.Add(-rateLimit.duration)) {
			validRecords = append(validRecords, record)
		}
	}

	if len(validRecords) >= rateLimit.count {
		fmt.Printf("Rate limit exceeded for user %s: %s\n", userID, message)
		return errors.New("rate limit exceeded")
	}

	validRecords = append(validRecords, now)
	n.userRecords[userID][notificationType] = validRecords
	n.gateway.Send(userID, message)
	return nil
}
