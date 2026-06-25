package service

import (
	"context"
	"crypto/subtle"
	"errors"
	"slices"

	"github.com/erewhile/iam/internal/ent/db"
	"github.com/erewhile/iam/internal/repository"
	"github.com/erewhile/iam/pkg/hash"
)

var (
	ErrClientNotFound     = errors.New("client not found")
	ErrRedirectURIInvalid = errors.New("redirect_uri is not registered for this client")
	ErrClientSecretWrong  = errors.New("client secret is invalid")
)

type OAuthService struct {
	appRepo repository.ApplicationRepository
}

func NewOAuthService(appRepo repository.ApplicationRepository) *OAuthService {
	return &OAuthService{appRepo: appRepo}
}

func (s *OAuthService) ValidateAuthorize(ctx context.Context, clientID, redirectURI string) (*db.Application, error) {
	app, err := s.appRepo.GetByClientID(ctx, clientID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, ErrClientNotFound
		}
		return nil, err
	}

	if !slices.Contains(app.RedirectUris, redirectURI) {
		return nil, ErrRedirectURIInvalid
	}

	return app, nil
}

func (s *OAuthService) ValidateClient(ctx context.Context, clientID, clientSecret string) (*db.Application, error) {
	app, err := s.appRepo.GetByClientID(ctx, clientID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, ErrClientNotFound
		}
		return nil, err
	}

	secretHash := hash.HashBlake2b256([]byte(clientSecret))
	if subtle.ConstantTimeCompare(app.ClientSecret, secretHash) != 1 {
		return nil, ErrClientSecretWrong
	}

	return app, nil
}
