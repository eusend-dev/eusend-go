package eusend

import (
	"context"
	"net/http"
)

// CreateTemplateRequest is the request object for Templates.Create. Use
// {{variable}} placeholders in Subject/Html; values are HTML-escaped at send time.
type CreateTemplateRequest struct {
	Name    string `json:"name"`
	Subject string `json:"subject"`
	Html    string `json:"html"`
}

// UpdateTemplateRequest updates a template. Nil fields are left unchanged.
type UpdateTemplateRequest struct {
	Name    *string `json:"name,omitempty"`
	Subject *string `json:"subject,omitempty"`
	Html    *string `json:"html,omitempty"`
}

// Template is a saved email template.
type Template struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Subject     string `json:"subject"`
	Html        string `json:"html"`
	ReactSource string `json:"reactSource"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// TemplateListItem is a row from Templates.List.
type TemplateListItem struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Subject   string `json:"subject"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// TemplatesSvc is the /templates API.
type TemplatesSvc interface {
	Create(params *CreateTemplateRequest) (*Template, error)
	CreateWithContext(ctx context.Context, params *CreateTemplateRequest) (*Template, error)
	List() ([]TemplateListItem, error)
	ListWithContext(ctx context.Context) ([]TemplateListItem, error)
	Get(templateId string) (*Template, error)
	GetWithContext(ctx context.Context, templateId string) (*Template, error)
	Update(templateId string, params *UpdateTemplateRequest) (*Template, error)
	UpdateWithContext(ctx context.Context, templateId string, params *UpdateTemplateRequest) (*Template, error)
	Remove(templateId string) (*GenericResponse, error)
	RemoveWithContext(ctx context.Context, templateId string) (*GenericResponse, error)
}

type TemplatesSvcImpl struct{ client *Client }

func (s *TemplatesSvcImpl) Create(params *CreateTemplateRequest) (*Template, error) {
	return s.CreateWithContext(context.Background(), params)
}

func (s *TemplatesSvcImpl) CreateWithContext(ctx context.Context, params *CreateTemplateRequest) (*Template, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, "templates", params)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(Template)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *TemplatesSvcImpl) List() ([]TemplateListItem, error) {
	return s.ListWithContext(context.Background())
}

func (s *TemplatesSvcImpl) ListWithContext(ctx context.Context) ([]TemplateListItem, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, "templates", nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	var resp struct {
		Data []TemplateListItem `json:"data"`
	}
	if _, err := s.client.Perform(req, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (s *TemplatesSvcImpl) Get(templateId string) (*Template, error) {
	return s.GetWithContext(context.Background(), templateId)
}

func (s *TemplatesSvcImpl) GetWithContext(ctx context.Context, templateId string) (*Template, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, "templates/"+templateId, nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(Template)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *TemplatesSvcImpl) Update(templateId string, params *UpdateTemplateRequest) (*Template, error) {
	return s.UpdateWithContext(context.Background(), templateId, params)
}

func (s *TemplatesSvcImpl) UpdateWithContext(ctx context.Context, templateId string, params *UpdateTemplateRequest) (*Template, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPatch, "templates/"+templateId, params)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(Template)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *TemplatesSvcImpl) Remove(templateId string) (*GenericResponse, error) {
	return s.RemoveWithContext(context.Background(), templateId)
}

func (s *TemplatesSvcImpl) RemoveWithContext(ctx context.Context, templateId string) (*GenericResponse, error) {
	req, err := s.client.NewRequest(ctx, http.MethodDelete, "templates/"+templateId, nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(GenericResponse)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
