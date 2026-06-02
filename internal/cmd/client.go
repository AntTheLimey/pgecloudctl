package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/AntTheLimey/pgecloudctl/internal/api"
	"github.com/AntTheLimey/pgecloudctl/internal/auth"
	"github.com/AntTheLimey/pgecloudctl/internal/config"
)

// Exit codes used by CLI commands.
const (
	ExitOK        = 0
	ExitGeneral   = 1
	ExitAuth      = 2
	ExitTimeout   = 3
	ExitNotFound  = 4
)

// ExitError carries a message and an exit code for use with command error handling.
// Exported so main.go can extract the code for os.Exit.
type ExitError struct {
	msg  string
	code int
}

func (e *ExitError) Error() string { return e.msg }
func (e *ExitError) Code() int    { return e.code }

// checkResponse inspects the HTTP status code and returns a descriptive
// *ExitError for any non-2xx status, or nil on success.
func checkResponse(status int, body string) error {
	if status >= 200 && status < 300 {
		return nil
	}
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return &ExitError{
			msg:  fmt.Sprintf("authentication error (%d): %s", status, body),
			code: ExitAuth,
		}
	case http.StatusNotFound:
		return &ExitError{
			msg:  fmt.Sprintf("resource not found (%d): %s", status, body),
			code: ExitNotFound,
		}
	case http.StatusRequestTimeout, http.StatusGatewayTimeout:
		return &ExitError{
			msg:  fmt.Sprintf("request timed out (%d): %s", status, body),
			code: ExitTimeout,
		}
	default:
		return &ExitError{
			msg:  fmt.Sprintf("API error (%d): %s", status, body),
			code: ExitGeneral,
		}
	}
}

// newAPIClient resolves credentials, obtains a token, and returns a configured
// *api.ClientWithResponses ready to make authenticated requests.
func newAPIClient() (*api.ClientWithResponses, error) {
	store := config.DefaultStore()
	a := &auth.Auth{Store: store, APIURL: flagAPIURL}

	creds, _, err := a.ResolveCredentials(flagClientID, flagSecret)
	if err != nil {
		return nil, &ExitError{msg: err.Error(), code: ExitAuth}
	}

	token, err := a.GetToken(creds.ClientID, creds.ClientSecret)
	if err != nil {
		return nil, &ExitError{
			msg:  fmt.Sprintf("authentication failed: %v", err),
			code: ExitAuth,
		}
	}

	bearerEditor := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	}

	client, err := api.NewClientWithResponses(
		flagAPIURL,
		api.WithRequestEditorFn(bearerEditor),
	)
	if err != nil {
		return nil, fmt.Errorf("create API client: %w", err)
	}
	return client, nil
}
