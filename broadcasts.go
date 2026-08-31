package eusend

import (
	"context"
	"net/http"
)

// CreateBroadcastRequest is the request object for Broadcasts.Create. Provide
// either Html or TemplateId.
type CreateBroadcastRequest struct {
	Name              string            `json:"name"`
	AudienceId        string            `json:"audience_id"`
	From              string            `json:"from"`
	Subject           string            `json:"subject"`
	Html              string            `json:"html,omitempty"`
	TemplateId        string            `json:"template_id,omitempty"`
	TemplateVariables map[string]string `json:"template_variables,omitempty"`
	// TrackOpens/TrackClicks are pointers so that an explicit false is sent rather than
	// dropped by omitempty. Nil means "use the organization default".
	TrackOpens  *bool `json:"track_opens,omitempty"`
	TrackClicks *bool `json:"track_clicks,omitempty"`
}

// UpdateBroadcastRequest updates a broadcast. Empty/nil fields are left unchanged.
type UpdateBroadcastRequest struct {
	Name              *string           `json:"name,omitempty"`
	AudienceId        *string           `json:"audience_id,omitempty"`
	From              *string           `json:"from,omitempty"`
	Subject           *string           `json:"subject,omitempty"`
	Html              *string           `json:"html,omitempty"`
	TemplateId        *string           `json:"template_id,omitempty"`
	TemplateVariables map[string]string `json:"template_variables,omitempty"`
	ScheduledAt       *string           `json:"scheduled_at,omitempty"`
	// Nil leaves the broadcast's current setting unchanged.
	TrackOpens  *bool `json:"track_opens,omitempty"`
	TrackClicks *bool `json:"track_clicks,omitempty"`
}

// SendBroadcastRequest sends or schedules a broadcast.
type SendBroadcastRequest struct {
	// ScheduledAt, when set, schedules the broadcast instead of sending immediately.
	ScheduledAt string `json:"scheduled_at,omitempty"`
}

// Broadcast is returned by Broadcasts.Create/Update/Cancel and (with stats) Get.
type Broadcast struct {
	Id                string            `json:"id"`
	OrganizationId    string            `json:"organizationId"`
	Name              string            `json:"name"`
	Status            string            `json:"status"`
	AudienceId        *string           `json:"audienceId"`
	FromAddress       string            `json:"fromAddress"`
	ReplyTo           *string           `json:"replyTo"`
	Subject           string            `json:"subject"`
	Html              *string           `json:"html"`
	ReactSource       *string           `json:"reactSource"`
	EditorJson        map[string]any    `json:"editorJson"`
	TemplateId        *string           `json:"templateId"`
	TemplateVariables map[string]string `json:"templateVariables"`
	HeldReason        *string           `json:"heldReason"`
	ScheduledAt       *string           `json:"scheduledAt"`
	TrackOpens        bool              `json:"trackOpens"`
	TrackClicks       bool              `json:"trackClicks"`
	RecipientCount    int               `json:"recipientCount"`
	SentCount         int               `json:"sentCount"`
	StartedAt         *string           `json:"startedAt"`
	CompletedAt       *string           `json:"completedAt"`
	// Stats is populated only by Broadcasts.Get.
	Stats             map[string]int    `json:"stats"`
	CreatedAt         string            `json:"createdAt"`
	UpdatedAt         string            `json:"updatedAt"`
}

// BroadcastListItem is a row from Broadcasts.List.
type BroadcastListItem struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	Status         string `json:"status"`
	AudienceId     *string `json:"audienceId"`
	FromAddress    string `json:"fromAddress"`
	Subject        string `json:"subject"`
	RecipientCount int    `json:"recipientCount"`
	SentCount      int    `json:"sentCount"`
	ScheduledAt    *string `json:"scheduledAt"`
	StartedAt      *string `json:"startedAt"`
	CompletedAt    *string `json:"completedAt"`
	CreatedAt      string `json:"createdAt"`
	AudienceName   *string `json:"audienceName"`
}

// SendBroadcastResponse is the response from Broadcasts.Send.
type SendBroadcastResponse struct {
	Id          string `json:"id"`
	Status      string `json:"status"`
	ScheduledAt *string `json:"scheduled_at"`
}

// TestBroadcastRequest is the payload for Broadcasts.Test.
type TestBroadcastRequest struct {
	// To holds up to 5 addresses, each on a domain verified on your account. A test send
	// delivers real mail without the paid-plan gate Send carries, so it is restricted to
	// inboxes you have already proved you control; anything else returns
	// DOMAIN_NOT_VERIFIED.
	To []string `json:"to"`
}

// TestBroadcastResponse is the response from Broadcasts.Test.
type TestBroadcastResponse struct {
	Id string `json:"id"`
	// SentTo lists the addresses actually mailed, lowercased and de-duplicated.
	SentTo []string `json:"sent_to"`
	// EmailIds holds one email id per recipient, for looking the delivery up in the logs.
	EmailIds []string `json:"email_ids"`
}

// BroadcastsSvc is the /broadcasts API.
type BroadcastsSvc interface {
	Create(params *CreateBroadcastRequest) (*Broadcast, error)
	CreateWithContext(ctx context.Context, params *CreateBroadcastRequest) (*Broadcast, error)
	List() ([]BroadcastListItem, error)
	ListWithContext(ctx context.Context) ([]BroadcastListItem, error)
	Get(broadcastId string) (*Broadcast, error)
	GetWithContext(ctx context.Context, broadcastId string) (*Broadcast, error)
	Update(broadcastId string, params *UpdateBroadcastRequest) (*Broadcast, error)
	UpdateWithContext(ctx context.Context, broadcastId string, params *UpdateBroadcastRequest) (*Broadcast, error)
	Send(broadcastId string, params *SendBroadcastRequest) (*SendBroadcastResponse, error)
	SendWithContext(ctx context.Context, broadcastId string, params *SendBroadcastRequest) (*SendBroadcastResponse, error)
	Test(broadcastId string, params *TestBroadcastRequest) (*TestBroadcastResponse, error)
	TestWithContext(ctx context.Context, broadcastId string, params *TestBroadcastRequest) (*TestBroadcastResponse, error)
	Cancel(broadcastId string) (*Broadcast, error)
	CancelWithContext(ctx context.Context, broadcastId string) (*Broadcast, error)
	Remove(broadcastId string) (*GenericResponse, error)
	RemoveWithContext(ctx context.Context, broadcastId string) (*GenericResponse, error)
}

type BroadcastsSvcImpl struct{ client *Client }

func (s *BroadcastsSvcImpl) Create(params *CreateBroadcastRequest) (*Broadcast, error) {
	return s.CreateWithContext(context.Background(), params)
}

func (s *BroadcastsSvcImpl) CreateWithContext(ctx context.Context, params *CreateBroadcastRequest) (*Broadcast, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, "broadcasts", params)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(Broadcast)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *BroadcastsSvcImpl) List() ([]BroadcastListItem, error) {
	return s.ListWithContext(context.Background())
}

func (s *BroadcastsSvcImpl) ListWithContext(ctx context.Context) ([]BroadcastListItem, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, "broadcasts", nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	var resp struct {
		Data []BroadcastListItem `json:"data"`
	}
	if _, err := s.client.Perform(req, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (s *BroadcastsSvcImpl) Get(broadcastId string) (*Broadcast, error) {
	return s.GetWithContext(context.Background(), broadcastId)
}

func (s *BroadcastsSvcImpl) GetWithContext(ctx context.Context, broadcastId string) (*Broadcast, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, "broadcasts/"+broadcastId, nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(Broadcast)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *BroadcastsSvcImpl) Update(broadcastId string, params *UpdateBroadcastRequest) (*Broadcast, error) {
	return s.UpdateWithContext(context.Background(), broadcastId, params)
}

func (s *BroadcastsSvcImpl) UpdateWithContext(ctx context.Context, broadcastId string, params *UpdateBroadcastRequest) (*Broadcast, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPatch, "broadcasts/"+broadcastId, params)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(Broadcast)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Send sends a broadcast immediately, or schedules it when params.ScheduledAt is
// set. Calling Send on a paused broadcast resumes it from where it stopped.
func (s *BroadcastsSvcImpl) Send(broadcastId string, params *SendBroadcastRequest) (*SendBroadcastResponse, error) {
	return s.SendWithContext(context.Background(), broadcastId, params)
}

func (s *BroadcastsSvcImpl) SendWithContext(ctx context.Context, broadcastId string, params *SendBroadcastRequest) (*SendBroadcastResponse, error) {
	if params == nil {
		params = &SendBroadcastRequest{}
	}
	req, err := s.client.NewRequest(ctx, http.MethodPost, "broadcasts/"+broadcastId+"/send", params)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(SendBroadcastResponse)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Test sends a copy of the broadcast to your own verified addresses — the real message
// through the real sending path, so it shows what a recipient will see.
//
// Unlike Send it works on every plan including Free. It costs daily and monthly quota like
// any other send, and does not move the broadcast's status: the campaign stays a draft no
// matter how many tests you send.
//
// Requires a LIVE api key. "Test" here means a dress rehearsal, not a sandbox — an
// eu_test_ key is refused because the mail really is delivered.
func (s *BroadcastsSvcImpl) Test(broadcastId string, params *TestBroadcastRequest) (*TestBroadcastResponse, error) {
	return s.TestWithContext(context.Background(), broadcastId, params)
}

func (s *BroadcastsSvcImpl) TestWithContext(ctx context.Context, broadcastId string, params *TestBroadcastRequest) (*TestBroadcastResponse, error) {
	if params == nil {
		params = &TestBroadcastRequest{}
	}
	req, err := s.client.NewRequest(ctx, http.MethodPost, "broadcasts/"+broadcastId+"/test", params)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(TestBroadcastResponse)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *BroadcastsSvcImpl) Cancel(broadcastId string) (*Broadcast, error) {
	return s.CancelWithContext(context.Background(), broadcastId)
}

func (s *BroadcastsSvcImpl) CancelWithContext(ctx context.Context, broadcastId string) (*Broadcast, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, "broadcasts/"+broadcastId+"/cancel", nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(Broadcast)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *BroadcastsSvcImpl) Remove(broadcastId string) (*GenericResponse, error) {
	return s.RemoveWithContext(context.Background(), broadcastId)
}

func (s *BroadcastsSvcImpl) RemoveWithContext(ctx context.Context, broadcastId string) (*GenericResponse, error) {
	req, err := s.client.NewRequest(ctx, http.MethodDelete, "broadcasts/"+broadcastId, nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(GenericResponse)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
