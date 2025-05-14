package errors

// Category represents a high level error category
type Category string

const (
	CategoryValidation       Category = "validation"
	CategoryAuth             Category = "authentication"
	CategoryAuthz            Category = "authorization"
	CategoryNotFound         Category = "not_found"
	CategoryConflict         Category = "conflict"
	CategoryRateLimit        Category = "rate_limit"
	CategoryBadInput         Category = "bad_input"
	CategoryInternal         Category = "internal"
	CategoryExternal         Category = "external"
	CateogryMiddleware       Category = "middleware"
	CategoryRouting          Category = "routing"
	CategoryHandler          Category = "handler"
	CategoryMethodNotAllowed Category = "method_not_allowed"
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
