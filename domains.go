package eusend

import (
	"context"
	"net/http"
)

// DnsRecord is a DNS entry to publish for a domain.
type DnsRecord struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

// CreateDomainRequest is the request object for Domains.Create.
type CreateDomainRequest struct {
	Name string `json:"name"`
}

// CreateDomainResponse is returned by Domains.Create and carries the DNS records to add.
type CreateDomainResponse struct {
	Id    string    `json:"id"`
	Name  string    `json:"name"`
	Dkim  DnsRecord `json:"dkim"`
	Spf   DnsRecord `json:"spf"`
	Dmarc DnsRecord `json:"dmarc"`
}

// DomainListItem is a row from Domains.List.
type DomainListItem struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

// Domain is the response from Domains.Get.
type Domain struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	DkimPublicKey string `json:"dkimPublicKey"`
	DkimSelector  string `json:"dkimSelector"`
	Status        string `json:"status"`
	CreatedAt     string `json:"createdAt"`
	VerifiedAt    string `json:"verifiedAt"`
}

// GenericResponse is a simple {"message": "..."} acknowledgement.
type GenericResponse struct {
	Message string `json:"message"`
}

// DomainsSvc is the /domains API.
type DomainsSvc interface {
	Create(params *CreateDomainRequest) (*CreateDomainResponse, error)
	CreateWithContext(ctx context.Context, params *CreateDomainRequest) (*CreateDomainResponse, error)
	List() ([]DomainListItem, error)
	ListWithContext(ctx context.Context) ([]DomainListItem, error)
	Get(domainId string) (*Domain, error)
	GetWithContext(ctx context.Context, domainId string) (*Domain, error)
	Verify(domainId string) (*GenericResponse, error)
	VerifyWithContext(ctx context.Context, domainId string) (*GenericResponse, error)
	Remove(domainId string) (*GenericResponse, error)
	RemoveWithContext(ctx context.Context, domainId string) (*GenericResponse, error)
}

type DomainsSvcImpl struct{ client *Client }

func (s *DomainsSvcImpl) Create(params *CreateDomainRequest) (*CreateDomainResponse, error) {
	return s.CreateWithContext(context.Background(), params)
}

func (s *DomainsSvcImpl) CreateWithContext(ctx context.Context, params *CreateDomainRequest) (*CreateDomainResponse, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, "domains", params)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(CreateDomainResponse)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *DomainsSvcImpl) List() ([]DomainListItem, error) {
	return s.ListWithContext(context.Background())
}

func (s *DomainsSvcImpl) ListWithContext(ctx context.Context) ([]DomainListItem, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, "domains", nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	var resp []DomainListItem
	if _, err := s.client.Perform(req, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *DomainsSvcImpl) Get(domainId string) (*Domain, error) {
	return s.GetWithContext(context.Background(), domainId)
}

func (s *DomainsSvcImpl) GetWithContext(ctx context.Context, domainId string) (*Domain, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, "domains/"+domainId, nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(Domain)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *DomainsSvcImpl) Verify(domainId string) (*GenericResponse, error) {
	return s.VerifyWithContext(context.Background(), domainId)
}

func (s *DomainsSvcImpl) VerifyWithContext(ctx context.Context, domainId string) (*GenericResponse, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, "domains/"+domainId+"/verify", nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(GenericResponse)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *DomainsSvcImpl) Remove(domainId string) (*GenericResponse, error) {
	return s.RemoveWithContext(context.Background(), domainId)
}

func (s *DomainsSvcImpl) RemoveWithContext(ctx context.Context, domainId string) (*GenericResponse, error) {
	req, err := s.client.NewRequest(ctx, http.MethodDelete, "domains/"+domainId, nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(GenericResponse)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
