package eusend

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestClient(handler http.HandlerFunc) (*Client, *httptest.Server) {
	srv := httptest.NewServer(handler)
	c := NewClient("eu_test_key")
	// Point the client at the test server (must end in "/" for BaseURL.Parse).
	c.BaseURL, _ = c.BaseURL.Parse(srv.URL + "/")
	return c, srv
}

func TestSendEncodesSnakeCaseAndIdempotencyKey(t *testing.T) {
	var gotBody map[string]any
	var gotAuth, gotIdem, gotUA string

	c, srv := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotIdem = r.Header.Get("Idempotency-Key")
		gotUA = r.Header.Get("User-Agent")
		if !strings.HasSuffix(r.URL.Path, "/emails") {
			t.Errorf("path = %q", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"id":"abc123"}`))
	})
	defer srv.Close()

	res, err := c.Emails.SendWithOptions(context.Background(), &SendEmailRequest{
		From:       "you@x.com",
		To:         []string{"a@b.com"},
		Subject:    "Hi",
		Html:       "<p>Hi</p>",
		ReplyTo:    []string{"r@x.com"},
		TrackOpens: Bool(false),
		Attachments: []*Attachment{
			{Filename: "a.txt", Content: []byte("hello"), ContentType: "text/plain"},
		},
	}, &SendEmailOptions{IdempotencyKey: "key-1"})
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if res.Id != "abc123" {
		t.Fatalf("id = %q", res.Id)
	}
	if gotAuth != "Bearer eu_test_key" {
		t.Fatalf("auth = %q", gotAuth)
	}
	if gotIdem != "key-1" {
		t.Fatalf("idem = %q", gotIdem)
	}
	if !strings.HasPrefix(gotUA, "eusend-go/") {
		t.Fatalf("user-agent = %q", gotUA)
	}
	if _, ok := gotBody["reply_to"]; !ok {
		t.Fatalf("expected reply_to in body, got %v", gotBody)
	}
	if gotBody["track_opens"] != false {
		t.Fatalf("track_opens = %v", gotBody["track_opens"])
	}
	att := gotBody["attachments"].([]any)[0].(map[string]any)
	if att["content"] != "aGVsbG8=" { // []byte -> base64
		t.Fatalf("attachment content = %v", att["content"])
	}
	if att["content_type"] != "text/plain" {
		t.Fatalf("content_type = %v", att["content_type"])
	}
}

func TestGetDecodesCamelCaseResponse(t *testing.T) {
	c, srv := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":"e1","from":"a@b.com","to":["x@y.com"],"testMode":true,"createdAt":"t","status":"delivered","events":[{"id":"ev1","type":"sent","createdAt":"t2"}]}`))
	})
	defer srv.Close()

	em, err := c.Emails.Get("e1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if em.From != "a@b.com" || !em.TestMode || em.Status != "delivered" {
		t.Fatalf("decoded wrong: %+v", em)
	}
	if len(em.Events) != 1 || em.Events[0].Type != "sent" {
		t.Fatalf("events wrong: %+v", em.Events)
	}
}

func TestBatchStripsAttachmentsAndScheduling(t *testing.T) {
	var got []map[string]any
	c, srv := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &got)
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"data":[{"id":"1"},{"error":"unverified domain","code":"DOMAIN_NOT_VERIFIED"}]}`))
	})
	defer srv.Close()

	in := &SendEmailRequest{From: "a", To: []string{"b"}, ScheduledAt: "in 1 hour",
		Attachments: []*Attachment{{Filename: "x", Content: []byte("y")}}}
	res, err := c.Batch.Send([]*SendEmailRequest{in, in})
	if err != nil {
		t.Fatalf("batch: %v", err)
	}
	if _, ok := got[0]["scheduled_at"]; ok {
		t.Fatalf("scheduled_at should be stripped: %v", got[0])
	}
	if _, ok := got[0]["attachments"]; ok {
		t.Fatalf("attachments should be stripped: %v", got[0])
	}
	if in.ScheduledAt != "in 1 hour" {
		t.Fatalf("caller struct mutated")
	}
	if res.Data[0].Id != "1" || res.Data[1].Code != CodeDomainNotVerified {
		t.Fatalf("results wrong: %+v", res.Data)
	}
}

func TestErrorResponseBecomesTypedError(t *testing.T) {
	c, srv := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(429)
		_, _ = w.Write([]byte(`{"error":"Monthly send limit exceeded","code":"MONTHLY_LIMIT_EXCEEDED"}`))
	})
	defer srv.Close()

	_, err := c.Emails.Send(&SendEmailRequest{From: "a", To: []string{"b"}})
	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.Code != CodeMonthlyLimitExceeded || apiErr.StatusCode != 429 {
		t.Fatalf("wrong error: %+v", apiErr)
	}
}

func TestListSendsQueryAndDecodesCursor(t *testing.T) {
	c, srv := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("status") != "delivered" || r.URL.Query().Get("limit") != "20" {
			t.Errorf("bad query: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"1","subject":"s","createdAt":"t"}],"next_cursor":"nc"}`))
	})
	defer srv.Close()

	page, err := c.Emails.List(&ListEmailsOptions{Limit: 20, Status: "delivered"})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if page.NextCursor != "nc" || len(page.Data) != 1 {
		t.Fatalf("page wrong: %+v", page)
	}
}

func TestListContactsUnderAudience(t *testing.T) {
	c, srv := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/audiences/aud1/contacts") {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("subscribed") != "true" {
			t.Errorf("query = %q", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"c1","email":"a@b.com","status":"subscribed"}],"next_cursor":""}`))
	})
	defer srv.Close()

	page, err := c.Audiences.ListContacts("aud1", &ListContactsOptions{Subscribed: Bool(true)})
	if err != nil {
		t.Fatalf("list contacts: %v", err)
	}
	if len(page.Data) != 1 || page.Data[0].Email != "a@b.com" {
		t.Fatalf("contacts wrong: %+v", page.Data)
	}
}
