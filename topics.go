package eusend

import (
	"context"
	"net/http"
)

// Topic is a named category of email a contact can subscribe to and leave independently of
// the audience they sit on. Scope a Broadcast to one with CreateBroadcastRequest.TopicId
// and it reaches only the contacts who want that kind of mail.
type Topic struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// DefaultSubscription is "opt_in" (everyone receives it until they leave) or "opt_out"
	// (nobody receives it until they join). Fixed at creation.
	DefaultSubscription string `json:"defaultSubscription"`
	// Visibility is "public" or "private" -- whether the topic is listed on the hosted
	// unsubscribe page to a contact who is not already opted in.
	Visibility string `json:"visibility"`
	// SubscriberCount counts explicit opt-ins only. For an opt_out topic that is the whole
	// reachable audience; for an opt_in topic it is not, because everyone who never chose
	// is subscribed too. Returned by Create and List, not by Get.
	SubscriberCount int    `json:"subscriberCount"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

// CreateTopicRequest is the request object for Topics.Create.
type CreateTopicRequest struct {
	// Name is at most 50 characters and unique within the organization. It is what the
	// recipient reads on the unsubscribe page.
	Name string `json:"name"`
	// DefaultSubscription is "opt_in" or "opt_out", and CANNOT be changed afterwards:
	// flipping an opt_out topic to opt_in would start mailing every contact who never
	// asked for it, and nothing stored says which of them would have agreed.
	DefaultSubscription string `json:"default_subscription"`
	// Description is at most 200 characters, shown under the name on the unsubscribe page.
	Description *string `json:"description,omitempty"`
	// Visibility is "public" or "private" (the default when empty).
	Visibility string `json:"visibility,omitempty"`
}

// UpdateTopicRequest changes a topic's presentation.
//
// DefaultSubscription is absent deliberately -- see the note on CreateTopicRequest.
type UpdateTopicRequest struct {
	Name *string `json:"name,omitempty"`
	// A nil Description is omitted; use eusend.String("") to clear it.
	Description *string `json:"description,omitempty"`
	Visibility  *string `json:"visibility,omitempty"`
}

// DeleteTopicResponse reports the delete.
type DeleteTopicResponse struct {
	Id      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

// TopicSubscriptionResponse reports a contact's recorded preference.
type TopicSubscriptionResponse struct {
	TopicId    string `json:"topicId"`
	ContactId  string `json:"contactId"`
	Subscribed bool   `json:"subscribed"`
}

// ClearTopicSubscriptionResponse reports a cleared preference.
type ClearTopicSubscriptionResponse struct {
	TopicId   string `json:"topicId"`
	ContactId string `json:"contactId"`
	Deleted   bool   `json:"deleted"`
}

// TopicsSvc is the /topics API.
type TopicsSvc interface {
	Create(params *CreateTopicRequest) (*Topic, error)
	CreateWithContext(ctx context.Context, params *CreateTopicRequest) (*Topic, error)
	List() ([]Topic, error)
	ListWithContext(ctx context.Context) ([]Topic, error)
	Get(topicId string) (*Topic, error)
	GetWithContext(ctx context.Context, topicId string) (*Topic, error)
	Update(topicId string, params *UpdateTopicRequest) (*Topic, error)
	UpdateWithContext(ctx context.Context, topicId string, params *UpdateTopicRequest) (*Topic, error)
	Remove(topicId string) (*DeleteTopicResponse, error)
	RemoveWithContext(ctx context.Context, topicId string) (*DeleteTopicResponse, error)
	Subscribe(topicId string, contactId string, subscribed bool) (*TopicSubscriptionResponse, error)
	SubscribeWithContext(ctx context.Context, topicId string, contactId string, subscribed bool) (*TopicSubscriptionResponse, error)
	ClearSubscription(topicId string, contactId string) (*ClearTopicSubscriptionResponse, error)
	ClearSubscriptionWithContext(ctx context.Context, topicId string, contactId string) (*ClearTopicSubscriptionResponse, error)
}

type TopicsSvcImpl struct{ client *Client }

func (s *TopicsSvcImpl) Create(params *CreateTopicRequest) (*Topic, error) {
	return s.CreateWithContext(context.Background(), params)
}

func (s *TopicsSvcImpl) CreateWithContext(ctx context.Context, params *CreateTopicRequest) (*Topic, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, "topics", params)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(Topic)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *TopicsSvcImpl) List() ([]Topic, error) {
	return s.ListWithContext(context.Background())
}

// ListWithContext returns every topic the organization has defined, ordered by name.
// Unpaginated -- the list is capped at 100 entries and every caller wants all of them.
func (s *TopicsSvcImpl) ListWithContext(ctx context.Context) ([]Topic, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, "topics", nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	var resp struct {
		Data []Topic `json:"data"`
	}
	if _, err := s.client.Perform(req, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (s *TopicsSvcImpl) Get(topicId string) (*Topic, error) {
	return s.GetWithContext(context.Background(), topicId)
}

func (s *TopicsSvcImpl) GetWithContext(ctx context.Context, topicId string) (*Topic, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, "topics/"+topicId, nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(Topic)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *TopicsSvcImpl) Update(topicId string, params *UpdateTopicRequest) (*Topic, error) {
	return s.UpdateWithContext(context.Background(), topicId, params)
}

func (s *TopicsSvcImpl) UpdateWithContext(ctx context.Context, topicId string, params *UpdateTopicRequest) (*Topic, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPatch, "topics/"+topicId, params)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(Topic)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *TopicsSvcImpl) Remove(topicId string) (*DeleteTopicResponse, error) {
	return s.RemoveWithContext(context.Background(), topicId)
}

// RemoveWithContext deletes the topic and every stored preference for it. A draft or
// scheduled broadcast scoped to the topic widens back to its whole audience; broadcasts
// already sent keep their recipient records.
func (s *TopicsSvcImpl) RemoveWithContext(ctx context.Context, topicId string) (*DeleteTopicResponse, error) {
	req, err := s.client.NewRequest(ctx, http.MethodDelete, "topics/"+topicId, nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(DeleteTopicResponse)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *TopicsSvcImpl) Subscribe(topicId string, contactId string, subscribed bool) (*TopicSubscriptionResponse, error) {
	return s.SubscribeWithContext(context.Background(), topicId, contactId, subscribed)
}

// SubscribeWithContext records an explicit preference for one contact.
func (s *TopicsSvcImpl) SubscribeWithContext(ctx context.Context, topicId string, contactId string, subscribed bool) (*TopicSubscriptionResponse, error) {
	body := struct {
		Subscribed bool `json:"subscribed"`
	}{Subscribed: subscribed}
	req, err := s.client.NewRequest(ctx, http.MethodPut, "topics/"+topicId+"/contacts/"+contactId, body)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(TopicSubscriptionResponse)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *TopicsSvcImpl) ClearSubscription(topicId string, contactId string) (*ClearTopicSubscriptionResponse, error) {
	return s.ClearSubscriptionWithContext(context.Background(), topicId, contactId)
}

// ClearSubscriptionWithContext forgets the contact's explicit choice so they fall back to
// the topic's default. Different from Subscribe(..., false), which records "no thanks": on
// an opt_in topic, clearing means the contact starts receiving it again.
func (s *TopicsSvcImpl) ClearSubscriptionWithContext(ctx context.Context, topicId string, contactId string) (*ClearTopicSubscriptionResponse, error) {
	req, err := s.client.NewRequest(ctx, http.MethodDelete, "topics/"+topicId+"/contacts/"+contactId, nil)
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}
	resp := new(ClearTopicSubscriptionResponse)
	if _, err := s.client.Perform(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
