package auth

// GoogleLoginRequest is the request for GoogleLogin usecase
type GoogleLoginRequest struct {
	RedirectTo string
}

// GoogleLoginResult is the result for GoogleLogin usecase
type GoogleLoginResult struct {
	RedirectURL string
}

// GoogleCallbackRequest is the request for GoogleCallback usecase
type GoogleCallbackRequest struct {
	Code      string
	State     string
	IPAddress string
	UserAgent string
}

// GoogleCallbackResult is the result for GoogleCallback usecase
type GoogleCallbackResult struct {
	SessionID  string
	RedirectTo string
}
