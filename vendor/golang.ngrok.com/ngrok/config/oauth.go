package config

import "golang.ngrok.com/ngrok/internal/pb"

type OAuthOption func(cfg *oauthOptions)

// oauthOptions configuration
type oauthOptions struct {
	// The OAuth provider to use
	Provider string
	// Email addresses of users to authorize.
	AllowEmails []string
	// Email domains of users to authorize.
	AllowDomains []string
	// OAuth scopes to request from the provider.
	Scopes []string
}

// Construct a new OAuth provider with the given name.
func oauthProvider(name string) *oauthOptions {
	return &oauthOptions{
		Provider: name,
	}
}

// Append email addresses to the list of allowed emails.
func WithAllowOAuthEmail(addr ...string) OAuthOption {
	return func(cfg *oauthOptions) {
		cfg.AllowEmails = append(cfg.AllowEmails, addr...)
	}
}

// Append email domains to the list of allowed domains.
func WithAllowOAuthDomain(domain ...string) OAuthOption {
	return func(cfg *oauthOptions) {
		cfg.AllowDomains = append(cfg.AllowDomains, domain...)
	}
}

// Append scopes to the list of scopes to request.
func WithOAuthScope(scope ...string) OAuthOption {
	return func(cfg *oauthOptions) {
		cfg.Scopes = append(cfg.Scopes, scope...)
	}
}

func (oauth *oauthOptions) toProtoConfig() *pb.MiddlewareConfiguration_OAuth {
	if oauth == nil {
		return nil
	}

	return &pb.MiddlewareConfiguration_OAuth{
		Provider:     string(oauth.Provider),
		AllowEmails:  oauth.AllowEmails,
		AllowDomains: oauth.AllowDomains,
		Scopes:       oauth.Scopes,
	}
}

// WithOAuth configures this edge with the the given OAuth provider.
// Overwrites any previously-set OAuth configuration.
func WithOAuth(provider string, opts ...OAuthOption) HTTPEndpointOption {
	return httpOptionFunc(func(cfg *httpOptions) {
		oauth := oauthProvider(provider)
		for _, opt := range opts {
			opt(oauth)
		}
		cfg.OAuth = oauth
	})
}
