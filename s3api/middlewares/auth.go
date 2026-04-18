package middlewares

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// AuthType represents the type of authentication used in the request.
type AuthType string

const (
	AuthTypeV4        AuthType = "AWS4-HMAC-SHA256"
	AuthTypeV2        AuthType = "AWS"
	AuthTypeAnonymous AuthType = "Anonymous"
	AuthTypeUnknown   AuthType = "Unknown"

	// authContextKey is the key used to store auth metadata in context.
	authContextKey = "s3_auth_type"
)

// AuthMetadata holds parsed authentication information from the request.
type AuthMetadata struct {
	Type      AuthType
	AccessKey string
	Region    string
	Service   string
}

// DetectAuthType inspects the Authorization header and returns the auth type.
func DetectAuthType(authHeader string) AuthType {
	switch {
	case strings.HasPrefix(authHeader, string(AuthTypeV4)):
		return AuthTypeV4
	case strings.HasPrefix(authHeader, string(AuthTypeV2)):
		return AuthTypeV2
	case authHeader == "":
		return AuthTypeAnonymous
	default:
		return AuthTypeUnknown
	}
}

// ParseV4AuthHeader parses an AWS Signature Version 4 Authorization header
// and extracts the access key, region, and service.
func ParseV4AuthHeader(authHeader string) AuthMetadata {
	meta := AuthMetadata{Type: AuthTypeV4}

	// Format: AWS4-HMAC-SHA256 Credential=<access>/<date>/<region>/<service>/aws4_request, ...
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) < 2 {
		return meta
	}

	for _, field := range strings.Split(parts[1], ",") {
		field = strings.TrimSpace(field)
		if strings.HasPrefix(field, "Credential=") {
			cred := strings.TrimPrefix(field, "Credential=")
			segments := strings.Split(cred, "/")
			if len(segments) >= 4 {
				meta.AccessKey = segments[0]
				meta.Region = segments[2]
				meta.Service = segments[3]
			}
			break
		}
	}

	return meta
}

// AuthParser is a middleware that parses and attaches authentication metadata
// to the request context. It does not validate credentials — only identifies
// the auth type and extracts available metadata for downstream handlers.
func AuthParser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get(fiber.HeaderAuthorization)
		authType := DetectAuthType(authHeader)

		var meta AuthMetadata
		switch authType {
		case AuthTypeV4:
			meta = ParseV4AuthHeader(authHeader)
		case AuthTypeV2:
			meta = AuthMetadata{Type: AuthTypeV2}
			// V2 format: AWS <access>:<signature>
			parts := strings.SplitN(strings.TrimPrefix(authHeader, "AWS "), ":", 2)
			if len(parts) == 2 {
				meta.AccessKey = parts[0]
			}
		default:
			meta = AuthMetadata{Type: authType}
		}

		c.Locals(authContextKey, meta)
		return c.Next()
	}
}

// GetAuthMetadata retrieves the AuthMetadata stored in the fiber context.
// Returns a zero-value AuthMetadata if none is set.
func GetAuthMetadata(c *fiber.Ctx) AuthMetadata {
	if meta, ok := c.Locals(authContextKey).(AuthMetadata); ok {
		return meta
	}
	return AuthMetadata{Type: AuthTypeUnknown}
}

// RequireAuth is a middleware that rejects ano
