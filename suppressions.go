package eusend

import (
	"context"
	"net/http"
	"net/url"
)

// Suppression reasons. Bounce and complaint are written automatically by the platform;
// manual is what you add yourself.
const (
	SuppressionReasonBounce    = "bounce"
	SuppressionReasonComplaint = "complaint"
	SuppressionReasonManual    = "manual"
)

// SuppressionEntry is one address on the suppression list.
type SuppressionEntry struct {
	Id        string `json:"id"`
	Email     string `json:"email"`
	Reason    string `json:"reason"`
	CreatedAt string `json:"created_at"`
}

// CreateSuppressionRequest is the request object for Suppressions.Create.
type CreateSuppressionRequest struct {
	Email string `json:"email"`
	// Reason defaults to "manual" when empty. An add never overwrites the reason an
	// address is already suppressed for.
	Reason string `json:"reason,omitempty"`
}

// SuppressionImportItem is one entry in an import. Reason is optional and defaults to
// "manual".
type SuppressionImportItem struct {
	Email  string `json:"email"`
	Reason string `json:"reason,omitempty"`
}

// ListSuppressionsOptions filters Suppressions.List. Zero-valued fields are omitted.
type ListSuppressionsOptions struct {
	// Email matches addresses containing this substring. Pass a domain ("@acme.com")
	// to see every suppressed address there.
	Email  string
	Reason string
	Limit  int
	Cursor string
}

// ListSuppressionsResponse is a page of suppression entries, newest first.
type ListSuppressionsResponse struct {
	Data       []SuppressionEntry `json:"data"`
	NextCursor string             `json:"next_cursor"`
}

// ImportSuppressionsResponse reports what an import did. Count is what was written,
// AlreadySuppressed was on the list before, and Duplicates is how many rows the payload
// repeated — the three add up to the number of items sent.
type ImportSuppressionsResponse struct {
	Count             int `json:"count"`
	AlreadySuppressed int `json:"already_suppressed"`
	Duplicates        int `json:"duplicates"`
}

// RemoveSuppressionResponse reports how many entries were removed.
type RemoveSuppressionResponse struct {
	Deleted int `json:"deleted"`
}

// SuppressionsSvc is the /suppressions API — the addresses your organization will not
// send to. Hard bounces and spam complaints are added automatically; these methods cover
// the ones you manage yourself. Suppression applies to live sending only, so test-mode
// keys can read the list but not modify it.
type SuppressionsSvc interface {
	List(options *ListSuppressionsOptions) (*ListSuppressionsResponse, error)
	ListWithContext(ctx context.Context, options *ListSuppressionsOptions) (*ListSuppressionsResponse, error)
	Create(params *CreateSuppressionRequest) (*SuppressionEntry, error)
	CreateWithContext(ctx context.Context, params *CreateSuppressionRequest) (*SuppressionEntry, error)
	Import(items []*SuppressionImportItem) (*ImportSuppressionsResponse, error)
	ImportWithContext(ctx context.Context, items []*SuppressionImportItem) (*ImportSuppressionsResponse, error)
	Remove(idOrEmail string) (*RemoveSuppressionResponse, error)
	RemoveWithContext(ctx context.Context, idOrEmail string) (*RemoveSuppressionResponse, error)
	Export() ([]byte, error)
	ExportWithContext(ctx context.Context) ([]byte, error)
}

type SuppressionsSvcImpl struct{ client *Client }

func (s *SuppressionsSvcImpl) List(options *ListSuppressionsOptions) (*ListSuppressionsResponse, error) {
	return s.ListWithContext(context.Background(), options)
}

func (s *SuppressionsSvcImpl) ListWithContext(ctx context.Context, options *ListSuppressionsOptions) (*ListSuppressionsResponse, error) {
	q := map[string]string{}
	if options != nil {
		if options.Limit > 0 {
			q["limit"] = itoa(options.Limit)
		}
		q["email"] = options.Email
		q["reason"] = options.Reason
		q["cursor"] = options.Cursor
	}
	req, err := s.client.NewRequest(ctx, http.MethodGet, "suppressions"+queryString(q), nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(ListSuppressionsResponse)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *SuppressionsSvcImpl) Create(params *CreateSuppressionRequest) (*SuppressionEntry, error) {
	return s.CreateWithContext(context.Background(), params)
}

func (s *SuppressionsSvcImpl) CreateWithContext(ctx context.Context, params *CreateSuppressionRequest) (*SuppressionEntry, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, "suppressions", params)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(SuppressionEntry)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *SuppressionsSvcImpl) Import(items []*SuppressionImportItem) (*ImportSuppressionsResponse, error) {
	return s.ImportWithContext(context.Background(), items)
}

// ImportWithContext adds up to 1000 addresses in one call — for carrying a suppression
// list over from another provider before your first send.
func (s *SuppressionsSvcImpl) ImportWithContext(ctx context.Context, items []*SuppressionImportItem) (*ImportSuppressionsResponse, error) {
	body := map[string]any{"emails": items}
	req, err := s.client.NewRequest(ctx, http.MethodPost, "suppressions/batch", body)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(ImportSuppressionsResponse)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *SuppressionsSvcImpl) Remove(idOrEmail string) (*RemoveSuppressionResponse, error) {
	return s.RemoveWithContext(context.Background(), idOrEmail)
}

// RemoveWithContext un-suppresses by entry id or by address, making the address sendable
// again. Removing addresses that hard-bounced or complained is what damages a sender's
// reputation when done in bulk.
func (s *SuppressionsSvcImpl) RemoveWithContext(ctx context.Context, idOrEmail string) (*RemoveSuppressionResponse, error) {
	req, err := s.client.NewRequest(ctx, http.MethodDelete, "suppressions/"+url.PathEscape(idOrEmail), nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(RemoveSuppressionResponse)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *SuppressionsSvcImpl) Export() ([]byte, error) {
	return s.ExportWithContext(context.Background())
}

// ExportWithContext returns the whole list as CSV ("email,reason,created_at"), for backup
// or migration. Unlike every other method here the body is not JSON, so it comes back as
// raw bytes.
func (s *SuppressionsSvcImpl) ExportWithContext(ctx context.Context) ([]byte, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, "suppressions/export", nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	req.Header.Set("Accept", "text/csv")
	return s.client.PerformRaw(req)
}
