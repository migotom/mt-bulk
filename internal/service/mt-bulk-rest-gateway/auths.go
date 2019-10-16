package mtbulkrestapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/dgrijalva/jwt-go"
	"github.com/migotom/mt-bulk/internal/entities"
)

// TokenClaims represents JWT authentication claims.
type TokenClaims struct {
	AllowedHostPatterns []string
	jwt.StandardClaims
}

// Authenticate represents authentication rules.
type Authenticate struct {
	Key                 string   `toml:"key" yaml:"key"`
	AllowedHostPatterns []string `toml:"allowed_host_patterns" yaml:"allowed_host_patterns"`
}

// AuthenticateToken creates authentication token.
func (mtbulk *MTbulkRESTGateway) AuthenticateToken(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var auth struct {
			Key string `json:"key"`
		}

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&auth); err != nil {
			http.Error(w, "Bad request", 400)
		}

		var claims TokenClaims
		for _, authrole := range mtbulk.Config.Authenticate {
			if authrole.Key == auth.Key {
				claims.AllowedHostPatterns = authrole.AllowedHostPatterns

				token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(mtbulk.Config.TokenSecret))
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}

				response := struct {
					Token string `json:"token"`
				}{
					Token: token,
				}
				if err := json.NewEncoder(w).Encode(&response); err != nil {
					http.Error(w, err.Error(), 500)
					return
				}
				return
			}
		}

		http.Error(w, "Not authenticated", 401)
	}
}

// AuthorizeMiddleware authenticates request, looks into auth token, decodes it and verifies.
func (mtbulk *MTbulkRESTGateway) AuthorizeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorizationToken := r.Header.Get("Authorization")

		var claims TokenClaims
		token, err := jwt.ParseWithClaims(authorizationToken, &claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("Not authenticated")
			}
			return []byte(mtbulk.Config.TokenSecret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Not authenticated", 401)
			return
		}

		ctx := context.WithValue(r.Context(), "claims", claims)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// AuthorizeRequest authorizes request by checking token claims and comapring them with requested job.
func (mtbulk *MTbulkRESTGateway) AuthorizeRequest(r *http.Request, job *entities.Job) error {
	claims := r.Context().Value("claims")
	if claims == nil {
		return errors.New("claims fetch error")
	}

	tokenClaims, ok := claims.(TokenClaims)
	if !ok {
		return errors.New("invalid claims")
	}

	for _, allowedHostPattern := range tokenClaims.AllowedHostPatterns {
		re, err := regexp.Compile(allowedHostPattern)
		if err != nil {
			return errors.New("invalid host pattern")
		}
		if re.MatchString(job.Host.String()) {
			return nil
		}
	}

	return fmt.Errorf("not authenticated to host %s", job.Host)
}
