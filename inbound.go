package eusend

import (
	"context"
	"net/http"
	"net/url"
)

// Inbound address statuses.
const (
	InboundActive           = "active"
	InboundPaidPlanRequired = "paid_plan_required"
	InboundOnHold           = "on_hold"
)

// InboundAddress is the workspace's receiving address. Mail to any address at Host is received.
type InboundAddress struct {
	// Host is <slug>.eusend.app.
	Host string `json:"host"`
	// Address is an example address, anything@<host>.
	Address string `json:"address"`
	// Status is active, paid_plan_required (the free plan) or on_hold (the account is under
	// review; mail is deferred rather than refused).
	Status string `json:"status"`
	Active bool   `json:"active"`
}

// InboundAttachment describes one attachment of a received message.
type InboundAttachment struct {
	// Index is the attachment's position, passed to Inbound.Attachment.
	Index       int     `json:"index"`
	Filename    *string `json:"filename"`
	ContentType string  `json:"contentType"`
	// Size is the decoded size in bytes.
	Size int `json:"size"`
	// ContentId is set for inline images referenced as cid: from the HTML.
	ContentId   *string `json:"contentId"`
	Disposition *string `json:"disposition"`
}

// InboundAuthResult is one authentication check as our receiver evaluated it. Result is pass,
// fail, none, neutral, softfail, temperror or permerror.
type InboundAuthResult struct {
	Result string            `json:"result"`
	Reason string            `json:"reason,omitempty"`
	Props  map[string]string `json:"props,omitempty"`
}

// InboundAuthentication holds the received message's SPF, DKIM (one per signature), DMARC and
// iprev results. Received mail is written by whoever sent it: check DMARC before trusting From.
type InboundAuthentication struct {
	Spf   *InboundAuthResult  `json:"spf,omitempty"`
	Dkim  []InboundAuthResult `json:"dkim,omitempty"`
	Dmarc *InboundAuthResult  `json:"dmarc,omitempty"`
	Iprev *InboundAuthResult  `json:"iprev,omitempty"`
}

// InboundHeader is one header line of a received message.
type InboundHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// InboundEmailListItem is one entry of Inbound.List.
type InboundEmailListItem struct {
	Id   string   `json:"id"`
	From *string  `json:"from"`
	To   []string `json:"to"`
	// ReceivedFor are your addresses the message was delivered to.
	ReceivedFor     []string `json:"receivedFor"`
	Subject         *string  `json:"subject"`
	Size            int      `json:"size"`
	AttachmentCount int      `json:"attachmentCount"`
	CreatedAt       string   `json:"createdAt"`
}

// InboundEmail is a received message, parsed from the stored original.
type InboundEmail struct {
	Id          string   `json:"id"`
	From        *string  `json:"from"`
	To          []string `json:"to"`
	Cc          []string `json:"cc"`
	ReplyTo     []string `json:"replyTo"`
	ReceivedFor []string `json:"receivedFor"`
	// EnvelopeFrom is the SMTP MAIL FROM.
	EnvelopeFrom *string `json:"envelopeFrom"`
	Subject      *string `json:"subject"`
	MessageId    *string `json:"messageId"`
	InReplyTo    *string `json:"inReplyTo"`
	// ReplyToEmailId is the email you sent through eusend that this message answers.
	ReplyToEmailId *string                `json:"replyToEmailId"`
	Text           *string                `json:"text"`
	Html           *string                `json:"html"`
	Headers        []InboundHeader        `json:"headers"`
	Attachments    []InboundAttachment    `json:"attachments"`
	Authentication *InboundAuthentication `json:"authentication"`
	// Size is the size of the original message in bytes.
	Size      int    `json:"size"`
	CreatedAt string `json:"createdAt"`
}

// ListInboundOptions filters Inbound.List. Zero-valued fields are omitted.
type ListInboundOptions struct {
	// Order is desc (newest first, default) or asc (oldest first). Keep it unchanged while paging.
	Order string
	// Limit is 1–100, default 10.
	Limit  int
	Cursor string
	// To keeps only messages received for this address.
	To string
}

// ListInboundResponse is a page of received mail.
type ListInboundResponse struct {
	Data       []InboundEmailListItem `json:"data"`
	NextCursor string                 `json:"next_cursor"`
}

// InboundSvc is the /inbound API: mail received at your eusend.app address and your domains'
// inbound hosts. Receiving is part of every paid plan.
type InboundSvc interface {
	Address() (*InboundAddress, error)
	AddressWithContext(ctx context.Context) (*InboundAddress, error)
	List(options *ListInboundOptions) (*ListInboundResponse, error)
	ListWithContext(ctx context.Context, options *ListInboundOptions) (*ListInboundResponse, error)
	Get(inboundEmailId string) (*InboundEmail, error)
	GetWithContext(ctx context.Context, inboundEmailId string) (*InboundEmail, error)
	Raw(inboundEmailId string) ([]byte, error)
	RawWithContext(ctx context.Context, inboundEmailId string) ([]byte, error)
	Attachment(inboundEmailId string, index int) ([]byte, error)
	AttachmentWithContext(ctx context.Context, inboundEmailId string, index int) ([]byte, error)
}

type InboundSvcImpl struct{ client *Client }

func (s *InboundSvcImpl) Address() (*InboundAddress, error) {
	return s.AddressWithContext(context.Background())
}

func (s *InboundSvcImpl) AddressWithContext(ctx context.Context) (*InboundAddress, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, "inbound/address", nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(InboundAddress)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *InboundSvcImpl) List(options *ListInboundOptions) (*ListInboundResponse, error) {
	return s.ListWithContext(context.Background(), options)
}

func (s *InboundSvcImpl) ListWithContext(ctx context.Context, options *ListInboundOptions) (*ListInboundResponse, error) {
	q := map[string]string{}
	if options != nil {
		if options.Limit > 0 {
			q["limit"] = itoa(options.Limit)
		}
		q["cursor"] = options.Cursor
		q["to"] = options.To
		q["order"] = options.Order
	}
	req, err := s.client.NewRequest(ctx, http.MethodGet, "inbound"+queryString(q), nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(ListInboundResponse)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *InboundSvcImpl) Get(inboundEmailId string) (*InboundEmail, error) {
	return s.GetWithContext(context.Background(), inboundEmailId)
}

func (s *InboundSvcImpl) GetWithContext(ctx context.Context, inboundEmailId string) (*InboundEmail, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, "inbound/"+url.PathEscape(inboundEmailId), nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(InboundEmail)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *InboundSvcImpl) Raw(inboundEmailId string) ([]byte, error) {
	return s.RawWithContext(context.Background(), inboundEmailId)
}

// RawWithContext returns the original message as received (message/rfc822).
func (s *InboundSvcImpl) RawWithContext(ctx context.Context, inboundEmailId string) ([]byte, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, "inbound/"+url.PathEscape(inboundEmailId)+"/raw", nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	req.Header.Set("Accept", "message/rfc822")
	return s.client.PerformRaw(req)
}

func (s *InboundSvcImpl) Attachment(inboundEmailId string, index int) ([]byte, error) {
	return s.AttachmentWithContext(context.Background(), inboundEmailId, index)
}

// AttachmentWithContext returns one attachment's bytes, by its Index.
func (s *InboundSvcImpl) AttachmentWithContext(ctx context.Context, inboundEmailId string, index int) ([]byte, error) {
	path := "inbound/" + url.PathEscape(inboundEmailId) + "/attachments/" + itoa(index)
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	req.Header.Set("Accept", "*/*")
	return s.client.PerformRaw(req)
}
