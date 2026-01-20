package errors_test

import (
	stdErrors "errors"
	"testing"

	"github.com/goliatone/go-errors"
)

func TestMapAuthErrors_TextCodes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name         string
		message      string
		wantTextCode string
		wantCode     int
		wantCategory errors.Category
	}{
		{
			name:         "unauthorized",
			message:      "Unauthorized request",
			wantTextCode: "UNAUTHORIZED",
			wantCode:     401,
			wantCategory: errors.CategoryAuth,
		},
		{
			name:         "token expired",
			message:      "token is expired",
			wantTextCode: errors.TextCodeTokenExpired,
			wantCode:     401,
			wantCategory: errors.CategoryAuth,
		},
		{
			name:         "authentication token expired",
			message:      "authentication token expired",
			wantTextCode: errors.TextCodeTokenExpired,
			wantCode:     401,
			wantCategory: errors.CategoryAuth,
		},
		{
			name:         "token malformed",
			message:      "token is malformed",
			wantTextCode: errors.TextCodeTokenMalformed,
			wantCode:     400,
			wantCategory: errors.CategoryAuth,
		},
		{
			name:         "too many attempts",
			message:      "too many login attempts",
			wantTextCode: errors.TextCodeTooManyAttempts,
			wantCode:     429,
			wantCategory: errors.CategoryRateLimit,
		},
		{
			name:         "account suspended",
			message:      "user account is suspended",
			wantTextCode: errors.TextCodeAccountSuspended,
			wantCode:     403,
			wantCategory: errors.CategoryAuth,
		},
		{
			name:         "account disabled",
			message:      "user account is disabled",
			wantTextCode: errors.TextCodeAccountDisabled,
			wantCode:     403,
			wantCategory: errors.CategoryAuth,
		},
		{
			name:         "account archived",
			message:      "user account is archived",
			wantTextCode: errors.TextCodeAccountArchived,
			wantCode:     403,
			wantCategory: errors.CategoryAuth,
		},
		{
			name:         "account pending",
			message:      "user account is pending",
			wantTextCode: errors.TextCodeAccountPending,
			wantCode:     403,
			wantCategory: errors.CategoryAuth,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mapped := errors.MapAuthErrors(stdErrors.New(tc.message))
			if mapped == nil {
				t.Fatalf("expected mapped error for %q", tc.message)
			}
			if mapped.TextCode != tc.wantTextCode {
				t.Fatalf("expected text code %q, got %q", tc.wantTextCode, mapped.TextCode)
			}
			if mapped.Code != tc.wantCode {
				t.Fatalf("expected code %d, got %d", tc.wantCode, mapped.Code)
			}
			if mapped.Category != tc.wantCategory {
				t.Fatalf("expected category %q, got %q", tc.wantCategory, mapped.Category)
			}
		})
	}
}

func TestMapOnboardingErrors_TextCodes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name         string
		message      string
		wantTextCode string
		wantCode     int
		wantCategory errors.Category
	}{
		{
			name:         "invite expired",
			message:      "invite token expired",
			wantTextCode: errors.TextCodeInviteExpired,
			wantCode:     410,
			wantCategory: errors.CategoryBadInput,
		},
		{
			name:         "invite used",
			message:      "invite already used",
			wantTextCode: errors.TextCodeInviteUsed,
			wantCode:     409,
			wantCategory: errors.CategoryConflict,
		},
		{
			name:         "token already used",
			message:      "token already used",
			wantTextCode: errors.TextCodeTokenAlreadyUsed,
			wantCode:     409,
			wantCategory: errors.CategoryConflict,
		},
		{
			name:         "reset rate limit",
			message:      "password reset rate limited",
			wantTextCode: errors.TextCodeResetRateLimit,
			wantCode:     429,
			wantCategory: errors.CategoryRateLimit,
		},
		{
			name:         "account locked",
			message:      "account locked",
			wantTextCode: errors.TextCodeAccountLocked,
			wantCode:     403,
			wantCategory: errors.CategoryAuth,
		},
		{
			name:         "verification required",
			message:      "verification required",
			wantTextCode: errors.TextCodeVerificationRequired,
			wantCode:     403,
			wantCategory: errors.CategoryAuth,
		},
		{
			name:         "verification expired",
			message:      "verification token expired",
			wantTextCode: errors.TextCodeVerificationExpired,
			wantCode:     403,
			wantCategory: errors.CategoryAuth,
		},
		{
			name:         "feature disabled",
			message:      "self registration disabled",
			wantTextCode: errors.TextCodeFeatureDisabled,
			wantCode:     403,
			wantCategory: errors.CategoryAuthz,
		},
		{
			name:         "reset not allowed",
			message:      "password reset not allowed",
			wantTextCode: errors.TextCodeResetNotAllowed,
			wantCode:     403,
			wantCategory: errors.CategoryAuthz,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mapped := errors.MapOnboardingErrors(stdErrors.New(tc.message))
			if mapped == nil {
				t.Fatalf("expected mapped error for %q", tc.message)
			}
			if mapped.TextCode != tc.wantTextCode {
				t.Fatalf("expected text code %q, got %q", tc.wantTextCode, mapped.TextCode)
			}
			if mapped.Code != tc.wantCode {
				t.Fatalf("expected code %d, got %d", tc.wantCode, mapped.Code)
			}
			if mapped.Category != tc.wantCategory {
				t.Fatalf("expected category %q, got %q", tc.wantCategory, mapped.Category)
			}
		})
	}
}

func TestDefaultErrorMappers_PrefersOnboardingErrors(t *testing.T) {
	t.Parallel()
	err := stdErrors.New("invite token expired")
	mapped := errors.MapToError(err, errors.DefaultErrorMappers())
	if mapped.TextCode != errors.TextCodeInviteExpired {
		t.Fatalf("expected text code %q, got %q", errors.TextCodeInviteExpired, mapped.TextCode)
	}
}

type statusError struct {
	code    int
	message string
}

func (e statusError) Error() string   { return e.message }
func (e statusError) StatusCode() int { return e.code }

func TestDefaultErrorMappers_PrefersOnboardingErrorsWithStatusCode(t *testing.T) {
	t.Parallel()
	err := statusError{code: 410, message: "invite token expired"}
	mapped := errors.MapToError(err, errors.DefaultErrorMappers())
	if mapped.TextCode != errors.TextCodeInviteExpired {
		t.Fatalf("expected text code %q, got %q", errors.TextCodeInviteExpired, mapped.TextCode)
	}
	if mapped.Code != 410 {
		t.Fatalf("expected code %d, got %d", 410, mapped.Code)
	}
}
