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
	"github.com/migotom/mt-bulk/internal/service"
)

// MTbulkRESTGateway service.
type MTbulkRESTGateway struct {
	Service *service.Service
	Config
}

// NewMTbulkRestGateway returns new MTbulkRESTGateway service.
func NewMTbulkRestGateway(arguments map[string]interface{}, version string) (MTbulkRESTGateway, error) {
	config, err := configParser(arguments, version)
	if err != nil {
		return MTbulkRESTGateway{}, err
	}

	return MTbulkRESTGateway{
		Config:  config,
		Service: service.NewService(config.Service),
	}, nil
}

// RunWorkers runs service workers and process all provided jobs.
// Returns after process of all jobs.
func (mtbulk *MTbulkRESTGateway) RunWorkers(ctx context.Context) {
	mtbulk.Service.Listen(ctx)
}

func (mtbulk *MTbulkRESTGateway) JobHandler(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var job entities.Job

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&job); err != nil {
			http.Error(w, "Bad request", 400)
			return
		}

		claims := r.Context().Value("claims")
		if claims == nil {
			http.Error(w, "Bad request", 400)
			return
		}

		tokenClaims, ok := claims.(TokenClaims)
		if !ok {
			http.Error(w, "Bad request", 400)
			return
		}

		job.Host.Parse()

		allowedToProcessHost := false
		for _, allowedHostPattern := range tokenClaims.AllowedHostPatterns {
			re, err := regexp.Compile(allowedHostPattern)
			if err != nil {
				http.Error(w, err.Error(), 500)
			}
			if re.MatchString(job.Host.String()) {
				allowedToProcessHost = true
				break
			}
		}
		if !allowedToProcessHost {
			http.Error(w, fmt.Sprintf("Not authenticated to host %s", job.Host), 401)
			return
		}

		resultChan := make(chan entities.Result)
		job.Result = resultChan

		select {
		case <-ctx.Done():
			http.Error(w, "Request cancelled by host", 410)
			return
		case mtbulk.Service.Jobs <- job:
		}

		select {
		case <-ctx.Done():
			http.Error(w, "Request cancelled by host", 410)
			return
		case result := <-resultChan:
			if result.Error != nil {
				w.WriteHeader(http.StatusNotAcceptable)
			}
			if err := json.NewEncoder(w).Encode(&result); err != nil {
				http.Error(w, err.Error(), 500)
			}
		}
	}
}

func (mtbulk *MTbulkRESTGateway) Authenticate(ctx context.Context) http.HandlerFunc {
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
