package mailbox

import (
	"fmt"
	"net/url"
)

// Graph API base URLs.
const (
	graphBaseURL = "https://graph.microsoft.com/v1.0"

	// endpoint path templates — use with fmt.Sprintf(endpointXxx, id)
	endpointMessages    = graphBaseURL + "/me/messages"
	endpointMessage     = graphBaseURL + "/me/messages/%s"
	endpointAttachments = graphBaseURL + "/me/messages/%s/attachments"
)

// Azure AD token endpoint template — interpolate with TenantID.
const tokenEndpoint = "https://login.microsoftonline.com/%s/oauth2/v2.0/token"

// OAuth2 ROPC form-field keys.
const (
	formGrantType    = "grant_type"
	formClientID     = "client_id"
	formClientSecret = "client_secret"
	formUsername     = "username"
	formPassword     = "password"
	formScope        = "scope"
)

// Graph API permission scopes.
const (
	scopeMailRead = "https://graph.microsoft.com/Mail.Read"
)

// OData query-parameter keys used across the Graph API.
const (
	odataTop     = "$top"
	odataOrderBy = "$orderby"
	odataFilter  = "$filter"
)

// messageURL builds the /me/messages list URL with the given OData params.
// filter may be empty, in which case the $filter param is omitted.
func messageURL(top int, filter string) string {
	params := url.Values{}
	params.Set(odataTop, fmt.Sprintf("%d", top))
	params.Set(odataOrderBy, "receivedDateTime desc")
	if filter != "" {
		params.Set(odataFilter, filter)
	}
	return endpointMessages + "?" + params.Encode()
}

// messageByIDURL returns the URL for a single message resource.
func messageByIDURL(messageID string) string {
	return fmt.Sprintf(endpointMessage, messageID)
}

// attachmentsURL returns the URL for a message's attachments collection.
func attachmentsURL(messageID string) string {
	return fmt.Sprintf(endpointAttachments, messageID)
}

// tokenEndpointURL returns the fully-qualified Azure AD token URL for tenantID.
func tokenEndpointURL(tenantID string) string {
	return fmt.Sprintf(tokenEndpoint, tenantID)
}

// tokenFormData builds the url.Values body for an ROPC token request.
// clientSecret is optional — omitted when empty.
func tokenFormData(cfg *Config) url.Values {
	data := url.Values{}
	data.Set(formGrantType, "password")
	data.Set(formClientID, cfg.ClientID)
	data.Set(formUsername, cfg.Username)
	data.Set(formPassword, cfg.Password)
	data.Set(formScope, scopeMailRead)
	if cfg.ClientSecret != "" {
		data.Set(formClientSecret, cfg.ClientSecret)
	}
	return data
}
