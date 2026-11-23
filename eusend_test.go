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

func TestBatchSendWithOptionsSetsIdempotencyKey(t *testing.T) {
	var key string
	c, srv := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		key = r.Header.Get("Idempotency-Key")
		_, _ = w.Write([]byte(`{"data":[{"id":"1"}]}`))
	})
	defer srv.Close()

	in := []*SendEmailRequest{{From: "a", To: []string{"b"}}}
	if _, err := c.Batch.SendWithOptions(context.Background(), in, &BatchSendOptions{IdempotencyKey: "digest-1"}); err != nil {
		t.Fatalf("batch: %v", err)
	}
	if key != "digest-1" {
		t.Fatalf("idempotency key not sent: %q", key)
	}
	if _, err := c.Batch.Send(in); err != nil {
		t.Fatalf("batch: %v", err)
	}
	if key != "" {
		t.Fatalf("plain Send should not set a key: %q", key)
	}
}

func TestAudienceGetAndUpdate(t *testing.T) {
	var method, path string
	var body map[string]any
	c, srv := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"id":"aud_1","name":"Product updates","createdAt":"c","updatedAt":"u","contactCount":7}`))
	})
	defer srv.Close()

	got, err := c.Audiences.Get("aud_1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if method != http.MethodGet || path != "/audiences/aud_1" || got.ContactCount != 7 || got.UpdatedAt != "u" {
		t.Fatalf("get wrong: %s %s %+v", method, path, got)
	}

	updated, err := c.Audiences.Update("aud_1", &UpdateAudienceRequest{Name: "Product updates"})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if method != http.MethodPatch || path != "/audiences/aud_1" || body["name"] != "Product updates" || updated.Name != "Product updates" {
		t.Fatalf("update wrong: %s %s %v %+v", method, path, body, updated)
	}
}

func TestBroadcastReplyToAndDomainRemoveCount(t *testing.T) {
	var body map[string]any
	c, srv := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			_, _ = w.Write([]byte(`{"message":"Domain deleted","canceled_scheduled_emails":2}`))
			return
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"id":"bc_1","replyTo":"support@example.com"}`))
	})
	defer srv.Close()

	if _, err := c.Broadcasts.Create(&CreateBroadcastRequest{Name: "n", ReplyTo: "support@example.com"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	if body["reply_to"] != "support@example.com" {
		t.Fatalf("reply_to not sent: %v", body)
	}
	if _, err := c.Broadcasts.Update("bc_1", &UpdateBroadcastRequest{ReplyTo: String("help@example.com")}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if body["reply_to"] != "help@example.com" {
		t.Fatalf("reply_to not sent on update: %v", body)
	}

	removed, err := c.Domains.Remove("dom_1")
	if err != nil {
		t.Fatalf("remove: %v", err)
	}
	if removed.CanceledScheduledEmails != 2 {
		t.Fatalf("count not decoded: %+v", removed)
	}
}

func TestReputationGetDecodesVerdict(t *testing.T) {
	c, srv := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.HasSuffix(r.URL.Path, "/reputation") {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"state":"warning","measured":true,` +
			`"bounce":{"rate":0.0333,"judged_rate":0.0151,"counted":4,"not_counted":2,"recipients":120,"window":"7d","limited_at":0.02,"paused_at":0.04},` +
			`"complaint":{"rate":0,"complaints":0,"limited_at":0.0004,"paused_at":0.0008},` +
			`"next":{"state":"limited","bounces":3}}`))
	})
	defer srv.Close()

	r, err := c.Reputation.Get()
	if err != nil {
		t.Fatal(err)
	}
	if r.State != ReputationWarning || r.Bounce.NotCounted != 2 || *r.Bounce.Window != "7d" {
		t.Errorf("decoded = %+v", r)
	}
	if r.Next == nil || r.Next.State != ReputationLimited || *r.Next.Bounces != 3 {
		t.Errorf("next = %+v", r.Next)
	}
}

func TestReputationPausedHasNoNextStep(t *testing.T) {
	c, srv := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"state":"paused","measured":true,"bounce":{"window":null},"complaint":{},"next":null}`))
	})
	defer srv.Close()

	r, err := c.Reputation.Get()
	if err != nil {
		t.Fatal(err)
	}
	if r.Next != nil || r.Bounce.Window != nil {
		t.Errorf("paused = %+v", r)
	}
}

func TestInboundAddressGetAndAttachment(t *testing.T) {
	c, srv := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/inbound/address":
			w.Write([]byte(`{"host":"abc123.eusend.app","address":"anything@abc123.eusend.app","status":"active","active":true}`))
		case "/inbound/ie_1":
			w.Write([]byte(`{"id":"ie_1","from":"Alice <a@x.test>","to":["s@abc123.eusend.app"],"cc":[],"replyTo":[],"receivedFor":["s@abc123.eusend.app"],"subject":"Hi","replyToEmailId":"em_9","text":"hello","html":null,"headers":[],"attachments":[{"index":0,"filename":"a.pdf","contentType":"application/pdf","size":3,"contentId":null,"disposition":"attachment"}],"authentication":{"dmarc":{"result":"pass"},"dkim":[{"result":"pass"}]},"size":120,"createdAt":"2026-11-23T07:00:00Z"}`))
		case "/inbound/ie_1/attachments/0":
			w.Header().Set("Content-Type", "application/pdf")
			w.Write([]byte("PDF"))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	})
	defer srv.Close()

	addr, err := c.Inbound.Address()
	if err != nil || addr.Host != "abc123.eusend.app" || !addr.Active {
		t.Fatalf("address: %+v %v", addr, err)
	}
	email, err := c.Inbound.Get("ie_1")
	if err != nil {
		t.Fatal(err)
	}
	if *email.ReplyToEmailId != "em_9" || email.Authentication.Dmarc.Result != "pass" || email.Attachments[0].Size != 3 {
		t.Fatalf("get: %+v", email)
	}
	body, err := c.Inbound.Attachment("ie_1", 0)
	if err != nil || string(body) != "PDF" {
		t.Fatalf("attachment: %q %v", body, err)
	}
}

func TestEnableInboundSendsSubdomain(t *testing.T) {
	c, srv := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/domains/d_1/inbound" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["subdomain"] != "in" {
			t.Fatalf("body %v", body)
		}
		w.Write([]byte(`{"host":"in.x.test","record":{"type":"MX","name":"in.x.test","value":"inbound.eusend.dev","priority":10},"mxPublished":false}`))
	})
	defer srv.Close()

	res, err := c.Domains.EnableInbound("d_1", &EnableInboundRequest{Subdomain: "in"})
	if err != nil || res.Record.Value != "inbound.eusend.dev" || res.MxPublished {
		t.Fatalf("enable: %+v %v", res, err)
	}
}
