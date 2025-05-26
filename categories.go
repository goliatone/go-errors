package errors

import "strings"

// Category represents a high level error category
type Category string

func (c Category) String() string { return string(c) }

func (c Category) Extend(s string) Category { return Category(string(c) + "_" + strings.ToLower(s)) }

const (
	CategoryValidation       Category = "validation"
	CategoryAuth             Category = "authentication"
	CategoryAuthz            Category = "authorization"
	CategoryOperation        Category = "operation"
	CategoryNotFound         Category = "not_found"
	CategoryConflict         Category = "conflict"
	CategoryRateLimit        Category = "rate_limit"
	CategoryBadInput         Category = "bad_input"
	CategoryInternal         Category = "internal"
	CategoryExternal         Category = "external"
	CategoryMiddleware       Category = "middleware"
	CategoryRouting          Category = "routing"
	CategoryHandler          Category = "handler"
	CategoryMethodNotAllowed Category = "method_not_allowed"
	CategoryCommand          Category = "command"
)

func IsCategory(err error, category Category) bool {
	var e *Error
	if As(err, &e) {
		return e.Category == category
	}
	return false
}

func IsValidation(err error) bool {
	return IsCategory(err, CategoryValidation)
}

func IsAuth(err error) bool {
	return IsCategory(err, CategoryAuth)
}

func IsNotFound(err error) bool {
	return IsCategory(err, CategoryNotFound)
}

func IsInternal(err error) bool {
	return IsCategory(err, CategoryInternal)
}

func IsCommand(err error) bool {
	return IsCategory(err, CategoryCommand)
}
