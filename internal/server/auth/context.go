package auth

const (
	// AuthTypeAPIToken identifies API token authentication.
	AuthTypeAPIToken string = "api_token"
)

// AuthContext carries request-scoped authentication inputs for a given listener.
type AuthContext struct {
	// AuthKind selects the validator in Auth.Validators (e.g. "api_token").
	AuthKind string
	// ListenerType is the inbound listener type (e.g. "http").
	ListenerType string
	// APIToken is the caller-provided API token for AuthTypeAPIToken.
	APIToken string
}

// NewAPITokenContext constructs an AuthContext for API token authentication.
func NewAPITokenContext(listenerType string, apiToken string) AuthContext {
	return AuthContext{
		AuthKind:     AuthTypeAPIToken,
		ListenerType: listenerType,
		APIToken:     apiToken,
	}
}
