package model

import (
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
)

type Order struct {
	OrderID     uint64     `json:"order_id"`
	CustomerID  uuid.UUID  `json:"customer_id"`
	LineItems   []LineItem `json:"line_items"`
	CreatedAt   *time.Time `json:"created_at"`
	ShippedAt   *time.Time `json:"shipped_at"`
	CompletedAt *time.Time `json:"completed_at"`
	CancelledAt *time.Time `json:"cancelled_at"`
}

type Status string

const (
	StatusShipped   Status = "shipped"
	StatusCompleted Status = "completed"
	StatusCancelled Status = "cancelled"
)

var AllStatuses = []Status{StatusShipped, StatusCompleted, StatusCancelled}
var AllStatusesStrings = func() []string {
	statuses := make([]string, len(AllStatuses))
	for i, s := range AllStatuses {
		statuses[i] = string(s)
	}
	return statuses
}

type LineItem struct {
	ItemID   uuid.UUID `json:"item_id"`
	Quantity uint64    `json:"quantity" validate:"required,numeric,gte=0,lte=1000"`
	Price    uint64    `json:"price" validate:"required,numeric,gte=0,lte=1000000"`
}

func (o *Order) SetStatus(status Status) error {
	switch status {
	case StatusShipped:
		if o.ShippedAt != nil {
			return fmt.Errorf("order is already shipped")
		}

		now := time.Now().UTC()
		o.ShippedAt = &now
	case StatusCompleted:
		if o.ShippedAt == nil {
			return fmt.Errorf("order must be shipped before it can be completed")
		}
		if o.CompletedAt != nil {
			return fmt.Errorf("order is already completed")
		}
		if o.CancelledAt != nil {
			return fmt.Errorf("order is cancelled and cannot be completed")
		}
		now := time.Now().UTC()
		o.CompletedAt = &now
	case StatusCancelled:
		if o.CompletedAt != nil {
			return fmt.Errorf("order is already completed and cannot be cancelled")
		}
		if o.CancelledAt != nil {
			return fmt.Errorf("order is already cancelled")
		}
		now := time.Now().UTC()
		o.CompletedAt = &now
	default:
		return fmt.Errorf("invalid status: %q", status)
	}

	return nil
}

func ParseStatus(s string) (Status, error) {
	status := Status(s)
	if slices.Contains(AllStatuses, status) {
		return status, nil
	}
	return "", fmt.Errorf("invalid status: %q", s)
}
