package service

import (
	"context"
	"errors"

	"github.com/erewhile/iam/internal/dto/req"
	"github.com/erewhile/iam/internal/dto/resp"
	"github.com/erewhile/iam/internal/ent/db"
	"github.com/erewhile/iam/internal/logger"
	"github.com/erewhile/iam/internal/repository"
	"github.com/erewhile/iam/pkg/utils"
)

type ApplicationService struct {
	repo repository.ApplicationRepository
}

func NewApplicationService(repo repository.ApplicationRepository) *ApplicationService {
	return &ApplicationService{repo: repo}
}

func (s *ApplicationService) List(ctx context.Context, params req.ApplicationList) ([]resp.ApplicationListItem, int, error) {
	content, count, err := s.repo.List(ctx, params)
	if err != nil {
		logger.Error("failed to retrieve the list ", err)
		return nil, 0, errors.New("failed to retrieve the list")
	}

	return content, count, nil
}

func (s *ApplicationService) Info(ctx context.Context, params req.InfoPathParams) (*resp.ApplicationInfo, error) {
	applicationInfo, err := s.repo.GetByID(ctx, params.ID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, errors.New("application not found")
		}
		logger.Error("failed to get application info", err)
		return nil, errors.New("failed to get application info")
	}

	return &resp.ApplicationInfo{
		ID:           applicationInfo.ID,
		Name:         applicationInfo.Name,
		ClientID:     applicationInfo.ClientID,
		RedirectUris: applicationInfo.RedirectUris,
	}, nil
}

func (s *ApplicationService) Create(ctx context.Context, body req.ApplicationCreate) (*resp.ApplicationCreate, error) {
	exists, err := s.repo.Duplicate(ctx, body.Name, body.ClientID)
	if err != nil {
		logger.Error("failed to check if application exists", err)
		return nil, errors.New("failed to check if application exists")
	}

	if exists {
		return nil, errors.New("name or client_id or client_secret already exists")
	}

	clientSecret, err := utils.RandomAlphanumeric(64)
	if err != nil {
		return nil, err
	}

	applicationInfo, err := s.repo.Create(ctx, body, clientSecret)
	if err != nil {
		logger.Error("failed to create application", err)
		return nil, errors.New("failed to create application")
	}

	return &resp.ApplicationCreate{
		ID:           applicationInfo.ID,
		Name:         applicationInfo.Name,
		ClientID:     applicationInfo.ClientID,
		ClientSecret: clientSecret,
		RedirectUris: applicationInfo.RedirectUris,
	}, nil
}

func (s *ApplicationService) Update(ctx context.Context, params req.ApplicationUpdatePathParams, body req.ApplicationUpdate) (*resp.ApplicationUpdate, error) {
	applicationInfo, err := s.repo.GetByID(ctx, params.ApplicationID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, errors.New("application not found")
		}
		logger.Error("get application failed", err)
		return nil, errors.New("failed to get application info")
	}

	exists, err := s.repo.Duplicate(ctx, body.Name, body.ClientID, params.ApplicationID)
	if err != nil {
		logger.Error("failed to check if application exists", err)
		return nil, errors.New("failed to check if application exists")
	}
	if exists {
		return nil, errors.New("name or client_id or client_secret already exists")
	}

	_, err = s.repo.Update(ctx, params, body)
	if err != nil {
		logger.Error("failed to update application", err)
		return nil, errors.New("failed to update application")
	}

	return &resp.ApplicationUpdate{
		ID:           applicationInfo.ID,
		Name:         applicationInfo.Name,
		ClientID:     applicationInfo.ClientID,
		RedirectUris: applicationInfo.RedirectUris,
	}, nil
}

func (s *ApplicationService) UpdateSecret(ctx context.Context, params req.ApplicationUpdatePathParams) (*resp.ApplicationUpdateSecret, error) {
	clientSecret, err := utils.RandomAlphanumeric(64)
	if err != nil {
		return nil, err
	}

	applicationInfo, err := s.repo.UpdateSecret(ctx, params.ApplicationID, clientSecret)
	if err != nil {
		return nil, err
	}

	return &resp.ApplicationUpdateSecret{
		ID:           applicationInfo.ID,
		Name:         applicationInfo.Name,
		ClientID:     applicationInfo.ClientID,
		ClientSecret: clientSecret,
		RedirectUris: applicationInfo.RedirectUris,
	}, nil
}

func (s *ApplicationService) Delete(ctx context.Context, params req.DeletePathParams) error {
	_, err := s.repo.GetByID(ctx, params.ID)
	if err != nil {
		if db.IsNotFound(err) {
			return errors.New("application not found")
		}
		logger.Error("get application failed", err.Error())
		return errors.New("failed to get application info")
	}

	if err := s.repo.Delete(ctx, params); err != nil {
		logger.Error("failed to delete application", err)
		return errors.New("failed to delete application")
	}
	return nil
}
