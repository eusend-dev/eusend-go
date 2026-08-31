package eusend

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Error is returned when the API responds with a non-2xx status, or when the
// request never reaches the server. Inspect it with errors.As:
//
//	var apiErr *eusend.Error
//	if errors.As(err, &apiErr) && apiErr.Code == eusend.CodeMonthlyLimitExceeded {
//	    // back off and retry later
//	}
type Error struct {
	// Message is a human-readable description (the API's `error` field).
	Message string `json:"error"`
	// Code is a stable machine-readable code (the API's `code` field); see the
	// Code* constants. Branch on this rather than on Message.
	Code string `json:"code"`
	// StatusCode is the HTTP status, or 0 for a network-level failure
	// (Code == CodeApplicationError).
	StatusCode int `json:"-"`

	// RateLimit* are populated from response headers on a 429.
	RateLimitReset     string `json:"-"`
	RateLimitRemaining string `json:"-"`
	RetryAfter         string `json:"-"`
}

func (e *Error) Error() string {
	if e.StatusCode == 0 {
		return fmt.Sprintf("eusend: %s (%s)", e.Message, e.Code)
	}
	return fmt.Sprintf("eusend: %s (%s, status %d)", e.Message, e.Code, e.StatusCode)
}

// handleError decodes a non-2xx response body ({"error","code"}) into *Error.
func handleError(resp *http.Response) error {
	apiErr := &Error{Message: resp.Status, Code: CodeInternalError, StatusCode: resp.StatusCode}

	if strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") {
		body, _ := io.ReadAll(resp.Body)
		var parsed Error
		if json.Unmarshal(body, &parsed) == nil {
			if parsed.Message != "" {
				apiErr.Message = parsed.Message
			}
			if parsed.Code != "" {
				apiErr.Code = parsed.Code
			}
		}
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		apiErr.RateLimitReset = resp.Header.Get("ratelimit-reset")
		apiErr.RateLimitRemaining = resp.Header.Get("ratelimit-remaining")
		apiErr.RetryAfter = resp.Header.Get("retry-after")
	}

	return apiErr
}

// Error codes returned by the API. CodeApplicationError is SDK-only and signals
// that the request never reached the server (network failure, DNS, timeout).
// CodeRecipientDomainUndeliverable means a recipient's domain publishes no mail
// exchanger, so the message could never be delivered — usually a typo.
const (
	CodeUnauthorized                 = "UNAUTHORIZED"
	CodeForbidden                    = "FORBIDDEN"
	CodeNotFound                     = "NOT_FOUND"
	CodeValidationError              = "VALIDATION_ERROR"
	CodeBadRequest                   = "BAD_REQUEST"
	CodePayloadTooLarge              = "PAYLOAD_TOO_LARGE"
	CodeConflict                     = "CONFLICT"
	CodeRateLimited                  = "RATE_LIMITED"
	CodeMonthlyLimitExceeded         = "MONTHLY_LIMIT_EXCEEDED"
	CodeDailyLimitExceeded           = "DAILY_LIMIT_EXCEEDED"
	CodePlanLimitExceeded            = "PLAN_LIMIT_EXCEEDED"
	CodeDomainNotVerified            = "DOMAIN_NOT_VERIFIED"
	CodeSendingSuspended             = "SENDING_SUSPENDED"
	CodeAccountRestricted            = "ACCOUNT_RESTRICTED"
	CodeSenderNotPermitted           = "SENDER_NOT_PERMITTED"
	CodeListSendHeld                 = "LIST_SEND_HELD"
	CodeBroadcastHeld                = "BROADCAST_HELD"
	CodeAllSuppressed                = "ALL_SUPPRESSED"
	CodeRecipientDomainUndeliverable = "RECIPIENT_DOMAIN_UNDELIVERABLE"
	CodeAttachmentStorageErr         = "ATTACHMENT_STORAGE_ERROR"
	CodeServicePaused                = "SERVICE_PAUSED"
	CodeInternalError                = "INTERNAL_ERROR"
	CodeApplicationError             = "application_error"
)

// Sentinel errors returned when a request object cannot be constructed.
var (
	ErrFailedToCreateRequest = errors.New("[ERROR]: Failed to create request")
)
