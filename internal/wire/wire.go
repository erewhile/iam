//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/erewhile/iam/internal/cache/rds"
	"github.com/erewhile/iam/internal/ent/db"
	"github.com/erewhile/iam/internal/handler"
	"github.com/erewhile/iam/internal/repository"
	"github.com/erewhile/iam/internal/service"
	"github.com/google/wire"
)

type App struct {
	Cert        *handler.CertHandler
	User        *handler.UserHandler
	Role        *handler.RoleHandler
	UserRole    *handler.UserRoleHandler
	Token       *handler.TokenHandler
	OAuth       *handler.OAuthHandler
	Application *handler.ApplicationHandler
}

var RepositorySet = wire.NewSet(
	repository.NewUserRepository,
	repository.NewTokenRepository,
	repository.NewTransactor,
	repository.NewUserRoleRepository,
	repository.NewRoleRepository,
	repository.NewApplicationRepository,
)

var ServiceSet = wire.NewSet(
	service.NewUserService,
	service.NewUserRoleService,
	service.NewRoleService,
	service.NewTokenService,
	service.NewApplicationService,
	rds.NewTokenCache,
	service.NewOAuthService,
	rds.NewIAMSessionCache,
)

var HandlerSet = wire.NewSet(
	handler.NewCertHandler,
	handler.NewUserHandler,
	handler.NewRoleHandler,
	handler.NewUserRoleHandler,
	handler.NewTokenHandler,
	handler.NewOAuthHandler,
	handler.NewApplicationHandler,
)

var providerSet = wire.NewSet(
	RepositorySet,
	ServiceSet,
	HandlerSet,
	wire.Struct(new(App), "*"),
)

func InitApp(client *db.Client) *App {
	wire.Build(providerSet)
	return &App{}
}
