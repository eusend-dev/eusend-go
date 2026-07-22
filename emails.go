package eusend

import (
	"context"
	"net/http"
)

// SendEmailOptions carries per-call options for Emails.SendWithOptions.
type SendEmailOptions struct {
	// IdempotencyKey makes a send safe to retry: retrying with the same key
	// never sends a duplicate and returns the original email's ID.
	IdempotencyKey string `json:"-"`
}

// GetIdempotencyKey implements the Options interface.
func (o SendEmailOptions) GetIdempotencyKey() string { return o.IdempotencyKey }

// Attachment is a file attached to an email. Provide either Content (raw bytes,
// base64-encoded on the wire) or Path (a public URL fetched at send time). Set
// ContentId for an inline attachment referenced from HTML with <img src="cid:...">.
type Attachment struct {
	Content     []byte `json:"content,omitempty"`
	Filename    string `json:"filename,omitempty"`
	Path        string `json:"path,omitempty"`
	ContentType string `json:"content_type,omitempty"`
	ContentId   string `json:"content_id,omitempty"`
}

// SendEmailRequest is the request object for Emails.Send. From and To are
// required; provide at least one of Html, Text, or TemplateId.
type SendEmailRequest struct {
	// From accepts a bare email ("you@yourdomain.com") or a display-name form
	// ("Acme <you@yourdomain.com>"). The domain must be verified on your account.
	From        string            `json:"from"`
	To          []string          `json:"to"`
	Subject     string            `json:"subject,omitempty"`
	Bcc         []string          `json:"bcc,omitempty"`
	Cc          []string          `json:"cc,omitempty"`
	ReplyTo     []string          `json:"reply_to,omitempty"`
	Html        string            `json:"html,omitempty"`
	Text        string            `json:"text,omitempty"`
	TemplateId  string            `json:"template_id,omitempty"`
	Variables   map[string]any    `json:"variables,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	TrackOpens  *bool             `json:"track_opens,omitempty"`
	TrackClicks *bool             `json:"track_clicks,omitempty"`
	Attachments []*Attachment     `json:"attachments,omitempty"`
	// ScheduledAt schedules the send for a future time, at most 30 days out.
	// Accepts an ISO 8601 string or natural language ("in 1 hour", "tomorrow at
	// 9am"), parsed server-side in UTC. Not supported by Batch.Send.
	ScheduledAt string `json:"scheduled_at,omitempty"`
}

// SendEmailResponse is the response from Emails.Send.
type SendEmailResponse struct {
	Id string `json:"id"`
}

// UpdateEmailRequest reschedules a scheduled email.
type UpdateEmailRequest struct {
	Id          string `json:"-"`
	ScheduledAt string `json:"scheduled_at"`
}

// UpdateEmailResponse is the response from Emails.Update.
type UpdateEmailResponse struct {
	Id          string `json:"id"`
	Status      string `json:"status"`
	ScheduledAt string `json:"scheduled_at"`
}

// CancelScheduledEmailResponse is the response from Emails.Cancel.
type CancelScheduledEmailResponse struct {
	Id     string `json:"id"`
	Status string `json:"status"`
}

// EmailEvent is one entry in an email's delivery timeline.
type EmailEvent struct {
	Id        string         `json:"id"`
	Type      string         `json:"type"`
	Metadata  map[string]any `json:"metadata"`
	CreatedAt string         `json:"createdAt"`
}

// Email is the response from Emails.Get.
type Email struct {
	Id          string       `json:"id"`
	From        string       `json:"from"`
	To          []string     `json:"to"`
	Cc          []string     `json:"cc"`
	Bcc         []string     `json:"bcc"`
	ReplyTo     []string     `json:"replyTo"`
	Subject     string       `json:"subject"`
	Html        string       `json:"html"`
	Text        string       `json:"text"`
	Status      string       `json:"status"`
	TestMode    bool         `json:"testMode"`
	TemplateId  string       `json:"templateId"`
	ScheduledAt string       `json:"scheduledAt"`
	CreatedAt   string       `json:"createdAt"`
	Events      []EmailEvent `json:"events"`
}

// EmailListItem is a row from Emails.List.
type EmailListItem struct {
	Id        string   `json:"id"`
	From      string   `json:"from"`
	To        []string `json:"to"`
	Subject   string   `json:"subject"`
	Status    string   `json:"status"`
	TestMode  bool     `json:"testMode"`
	CreatedAt string   `json:"createdAt"`
}

// ListEmailsResponse is the response from Emails.List.
type ListEmailsResponse struct {
	Data       []EmailListItem `json:"data"`
	NextCursor string          `json:"next_cursor"`
}

// ListEmailsOptions filters Emails.List. Zero-valued fields are omitted.
type ListEmailsOptions struct {
	Limit  int
	Cursor string
	Status string
	From   string
	To     string
}

// EmailsSvc is the /emails API.
type EmailsSvc interface {
	Send(params *SendEmailRequest) (*SendEmailResponse, error)
	SendWithContext(ctx context.Context, params *SendEmailRequest) (*SendEmailResponse, error)
	SendWithOptions(ctx context.Context, params *SendEmailRequest, options *SendEmailOptions) (*SendEmailResponse, error)
	Get(emailId string) (*Email, error)
	GetWithContext(ctx context.Context, emailId string) (*Email, error)
	List(options *ListEmailsOptions) (*ListEmailsResponse, error)
	ListWithContext(ctx context.Context, options *ListEmailsOptions) (*ListEmailsResponse, error)
	Update(params *UpdateEmailRequest) (*UpdateEmailResponse, error)
	UpdateWithContext(ctx context.Context, params *UpdateEmailRequest) (*UpdateEmailResponse, error)
	Cancel(emailId string) (*CancelScheduledEmailResponse, error)
	CancelWithContext(ctx context.Context, emailId string) (*CancelScheduledEmailResponse, error)
}

type EmailsSvcImpl struct{ client *Client }

// Send sends a single email.
func (s *EmailsSvcImpl) Send(params *SendEmailRequest) (*SendEmailResponse, error) {
	return s.SendWithContext(context.Background(), params)
}

// SendWithContext is Send with a caller-supplied context.
func (s *EmailsSvcImpl) SendWithContext(ctx context.Context, params *SendEmailRequest) (*SendEmailResponse, error) {
	return s.SendWithOptions(ctx, params, nil)
}

// SendWithOptions is Send with a context and per-call options (e.g. an idempotency key).
func (s *EmailsSvcImpl) SendWithOptions(ctx context.Context, params *SendEmailRequest, options *SendEmailOptions) (*SendEmailResponse, error) {
	var opts Options
	if options != nil {
		opts = options
	}
	req, err := s.client.NewRequestWithOptions(ctx, http.MethodPost, "emails", params, opts)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(SendEmailResponse)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Get retrieves an email by ID, including its delivery events.
func (s *EmailsSvcImpl) Get(emailId string) (*Email, error) {
	return s.GetWithContext(context.Background(), emailId)
}

func (s *EmailsSvcImpl) GetWithContext(ctx context.Context, emailId string) (*Email, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, "emails/"+emailId, nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(Email)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// List returns a page of emails, most recent first.
func (s *EmailsSvcImpl) List(options *ListEmailsOptions) (*ListEmailsResponse, error) {
	return s.ListWithContext(context.Background(), options)
}

func (s *EmailsSvcImpl) ListWithContext(ctx context.Context, options *ListEmailsOptions) (*ListEmailsResponse, error) {
	q := map[string]string{}
	if options != nil {
		if options.Limit > 0 {
			q["limit"] = itoa(options.Limit)
		}
		q["cursor"] = options.Cursor
		q["status"] = options.Status
		q["from"] = options.From
		q["to"] = options.To
	}
	req, err := s.client.NewRequest(ctx, http.MethodGet, "emails"+queryString(q), nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(ListEmailsResponse)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Update reschedules a scheduled email. It fails once the email has started sending.
func (s *EmailsSvcImpl) Update(params *UpdateEmailRequest) (*UpdateEmailResponse, error) {
	return s.UpdateWithContext(context.Background(), params)
}

func (s *EmailsSvcImpl) UpdateWithContext(ctx context.Context, params *UpdateEmailRequest) (*UpdateEmailResponse, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPatch, "emails/"+params.Id, params)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(UpdateEmailResponse)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Cancel cancels a scheduled email. It fails once the email has started sending.
func (s *EmailsSvcImpl) Cancel(emailId string) (*CancelScheduledEmailResponse, error) {
	return s.CancelWithContext(context.Background(), emailId)
}

func (s *EmailsSvcImpl) CancelWithContext(ctx context.Context, emailId string) (*CancelScheduledEmailResponse, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, "emails/"+emailId+"/cancel", nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(CancelScheduledEmailResponse)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
