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
