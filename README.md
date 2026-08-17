# eusend-go

Official Go SDK for the [Eusend](https://eusend.dev) API — the EU-native transactional email platform.

Its shape mirrors [`resend-go`](https://github.com/resend/resend-go), so migrating from Resend is largely a `resend` → `eusend` rename.

- **`NewClient` + service methods** — `client.Emails.Send(...)`, with `WithContext` variants.
- **Zero dependencies** — standard library only.
- **Concurrency-safe** — create one `*Client` and share it.

```bash
go get github.com/eusend-dev/eusend-go
```

Requires Go 1.21+.

---

## Getting started

```go
package main

import (
	"fmt"
	"log"

	eusend "github.com/eusend-dev/eusend-go"
)

func main() {
	client := eusend.NewClient("eu_live_...") // or NewClient("") to read EUSEND_API_KEY

	sent, err := client.Emails.Send(&eusend.SendEmailRequest{
		// From accepts a bare email or a display-name form: "Acme <you@yourdomain.com>"
		From:    "Acme <you@yourdomain.com>",
		To:      []string{"user@example.com"},
		Subject: "Hello",
		Html:    "<p>Hello world</p>",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(sent.Id) // 9a8b7c6d-... (UUID)
}
```

Every method has a `WithContext` variant that takes a `context.Context` as its
first argument (e.g. `client.Emails.SendWithContext(ctx, params)`). The
context-free forms use `context.Background()`.

Optional pointer fields (`*bool`, `*string`) have helpers: `eusend.Bool(true)`, `eusend.String("x")`.

---

## Emails

### Send

`From` and `To` are required; provide at least one of `Html`, `Text`, or `TemplateId`.

| Field                        | Type                | Notes                                                                                                      |
| ---------------------------- | ------------------- | ---------------------------------------------------------------------------------------------------------- |
| `From`                       | `string`            | Verified domain; bare or display-name form.                                                                |
| `To` `Cc` `Bcc` `ReplyTo`    | `[]string`          | Max 50 each.                                                                                               |
| `Subject`                    | `string`            |                                                                                                            |
| `Html` / `Text`              | `string`            |                                                                                                            |
| `TemplateId`                 | `string`            | Saved template.                                                                                            |
| `Variables`                  | `map[string]any`    | Template substitutions (HTML-escaped).                                                                     |
| `Headers`                    | `map[string]string` | No line breaks in names or values.                                                                         |
| `TrackOpens` / `TrackClicks` | `*bool`             | Nil uses your organization default (Settings → General → Email tracking); `eusend.Bool(false)` to disable. |
| `Attachments`                | `[]*Attachment`     | Up to 20, 10 MB combined.                                                                                  |
| `ScheduledAt`                | `string`            | Future send, ≤ 30 days.                                                                                    |

### Attachments

Provide `Content` (raw bytes, base64-encoded on the wire) **or** `Path` (a public
URL fetched at send time). Set `ContentId` for an inline `<img src="cid:...">`.

```go
pdf, _ := os.ReadFile("invoice.pdf")

client.Emails.Send(&eusend.SendEmailRequest{
	From:    "you@yourdomain.com",
	To:      []string{"user@example.com"},
	Subject: "Your invoice",
	Html:    "<p>Attached.</p>",
	Attachments: []*eusend.Attachment{
		{Filename: "invoice.pdf", Content: pdf, ContentType: "application/pdf"},
	},
})
```

### Idempotent sends

```go
client.Emails.SendWithOptions(ctx, params, &eusend.SendEmailOptions{
	IdempotencyKey: "receipt-" + orderID,
})
```

Retrying with the same key never sends a duplicate and returns the original ID.

### Scheduled sends

`ScheduledAt` accepts an ISO 8601 string or natural language (`"in 1 hour"`,
`"tomorrow at 9am"`), parsed server-side in UTC.

```go
sent, _ := client.Emails.Send(&eusend.SendEmailRequest{
	From: "you@yourdomain.com", To: []string{"user@example.com"},
	Subject: "Reminder", Html: "<p>Soon.</p>",
	ScheduledAt: "in 1 hour",
})

client.Emails.Update(&eusend.UpdateEmailRequest{Id: sent.Id, ScheduledAt: "in 2 hours"})
client.Emails.Cancel(sent.Id)
```

### Batch

Up to 100 emails in one request. Attachments and scheduling are stripped (not
supported on the batch endpoint). Results map positionally to the input: queued
items carry `Id`, rejected items carry `Error` and `Code`.

```go
res, _ := client.Batch.Send([]*eusend.SendEmailRequest{
	{From: "you@yourdomain.com", To: []string{"alice@example.com"}, Subject: "Hi", Html: "<p>Hi</p>"},
	{From: "you@yourdomain.com", To: []string{"bob@example.com"},   Subject: "Hi", Html: "<p>Hi</p>"},
})
for _, r := range res.Data {
	if r.Id != "" {
		fmt.Println("queued", r.Id)
	} else {
		fmt.Printf("failed: %s (%s)\n", r.Error, r.Code)
	}
}
```

### Retrieve & list

```go
email, _ := client.Emails.Get("9a8b7c6d-...")
fmt.Println(email.Status, email.Events[0].Type)

page, _ := client.Emails.List(&eusend.ListEmailsOptions{Limit: 20, Status: "delivered"})
for _, e := range page.Data {
	fmt.Println(e.Id, e.Subject)
}
if page.NextCursor != "" {
	page, _ = client.Emails.List(&eusend.ListEmailsOptions{Cursor: page.NextCursor})
}
```

Statuses: `queued` `scheduled` `sending` `sent` `delivered` `bounced` `complained` `suppressed` `failed`.

---

## Domains

```go
created, _ := client.Domains.Create(&eusend.CreateDomainRequest{Name: "yourdomain.com"})
fmt.Println(created.Dkim.Name, created.Dkim.Value) // DNS records to add
fmt.Println(created.Spf, created.Dmarc)

client.Domains.Verify(created.Id) // after publishing the DNS records
client.Domains.List()
client.Domains.Get(created.Id)
client.Domains.Remove(created.Id)
```

---

## API keys

```go
key, _ := client.ApiKeys.Create(&eusend.CreateApiKeyRequest{Name: "Production"})
fmt.Println(key.Key) // eu_live_... — returned only once

client.ApiKeys.Create(&eusend.CreateApiKeyRequest{Name: "Sandbox", TestMode: true}) // eu_test_... key
client.ApiKeys.List()                                                               // prefixes only
client.ApiKeys.Remove(key.Id)
```

Emails sent with a test key are accepted and tracked but never delivered.

`Permission` defaults to `PermissionFullAccess` — every resource. `PermissionSendingAccess`
limits the key to sending email (plus rescheduling and canceling a scheduled send); every
other endpoint, including reading email logs, returns `403 FORBIDDEN`. Such a key can also
be pinned to one sending domain with `DomainId`, which is rejected on a full-access key.

```go
client.ApiKeys.Create(&eusend.CreateApiKeyRequest{
	Name:       "Billing service",
	Permission: eusend.PermissionSendingAccess,
	DomainId:   domainId, // omit for any verified domain
})
```

Deleting a domain revokes every key restricted to it.

---

## Audiences & contacts

Contact operations are grouped under `Audiences` (they live under a specific audience).

```go
audience, _ := client.Audiences.Create(&eusend.CreateAudienceRequest{Name: "Newsletter"})

client.Audiences.CreateContact(audience.Id, &eusend.CreateContactRequest{
	Email: "user@example.com", FirstName: "Jane",
})

// Bulk upsert (up to 1,000)
client.Audiences.BatchCreateContacts(audience.Id, []*eusend.CreateContactRequest{
	{Email: "alice@example.com", FirstName: "Alice"},
	{Email: "bob@example.com", FirstName: "Bob"},
})

page, _ := client.Audiences.ListContacts(audience.Id, &eusend.ListContactsOptions{
	Subscribed: eusend.Bool(true), Search: "gmail.com",
})
contact := page.Data[0]

client.Audiences.UpdateContact(audience.Id, contact.Id, &eusend.UpdateContactRequest{
	Unsubscribed: eusend.Bool(true),
})
client.Audiences.GetContact(audience.Id, contact.Id)
client.Audiences.RemoveContact(audience.Id, contact.Id)

client.Audiences.List()
client.Audiences.Remove(audience.Id)
```

---

## Suppressions

Addresses the account will not send to. Hard bounces and spam complaints are added
automatically; these methods cover the ones you manage yourself. A send to a suppressed
address is skipped and recorded with status `suppressed`; if every recipient is
suppressed the send fails with `ALL_SUPPRESSED`.

Test-mode keys can read the list but not modify it.

```go
// Everything suppressed for a hard bounce, or at one domain
page, _ := client.Suppressions.List(&eusend.ListSuppressionsOptions{
	Reason: eusend.SuppressionReasonBounce, Limit: 50,
})
client.Suppressions.List(&eusend.ListSuppressionsOptions{Email: "@acme.com"})

// Suppress an address. Already suppressed? The existing entry comes back unchanged —
// a manual add never rewrites a real bounce or complaint.
client.Suppressions.Create(&eusend.CreateSuppressionRequest{Email: "opted-out@example.com"})

// Import up to 1,000 at a time — do this before your first send when migrating, so
// addresses that already bounced elsewhere don't get a fresh attempt from a new IP.
res, _ := client.Suppressions.Import([]*eusend.SuppressionImportItem{
	{Email: "one@example.com"},
	{Email: "two@example.com", Reason: eusend.SuppressionReasonComplaint},
})
fmt.Println(res.Count, res.AlreadySuppressed, res.Duplicates)

// Un-suppress by entry id or by address
client.Suppressions.Remove("invalid@example.com")

// The whole list as CSV
csv, _ := client.Suppressions.Export()

_ = page
```

---

## Templates

`{{variable}}` placeholders are substituted at send time; values are HTML-escaped.

```go
tpl, _ := client.Templates.Create(&eusend.CreateTemplateRequest{
	Name:    "Welcome email",
	Subject: "Welcome, {{name}}!",
	Html:    "<h1>Hi {{name}}</h1><p>Welcome to {{product}}.</p>",
})

client.Emails.Send(&eusend.SendEmailRequest{
	From: "you@yourdomain.com", To: []string{"user@example.com"},
	TemplateId: tpl.Id,
	Variables:  map[string]any{"name": "Jane", "product": "Acme"},
})

client.Templates.List()
client.Templates.Get(tpl.Id)
client.Templates.Update(tpl.Id, &eusend.UpdateTemplateRequest{Subject: eusend.String("New subject")})
client.Templates.Remove(tpl.Id)
```

---

## Webhooks

```go
hook, _ := client.Webhooks.Create(&eusend.CreateWebhookRequest{
	Url:    "https://yourapp.com/webhooks/eusend",
	Events: []string{"email.delivered", "email.bounced", "email.complained"}, // or []string{"*"}
})
fmt.Println(hook.Secret) // signing secret — returned only once

client.Webhooks.List()
client.Webhooks.Get(hook.Id) // includes recent deliveries
client.Webhooks.Update(hook.Id, &eusend.UpdateWebhookRequest{Events: []string{"email.bounced"}})
client.Webhooks.Remove(hook.Id)
```

Events: `email.sent` `email.delivered` `email.bounced` `email.complained`
`email.opened` `email.clicked`. The endpoint must be a public `http(s)` URL
returning `2xx` directly (redirects count as failures).

### Verifying signatures

Every delivery is signed with HMAC-SHA256 over `{webhook-id}.{webhook-timestamp}.{body}`:

```go
import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

func verify(r *http.Request, body []byte, secret string) bool {
	signed := r.Header.Get("webhook-id") + "." + r.Header.Get("webhook-timestamp") + "." + string(body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signed))
	expected := "v1," + base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(r.Header.Get("webhook-signature")), []byte(expected))
}
```

---

## Broadcasts

Send one email to every contact in an audience. `{{first_name}}`, `{{last_name}}`,
`{{full_name}}`, and `{{email}}` are available per recipient, and RFC 8058
one-click unsubscribe headers are added automatically.

```go
bc, _ := client.Broadcasts.Create(&eusend.CreateBroadcastRequest{
	Name:       "May newsletter",
	AudienceId: audience.Id,
	From:       "Sivert <hello@yourdomain.com>",
	Subject:    "May update",
	Html:       "<p>Hi {{first_name}}, your monthly update is here...</p>",
})

client.Broadcasts.Send(bc.Id, nil)                                          // send now
client.Broadcasts.Send(bc.Id, &eusend.SendBroadcastRequest{ScheduledAt: "2026-06-01T09:00:00Z"}) // or schedule
client.Broadcasts.Cancel(bc.Id)

client.Broadcasts.List()
client.Broadcasts.Get(bc.Id) // includes delivery stats
client.Broadcasts.Update(bc.Id, &eusend.UpdateBroadcastRequest{Subject: eusend.String("Updated")})
client.Broadcasts.Remove(bc.Id)
```

Calling `Send` on a paused broadcast resumes it from where it stopped.

---

## Error handling

Every method returns `(result, error)`. Any non-2xx response, and any network
failure, is an `*eusend.Error`:

```go
sent, err := client.Emails.Send(params)
if err != nil {
	var apiErr *eusend.Error
	if errors.As(err, &apiErr) {
		fmt.Println(apiErr.Code)       // "MONTHLY_LIMIT_EXCEEDED"
		fmt.Println(apiErr.Message)    // "Monthly send limit exceeded"
		fmt.Println(apiErr.StatusCode) // 429 (0 for a network failure)

		if apiErr.Code == eusend.CodeMonthlyLimitExceeded {
			// back off and retry later
		}
	}
	return
}
```

On a `429`, `apiErr.RetryAfter`, `RateLimitReset`, and `RateLimitRemaining` are
populated from the response headers.

| Code constant              | Wire value               | Status              |
| -------------------------- | ------------------------ | ------------------- |
| `CodeUnauthorized`         | `UNAUTHORIZED`           | 401                 |
| `CodeForbidden`            | `FORBIDDEN`              | 403                 |
| `CodeNotFound`             | `NOT_FOUND`              | 404                 |
| `CodeValidationError`      | `VALIDATION_ERROR`       | 400                 |
| `CodeBadRequest`           | `BAD_REQUEST`            | 400                 |
| `CodeConflict`             | `CONFLICT`               | 409                 |
| `CodeRateLimited`          | `RATE_LIMITED`           | 429                 |
| `CodeMonthlyLimitExceeded` | `MONTHLY_LIMIT_EXCEEDED` | 429                 |
| `CodeDailyLimitExceeded`   | `DAILY_LIMIT_EXCEEDED`   | 429                 |
| `CodePlanLimitExceeded`    | `PLAN_LIMIT_EXCEEDED`    | 403                 |
| `CodeDomainNotVerified`    | `DOMAIN_NOT_VERIFIED`    | 403                 |
| `CodeSendingSuspended`     | `SENDING_SUSPENDED`      | 403                 |
| `CodeListSendHeld`         | `LIST_SEND_HELD`         | 403                 |
| `CodeBroadcastHeld`        | `BROADCAST_HELD`         | 403                 |
| `CodeAllSuppressed`        | `ALL_SUPPRESSED`         | 422                 |
| `CodeServicePaused`        | `SERVICE_PAUSED`         | 503                 |
| `CodeInternalError`        | `INTERNAL_ERROR`         | 500                 |
| `CodeApplicationError`     | `application_error`      | — (network failure) |

---

## Configuration

```go
// Custom http.Client (timeouts, proxies, ...):
client := eusend.NewCustomClient(&http.Client{Timeout: 60 * time.Second}, "eu_live_...")

// Override the base URL (e.g. for testing) after construction:
client.BaseURL, _ = url.Parse("https://api.eusend.dev/")
```
