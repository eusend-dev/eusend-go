package eusend

import (
	"context"
	"net/http"
)

// BatchItemResult is one outcome of a batch send, positionally mapped to the
// input slice: result[i] describes params[i]. Queued items carry Id; rejected
// items carry Error and Code. Branch on `if result.Id != ""`.
type BatchItemResult struct {
	Id    string `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
	Code  string `json:"code,omitempty"`
}

// BatchEmailResponse is the response from Batch.Send.
type BatchEmailResponse struct {
	Data []BatchItemResult `json:"data"`
}

// BatchSvc is the /emails/batch API.
type BatchSvc interface {
	Send(params []*SendEmailRequest) (*BatchEmailResponse, error)
	SendWithContext(ctx context.Context, params []*SendEmailRequest) (*BatchEmailResponse, error)
}

type BatchSvcImpl struct{ client *Client }

// Send sends up to 100 emails in a single request. Attachments and scheduling
// are not supported on the batch endpoint and are stripped from each item —
// send those individually via Emails.Send. A failed item never fails the whole
// batch; branch on the presence of Id per returned result.
func (s *BatchSvcImpl) Send(params []*SendEmailRequest) (*BatchEmailResponse, error) {
	return s.SendWithContext(context.Background(), params)
}

func (s *BatchSvcImpl) SendWithContext(ctx context.Context, params []*SendEmailRequest) (*BatchEmailResponse, error) {
	// Strip fields the batch endpoint rejects, without mutating the caller's structs.
	payload := make([]SendEmailRequest, len(params))
	for i, p := range params {
		item := *p
		item.Attachments = nil
		item.ScheduledAt = ""
		payload[i] = item
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, "emails/batch", payload)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(BatchEmailResponse)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
