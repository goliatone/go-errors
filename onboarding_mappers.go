package errors

import "net/http"

// MapOnboardingErrors normalizes invite, reset, verification, and feature gate errors.
func MapOnboardingErrors(err error) *Error {
	msg := normalizeErrorMessage(err)
	switch {
	case containsAny(msg, "invite expired", "invitation expired") || containsAll(msg, "invite", "expired"):
		return New(err.Error(), CategoryBadInput).
			WithCode(http.StatusGone).
			WithTextCode(TextCodeInviteExpired)
	case containsAny(msg, "invite used", "invitation used", "invite already used") || containsAll(msg, "invite", "used"):
		return New(err.Error(), CategoryConflict).
			WithCode(http.StatusConflict).
			WithTextCode(TextCodeInviteUsed)
	case containsAny(msg, "token already used"):
		return New(err.Error(), CategoryConflict).
			WithCode(http.StatusConflict).
			WithTextCode(TextCodeTokenAlreadyUsed)
	case containsAny(msg, "reset not allowed", "password reset not allowed"):
		return New(err.Error(), CategoryAuthz).
			WithCode(http.StatusForbidden).
			WithTextCode(TextCodeResetNotAllowed)
	case containsAny(msg, "reset rate limit", "password reset rate limit", "password reset rate limited", "password reset is rate limited"):
		return New(err.Error(), CategoryRateLimit).
			WithCode(http.StatusTooManyRequests).
			WithTextCode(TextCodeResetRateLimit)
	case containsAny(msg, "account locked", "account lockout", "locked out"):
		return New(err.Error(), CategoryAuth).
			WithCode(http.StatusForbidden).
			WithTextCode(TextCodeAccountLocked)
	case containsAny(msg, "verification required", "verification needed", "email not verified", "email verification required"):
		return New(err.Error(), CategoryAuth).
			WithCode(http.StatusForbidden).
			WithTextCode(TextCodeVerificationRequired)
	case containsAny(msg, "verification expired", "verification token expired"):
		return New(err.Error(), CategoryAuth).
			WithCode(http.StatusForbidden).
			WithTextCode(TextCodeVerificationExpired)
	case containsAny(msg, "feature disabled", "signup disabled", "registration disabled", "self registration disabled"):
		return New(err.Error(), CategoryAuthz).
			WithCode(http.StatusForbidden).
			WithTextCode(TextCodeFeatureDisabled)
	}
	return nil
}
