package eusend

import (
	"context"
	"net/http"
)

// What a key may reach: PermissionFullAccess for every resource,
// PermissionSendingAccess for sending email only.
const (
	PermissionFullAccess    = "full_access"
	PermissionSendingAccess = "sending_access"
)

// CreateApiKeyRequest is the request object for ApiKeys.Create.
type CreateApiKeyRequest struct {
	Name string `json:"name"`
	// TestMode issues a sandbox key. Emails sent with a test key are accepted
	// and tracked but never delivered.
	TestMode bool `json:"test_mode"`
	// Permission defaults to PermissionFullAccess when empty.
	Permission string `json:"permission,omitempty"`
	// DomainId restricts the key to sending from a single domain. Only valid
	// alongside PermissionSendingAccess; leave empty for any verified domain.
	DomainId string `json:"domain_id,omitempty"`
}

// CreateApiKeyResponse is returned by ApiKeys.Create. Key holds the full secret
// and is returned only once — store it securely.
type CreateApiKeyResponse struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Key        string `json:"key"`
	Prefix     string `json:"prefix"`
	TestMode   bool   `json:"test_mode"`
	Permission string `json:"permission"`
	DomainId   string `json:"domain_id"`
	DomainName string `json:"domain_name"`
	CreatedAt  string `json:"created_at"`
}

// ApiKey is a row from ApiKeys.List. The full key is never returned after creation.
type ApiKey struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Prefix     string `json:"prefix"`
	TestMode   bool   `json:"test_mode"`
	Permission string `json:"permission"`
	DomainId   string `json:"domain_id"`
	DomainName string `json:"domain_name"`
	CreatedAt  string `json:"created_at"`
	LastUsedAt string `json:"last_used_at"`
}

// ApiKeysSvc is the /api-keys API.
type ApiKeysSvc interface {
	Create(params *CreateApiKeyRequest) (*CreateApiKeyResponse, error)
	CreateWithContext(ctx context.Context, params *CreateApiKeyRequest) (*CreateApiKeyResponse, error)
	List() ([]ApiKey, error)
	ListWithContext(ctx context.Context) ([]ApiKey, error)
	Remove(apiKeyId string) (*GenericResponse, error)
	RemoveWithContext(ctx context.Context, apiKeyId string) (*GenericResponse, error)
}

type ApiKeysSvcImpl struct{ client *Client }

func (s *ApiKeysSvcImpl) Create(params *CreateApiKeyRequest) (*CreateApiKeyResponse, error) {
	return s.CreateWithContext(context.Background(), params)
}

func (s *ApiKeysSvcImpl) CreateWithContext(ctx context.Context, params *CreateApiKeyRequest) (*CreateApiKeyResponse, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, "api-keys", params)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(CreateApiKeyResponse)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *ApiKeysSvcImpl) List() ([]ApiKey, error) {
	return s.ListWithContext(context.Background())
}

func (s *ApiKeysSvcImpl) ListWithContext(ctx context.Context) ([]ApiKey, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, "api-keys", nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	var resp []ApiKey
	if _, err := s.client.Perform(req, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *ApiKeysSvcImpl) Remove(apiKeyId string) (*GenericResponse, error) {
	return s.RemoveWithContext(context.Background(), apiKeyId)
}

func (s *ApiKeysSvcImpl) RemoveWithContext(ctx context.Context, apiKeyId string) (*GenericResponse, error) {
	req, err := s.client.NewRequest(ctx, http.MethodDelete, "api-keys/"+apiKeyId, nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(GenericResponse)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
