package eusend

import (
	"context"
	"net/http"
)

// ContactProperty is one of the organization's declared custom properties -- the schema
// behind the Properties bag each Contact carries.
type ContactProperty struct {
	Id  string `json:"id"`
	Key string `json:"key"`
	// Type is "string" or "number". Values are carried and rendered as strings either
	// way; "number" buys a write-time check, not arithmetic or locale formatting.
	Type string `json:"type"`
	// FallbackValue is substituted into {{key}} for contacts that carry no value of
	// their own. Empty renders as empty, which is what an unset property did before the
	// registry existed.
	FallbackValue string `json:"fallbackValue"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

// CreateContactPropertyRequest is the request object for ContactProperties.Create.
type CreateContactPropertyRequest struct {
	// Key is lowercase letters, digits and underscores, starting with a letter; at most
	// 40 characters. "email", "name", "first_name", "last_name" and "full_name" are
	// built in and cannot be declared.
	Key string `json:"key"`
	// Type is "string" (the default when empty) or "number".
	Type string `json:"type,omitempty"`
	// FallbackValue must match Type. Nil leaves the property without a fallback.
	FallbackValue *string `json:"fallback_value,omitempty"`
}

// UpdateContactPropertyRequest changes a property's fallback.
//
// FallbackValue is the only mutable field: Key and Type are fixed at creation, because a
// rename would have to rewrite every contact AND every broadcast body spelling
// {{old_key}}. Changing either is a delete and a create.
type UpdateContactPropertyRequest struct {
	// A nil FallbackValue clears the fallback.
	FallbackValue *string `json:"fallback_value"`
}

// DeleteContactPropertyResponse reports the delete and how many contacts had the key
// stripped from them.
type DeleteContactPropertyResponse struct {
	Deleted         bool `json:"deleted"`
	ContactsUpdated int  `json:"contactsUpdated"`
}

// ContactPropertiesSvc is the /contact-properties API.
//
// Declaring a property is optional: one arriving on a contact write that has never been
// declared registers itself as a string, so an existing integration needs no changes.
// Declaring up front is what buys a type check and a fallback value.
type ContactPropertiesSvc interface {
	Create(params *CreateContactPropertyRequest) (*ContactProperty, error)
	CreateWithContext(ctx context.Context, params *CreateContactPropertyRequest) (*ContactProperty, error)
	List() ([]ContactProperty, error)
	ListWithContext(ctx context.Context) ([]ContactProperty, error)
	Get(propertyId string) (*ContactProperty, error)
	GetWithContext(ctx context.Context, propertyId string) (*ContactProperty, error)
	Update(propertyId string, params *UpdateContactPropertyRequest) (*ContactProperty, error)
	UpdateWithContext(ctx context.Context, propertyId string, params *UpdateContactPropertyRequest) (*ContactProperty, error)
	Remove(propertyId string) (*DeleteContactPropertyResponse, error)
	RemoveWithContext(ctx context.Context, propertyId string) (*DeleteContactPropertyResponse, error)
}

type ContactPropertiesSvcImpl struct{ client *Client }

func (s *ContactPropertiesSvcImpl) Create(params *CreateContactPropertyRequest) (*ContactProperty, error) {
	return s.CreateWithContext(context.Background(), params)
}

func (s *ContactPropertiesSvcImpl) CreateWithContext(ctx context.Context, params *CreateContactPropertyRequest) (*ContactProperty, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, "contact-properties", params)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(ContactProperty)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *ContactPropertiesSvcImpl) List() ([]ContactProperty, error) {
	return s.ListWithContext(context.Background())
}

// ListWithContext returns every property the organization has defined. Unpaginated -- the
// registry is capped at 100 entries, and a variable picker wants all of them.
func (s *ContactPropertiesSvcImpl) ListWithContext(ctx context.Context) ([]ContactProperty, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, "contact-properties", nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	var resp struct {
		Data []ContactProperty `json:"data"`
	}
	if _, err := s.client.Perform(req, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (s *ContactPropertiesSvcImpl) Get(propertyId string) (*ContactProperty, error) {
	return s.GetWithContext(context.Background(), propertyId)
}

func (s *ContactPropertiesSvcImpl) GetWithContext(ctx context.Context, propertyId string) (*ContactProperty, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, "contact-properties/"+propertyId, nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(ContactProperty)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *ContactPropertiesSvcImpl) Update(propertyId string, params *UpdateContactPropertyRequest) (*ContactProperty, error) {
	return s.UpdateWithContext(context.Background(), propertyId, params)
}

func (s *ContactPropertiesSvcImpl) UpdateWithContext(ctx context.Context, propertyId string, params *UpdateContactPropertyRequest) (*ContactProperty, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPatch, "contact-properties/"+propertyId, params)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(ContactProperty)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *ContactPropertiesSvcImpl) Remove(propertyId string) (*DeleteContactPropertyResponse, error) {
	return s.RemoveWithContext(context.Background(), propertyId)
}

// RemoveWithContext deletes the definition AND strips the key from every contact in the
// organization. Destructive and not undoable: leaving the values in place would keep the
// property rendering in broadcasts, and the next contact write carrying that key would
// register it again.
func (s *ContactPropertiesSvcImpl) RemoveWithContext(ctx context.Context, propertyId string) (*DeleteContactPropertyResponse, error) {
	req, err := s.client.NewRequest(ctx, http.MethodDelete, "contact-properties/"+propertyId, nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(DeleteContactPropertyResponse)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
