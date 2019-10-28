package auth

import (
	"context"
	"github.com/coreos/go-oidc"
	"github.com/lyft/flyteadmin/pkg/auth/config"
	"github.com/lyft/flyteadmin/pkg/auth/interfaces"
	"github.com/lyft/flytestdlib/errors"
	"github.com/lyft/flytestdlib/logger"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
	"strings"
)

type Context struct {
	oauth2        *oauth2.Config
	claims        config.Claims
	cookieManager interfaces.CookieHandler
	oidcProvider  *oidc.Provider
	options       config.OAuthOptions
}

func (c Context) OAuth2Config() *oauth2.Config {
	return c.oauth2
}

func (c Context) Claims() config.Claims {
	return c.claims
}

func (c Context) OidcProvider() *oidc.Provider {
	return c.oidcProvider
}

func (c Context) CookieManager() interfaces.CookieHandler {
	return c.cookieManager
}

func (c Context) Options() config.OAuthOptions {
	return c.options
}

const (
	ErrAuthContext errors.ErrorCode = "AUTH_CONTEXT_SETUP_FAILED"
	ErrConfigFileRead errors.ErrorCode = "CONFIG_OPTION_FILE_READ_FAILED"
)

func NewAuthenticationContext(ctx context.Context, options config.OAuthOptions) (Context, error) {
	oauth2Config, err := GetOauth2Config(options)
	if err != nil {
		return Context{}, errors.Wrapf(ErrAuthContext, err, "Error creating OAuth2 library configuration")
	}

	hashKeyBytes, err := ioutil.ReadFile(options.CookieHashKeyFile)
	if err != nil {
		return Context{}, errors.Wrapf(ErrConfigFileRead, err, "Could not read hash key file")
	}
	blockKeyBytes, err := ioutil.ReadFile(options.CookieBlockKeyFile)
	if err != nil {
		return Context{}, errors.Wrapf(ErrConfigFileRead, err, "Could not read block key file")
	}

	cookieManager, err := NewCookieManager(ctx, string(hashKeyBytes), string(blockKeyBytes))
	if err != nil {
		logger.Errorf(ctx, "Error creating cookie manager %s", err)
		return Context{}, errors.Wrapf(ErrAuthContext, err, "Error creating cookie manager")
	}

	oidcCtx := oidc.ClientContext(ctx, &http.Client{})
	provider, err := oidc.NewProvider(oidcCtx, options.Claims.Issuer)
	if err != nil {
		return Context{}, errors.Wrapf(ErrAuthContext, err, "Error creating oidc provider")
	}

	return Context{
		oauth2:        &oauth2Config,
		claims:        options.Claims,
		cookieManager: cookieManager,
		oidcProvider:  provider,
		options:       options,
	}, nil
}

// This creates a oauth2 library config object, with values from the Flyte Admin config
func GetOauth2Config(options config.OAuthOptions) (oauth2.Config, error) {
	secretBytes, err := ioutil.ReadFile(options.ClientSecretFile)
	if err != nil {
		return oauth2.Config{}, err
	}
	secret := strings.TrimSuffix(string(secretBytes), "\n")
	return oauth2.Config{
		RedirectURL:  options.CallbackUrl,
		ClientID:     options.ClientId,
		ClientSecret: secret,
		// Offline access needs to be specified in order to return a refresh token in the exchange.
		// TODO: Second parameter is IDP specific - move to config. Also handle case where a refresh token is not allowed
		Scopes: []string{OidcScope, OfflineAccessType, ProfileScope},
		Endpoint: oauth2.Endpoint{
			AuthURL:  options.AuthorizeUrl,
			TokenURL: options.TokenUrl,
		},
	}, nil
}