package eusend

import (
	"context"
	"net/http"
)

// Reputation states.
const (
	ReputationHealthy = "healthy"
	ReputationWarning = "warning"
	ReputationLimited = "limited"
	ReputationPaused  = "paused"
)

// ReputationBounce is the bounce rate as the sending limits judge it.
type ReputationBounce struct {
	// Rate is the raw rate over the judged window.
	Rate float64 `json:"rate"`
	// JudgedRate is the 95% confidence lower bound compared against the lines; 0 without
	// enough evidence.
	JudgedRate float64 `json:"judged_rate"`
	Counted    int     `json:"counted"`
	// NotCounted are bounces about eusend's sending infrastructure rather than the address.
	NotCounted int `json:"not_counted"`
	Recipients int `json:"recipients"`
	// Window is "7d", "30d", or nil when no window has enough evidence.
	Window    *string `json:"window"`
	LimitedAt float64 `json:"limited_at"`
	PausedAt  float64 `json:"paused_at"`
}

// ReputationComplaint is the raw 7-day complaint rate.
type ReputationComplaint struct {
	Rate       float64 `json:"rate"`
	Complaints int     `json:"complaints"`
	LimitedAt  float64 `json:"limited_at"`
	PausedAt   float64 `json:"paused_at"`
}

// ReputationNext is the next state and how many more counted bounces reach it. Bounces is
// nil when no realistic run of bounces would.
type ReputationNext struct {
	State   string `json:"state"`
	Bounces *int   `json:"bounces"`
}

// Reputation is the verdict the auto-suspend and the daily ceiling act on.
type Reputation struct {
	// State is healthy, warning, limited (daily ceiling halved) or paused (suspended).
	State string `json:"state"`
	// Measured is false until at least 100 recipients have a delivery result.
	Measured  bool                `json:"measured"`
	Bounce    ReputationBounce    `json:"bounce"`
	Complaint ReputationComplaint `json:"complaint"`
	// Next is nil when the account is paused.
	Next *ReputationNext `json:"next"`
}

// ReputationSvc is the /reputation API. Not available to sending-only keys; the API caches
// the answer for up to 30 seconds.
type ReputationSvc interface {
	Get() (*Reputation, error)
	GetWithContext(ctx context.Context) (*Reputation, error)
}

type ReputationSvcImpl struct{ client *Client }

func (s *ReputationSvcImpl) Get() (*Reputation, error) {
	return s.GetWithContext(context.Background())
}

func (s *ReputationSvcImpl) GetWithContext(ctx context.Context) (*Reputation, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, "reputation", nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(Reputation)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
