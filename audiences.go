package eusend

import (
	"context"
	"net/http"
)

// CreateAudienceRequest is the request object for Audiences.Create.
type CreateAudienceRequest struct {
	Name string `json:"name"`
}

// Audience is returned by Audiences.Create.
type Audience struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	OrganizationId string `json:"organizationId"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

// AudienceListItem is a row from Audiences.List.
type AudienceListItem struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	CreatedAt    string `json:"createdAt"`
	ContactCount int    `json:"contactCount"`
}

// Contact is a member of an audience.
type Contact struct {
	Id             string `json:"id"`
	AudienceId     string `json:"audienceId"`
	Email          string `json:"email"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	Status         string `json:"status"`
	UnsubscribedAt string `json:"unsubscribedAt"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

// CreateContactRequest is the request object for Audiences.CreateContact and the
// item shape for BatchCreateContacts.
type CreateContactRequest struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`

	// Unsubscribed and CreatedAt are accepted by BatchCreateContacts only — they exist
	// for migrating a list in from another provider. CreateContact rejects them.
	//
	// Unsubscribed marks the contact as opted out. An import can only ever ADD an
	// opt-out: false will not re-subscribe someone who has already unsubscribed, which
	// is a consent decision and stays on UpdateContact.
	Unsubscribed *bool `json:"unsubscribed,omitempty"`
	// CreatedAt is the original signup time (ISO 8601). Applied on insert only; an
	// existing contact keeps the date it already has.
	CreatedAt string `json:"created_at,omitempty"`
}

// UpdateContactRequest updates a contact. Nil fields are left unchanged.
type UpdateContactRequest struct {
	FirstName    *string `json:"first_name,omitempty"`
	LastName     *string `json:"last_name,omitempty"`
	Unsubscribed *bool   `json:"unsubscribed,omitempty"`
}

// ListContactsResponse is a page of contacts.
type ListContactsResponse struct {
	Data       []Contact `json:"data"`
	NextCursor string    `json:"next_cursor"`
}

// ListContactsOptions filters Audiences.ListContacts. Zero-valued fields are omitted.
type ListContactsOptions struct {
	Limit      int
	Cursor     string
	Search     string
	Subscribed *bool
}

// BatchCreateContactsResponse reports how many contacts were written.
type BatchCreateContactsResponse struct {
	Count int `json:"count"`
	// Duplicates is how many repeated addresses were collapsed to reach Count, which
	// is what explains a count lower than the number of rows sent.
	Duplicates int `json:"duplicates"`
}

// AudiencesSvc is the /audiences API, including nested contact operations.
type AudiencesSvc interface {
	Create(params *CreateAudienceRequest) (*Audience, error)
	CreateWithContext(ctx context.Context, params *CreateAudienceRequest) (*Audience, error)
	List() ([]AudienceListItem, error)
	ListWithContext(ctx context.Context) ([]AudienceListItem, error)
	Remove(audienceId string) (*GenericResponse, error)
	RemoveWithContext(ctx context.Context, audienceId string) (*GenericResponse, error)

	CreateContact(audienceId string, params *CreateContactRequest) (*Contact, error)
	CreateContactWithContext(ctx context.Context, audienceId string, params *CreateContactRequest) (*Contact, error)
	BatchCreateContacts(audienceId string, contacts []*CreateContactRequest) (*BatchCreateContactsResponse, error)
	BatchCreateContactsWithContext(ctx context.Context, audienceId string, contacts []*CreateContactRequest) (*BatchCreateContactsResponse, error)
	ListContacts(audienceId string, options *ListContactsOptions) (*ListContactsResponse, error)
	ListContactsWithContext(ctx context.Context, audienceId string, options *ListContactsOptions) (*ListContactsResponse, error)
	GetContact(audienceId, contactId string) (*Contact, error)
	GetContactWithContext(ctx context.Context, audienceId, contactId string) (*Contact, error)
	UpdateContact(audienceId, contactId string, params *UpdateContactRequest) (*Contact, error)
	UpdateContactWithContext(ctx context.Context, audienceId, contactId string, params *UpdateContactRequest) (*Contact, error)
	RemoveContact(audienceId, contactId string) (*GenericResponse, error)
	RemoveContactWithContext(ctx context.Context, audienceId, contactId string) (*GenericResponse, error)
}

type AudiencesSvcImpl struct{ client *Client }

func (s *AudiencesSvcImpl) Create(params *CreateAudienceRequest) (*Audience, error) {
	return s.CreateWithContext(context.Background(), params)
}

func (s *AudiencesSvcImpl) CreateWithContext(ctx context.Context, params *CreateAudienceRequest) (*Audience, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, "audiences", params)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(Audience)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *AudiencesSvcImpl) List() ([]AudienceListItem, error) {
	return s.ListWithContext(context.Background())
}

func (s *AudiencesSvcImpl) ListWithContext(ctx context.Context) ([]AudienceListItem, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, "audiences", nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	var resp struct {
		Data []AudienceListItem `json:"data"`
	}
	if _, err := s.client.Perform(req, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (s *AudiencesSvcImpl) Remove(audienceId string) (*GenericResponse, error) {
	return s.RemoveWithContext(context.Background(), audienceId)
}

func (s *AudiencesSvcImpl) RemoveWithContext(ctx context.Context, audienceId string) (*GenericResponse, error) {
	req, err := s.client.NewRequest(ctx, http.MethodDelete, "audiences/"+audienceId, nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(GenericResponse)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *AudiencesSvcImpl) CreateContact(audienceId string, params *CreateContactRequest) (*Contact, error) {
	return s.CreateContactWithContext(context.Background(), audienceId, params)
}

func (s *AudiencesSvcImpl) CreateContactWithContext(ctx context.Context, audienceId string, params *CreateContactRequest) (*Contact, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, "audiences/"+audienceId+"/contacts", params)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(Contact)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *AudiencesSvcImpl) BatchCreateContacts(audienceId string, contacts []*CreateContactRequest) (*BatchCreateContactsResponse, error) {
	return s.BatchCreateContactsWithContext(context.Background(), audienceId, contacts)
}

func (s *AudiencesSvcImpl) BatchCreateContactsWithContext(ctx context.Context, audienceId string, contacts []*CreateContactRequest) (*BatchCreateContactsResponse, error) {
	body := map[string]any{"contacts": contacts}
	req, err := s.client.NewRequest(ctx, http.MethodPost, "audiences/"+audienceId+"/contacts/batch", body)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(BatchCreateContactsResponse)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *AudiencesSvcImpl) ListContacts(audienceId string, options *ListContactsOptions) (*ListContactsResponse, error) {
	return s.ListContactsWithContext(context.Background(), audienceId, options)
}

func (s *AudiencesSvcImpl) ListContactsWithContext(ctx context.Context, audienceId string, options *ListContactsOptions) (*ListContactsResponse, error) {
	q := map[string]string{}
	if options != nil {
		if options.Limit > 0 {
			q["limit"] = itoa(options.Limit)
		}
		q["cursor"] = options.Cursor
		q["search"] = options.Search
		if options.Subscribed != nil {
			if *options.Subscribed {
				q["subscribed"] = "true"
			} else {
				q["subscribed"] = "false"
			}
		}
	}
	req, err := s.client.NewRequest(ctx, http.MethodGet, "audiences/"+audienceId+"/contacts"+queryString(q), nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(ListContactsResponse)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *AudiencesSvcImpl) GetContact(audienceId, contactId string) (*Contact, error) {
	return s.GetContactWithContext(context.Background(), audienceId, contactId)
}

func (s *AudiencesSvcImpl) GetContactWithContext(ctx context.Context, audienceId, contactId string) (*Contact, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, "audiences/"+audienceId+"/contacts/"+contactId, nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(Contact)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *AudiencesSvcImpl) UpdateContact(audienceId, contactId string, params *UpdateContactRequest) (*Contact, error) {
	return s.UpdateContactWithContext(context.Background(), audienceId, contactId, params)
}

func (s *AudiencesSvcImpl) UpdateContactWithContext(ctx context.Context, audienceId, contactId string, params *UpdateContactRequest) (*Contact, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPatch, "audiences/"+audienceId+"/contacts/"+contactId, params)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(Contact)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *AudiencesSvcImpl) RemoveContact(audienceId, contactId string) (*GenericResponse, error) {
	return s.RemoveContactWithContext(context.Background(), audienceId, contactId)
}

func (s *AudiencesSvcImpl) RemoveContactWithContext(ctx context.Context, audienceId, contactId string) (*GenericResponse, error) {
	req, err := s.client.NewRequest(ctx, http.MethodDelete, "audiences/"+audienceId+"/contacts/"+contactId, nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(GenericResponse)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
