package eusend

import (
	"context"
	"net/http"
)

// CreateWebhookRequest is the request object for Webhooks.Create. Pass
// Events []string{"*"} to subscribe to every event.
type CreateWebhookRequest struct {
	Url    string   `json:"url"`
	Events []string `json:"events"`
}

// UpdateWebhookRequest updates a webhook. Nil/empty fields are left unchanged.
type UpdateWebhookRequest struct {
	Url    *string  `json:"url,omitempty"`
	Events []string `json:"events,omitempty"`
}

// WebhookDelivery is one delivery attempt of a webhook event.
type WebhookDelivery struct {
	Id             string         `json:"id"`
	WebhookId      string         `json:"webhookId"`
	EmailId        string         `json:"emailId"`
	EventType      string         `json:"eventType"`
	Payload        map[string]any `json:"payload"`
	Status         string         `json:"status"`
	ResponseStatus int            `json:"responseStatus"`
	Attempts       int            `json:"attempts"`
	CreatedAt      string         `json:"createdAt"`
	LastAttemptAt  string         `json:"lastAttemptAt"`
}

// Webhook is a webhook subscription. Secret is populated only by Create;
// Deliveries only by Get.
type Webhook struct {
	Id         string            `json:"id"`
	Url        string            `json:"url"`
	Events     []string          `json:"events"`
	Secret     string            `json:"secret,omitempty"`
	CreatedAt  string            `json:"createdAt"`
	Deliveries []WebhookDelivery `json:"deliveries,omitempty"`
}

// WebhooksSvc is the /webhooks API.
type WebhooksSvc interface {
	Create(params *CreateWebhookRequest) (*Webhook, error)
	CreateWithContext(ctx context.Context, params *CreateWebhookRequest) (*Webhook, error)
	List() ([]Webhook, error)
	ListWithContext(ctx context.Context) ([]Webhook, error)
	Get(webhookId string) (*Webhook, error)
	GetWithContext(ctx context.Context, webhookId string) (*Webhook, error)
	Update(webhookId string, params *UpdateWebhookRequest) (*Webhook, error)
	UpdateWithContext(ctx context.Context, webhookId string, params *UpdateWebhookRequest) (*Webhook, error)
	Remove(webhookId string) (*GenericResponse, error)
	RemoveWithContext(ctx context.Context, webhookId string) (*GenericResponse, error)
}

type WebhooksSvcImpl struct{ client *Client }

func (s *WebhooksSvcImpl) Create(params *CreateWebhookRequest) (*Webhook, error) {
	return s.CreateWithContext(context.Background(), params)
}

func (s *WebhooksSvcImpl) CreateWithContext(ctx context.Context, params *CreateWebhookRequest) (*Webhook, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, "webhooks", params)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(Webhook)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *WebhooksSvcImpl) List() ([]Webhook, error) {
	return s.ListWithContext(context.Background())
}

func (s *WebhooksSvcImpl) ListWithContext(ctx context.Context) ([]Webhook, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, "webhooks", nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	var resp struct {
		Data []Webhook `json:"data"`
	}
	if _, err := s.client.Perform(req, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (s *WebhooksSvcImpl) Get(webhookId string) (*Webhook, error) {
	return s.GetWithContext(context.Background(), webhookId)
}

func (s *WebhooksSvcImpl) GetWithContext(ctx context.Context, webhookId string) (*Webhook, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, "webhooks/"+webhookId, nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(Webhook)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *WebhooksSvcImpl) Update(webhookId string, params *UpdateWebhookRequest) (*Webhook, error) {
	return s.UpdateWithContext(context.Background(), webhookId, params)
}

func (s *WebhooksSvcImpl) UpdateWithContext(ctx context.Context, webhookId string, params *UpdateWebhookRequest) (*Webhook, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPatch, "webhooks/"+webhookId, params)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(Webhook)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *WebhooksSvcImpl) Remove(webhookId string) (*GenericResponse, error) {
	return s.RemoveWithContext(context.Background(), webhookId)
}

func (s *WebhooksSvcImpl) RemoveWithContext(ctx context.Context, webhookId string) (*GenericResponse, error) {
	req, err := s.client.NewRequest(ctx, http.MethodDelete, "webhooks/"+webhookId, nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(GenericResponse)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
