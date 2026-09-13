package eusend

import (
	"context"
	"net/http"
)

// DnsRecord is a DNS entry to publish for a domain.
type DnsRecord struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Value       string `json:"value"`
	Priority    int    `json:"priority,omitempty"`    // MX records only
	Purpose     string `json:"purpose,omitempty"`     // authentication | policy | alignment
	Description string `json:"description,omitempty"`
}

// CreateDomainRequest is the request object for Domains.Create.
type CreateDomainRequest struct {
	Name string `json:"name"`
}

// CreateDomainResponse is returned by Domains.Create and carries the DNS records to add.
type CreateDomainResponse struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	// Records is every record to publish, in presentation order. Prefer it over the
	// individual fields below — it is the only place the optional Return-Path
	// alignment records appear.
	Records []DnsRecord `json:"records"`
	Dkim    DnsRecord   `json:"dkim"`
	Dmarc   DnsRecord   `json:"dmarc"`
}

// DomainListItem is a row from Domains.List.
type DomainListItem struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

// DomainVerification reports whether a verification chain is polling DNS for the
// domain right now, and when that chain started.
type DomainVerification struct {
	Running   bool   `json:"running"`
	StartedAt string `json:"startedAt"`
}

// DomainDnsProvider is the DNS host serving a domain's zone, recognised from its
// nameservers. Guide is a path on eusend.dev, empty where no guide exists.
type DomainDnsProvider struct {
	Id    string `json:"id"`
	Label string `json:"label"`
	Guide string `json:"guide"`
}

// DomainDiagnostic is what the last unmatched DNS check found, when it found a
// mistake rather than an absence. Nil while nothing is wrong beyond the records
// not having propagated yet.
//
// Code is the mistake: "doubled_domain" (the record sits under the domain twice,
// because the control panel appends it to whatever you type), "truncated_key" (the
// value was cut at the 255-character limit for a single DNS string instead of being
// split into two), "foreign_key" (a DKIM key we did not issue is published at the
// selector), "quoted_value", "multiple_records", "cname_at_selector". New codes may
// be added, so treat an unknown one as generic.
type DomainDiagnostic struct {
	Code string `json:"code"`
	// FoundAt is the name the record was actually found at, for "doubled_domain".
	FoundAt string `json:"foundAt"`
	// PublishedChars and ExpectedChars describe a "truncated_key".
	PublishedChars int `json:"publishedChars"`
	ExpectedChars  int `json:"expectedChars"`
	// Target is where the CNAME points, for "cname_at_selector".
	Target   string             `json:"target"`
	Provider *DomainDnsProvider `json:"provider"`
}

// Domain is the response from Domains.Get.
type Domain struct {
	Id            string             `json:"id"`
	Name          string             `json:"name"`
	DkimPublicKey string             `json:"dkimPublicKey"`
	DkimSelector  string             `json:"dkimSelector"`
	Status        string             `json:"status"`
	CreatedAt     string             `json:"createdAt"`
	VerifiedAt    string             `json:"verifiedAt"`
	Verification  DomainVerification `json:"verification"`
	Diagnostic    *DomainDiagnostic  `json:"diagnostic"`
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
