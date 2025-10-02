package api

import (
	"net/http"
	"strings"

	"firebase.google.com/go/auth"
)

func VerifyToken(c *auth.Client) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")

			accessToken, ok := strings.CutPrefix(h, "Bearer ")
			if !ok || accessToken == "" {
				w.Header().Set("WWW-Authenticate", `Bearer realm="api.bollocks.social" error="invalid_request" error_description="missing parameter: access_token"`)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			token, err := c.VerifyIDToken(r.Context(), accessToken)

			if err != nil {
				// TODO: better inspection of this error - do not send back implemnentation details to the client.
				w.Header().Set("WWW-Authenticate", `Bearer realm="api.bollocks.social" error="invalid_token" error_description="`+err.Error()+`"`)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := ContextWithUserId(r.Context(), token.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
