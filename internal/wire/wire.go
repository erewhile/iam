//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/erewhile/iam/internal/ent/db"
	"github.com/erewhile/iam/internal/handler"
	"github.com/erewhile/iam/internal/repository"
	"github.com/erewhile/iam/internal/service"
	"github.com/google/wire"
)

type App struct {
	Cert *handler.CertHandler
	User *handler.UserHandler
}

var RepositorySet = wire.NewSet(
	repository.NewUserRepository,
	repository.NewTokenRepository,
)

var ServiceSet = wire.NewSet(
	service.NewUserService,
)

var HandlerSet = wire.NewSet(
	handler.NewCertHandler,
	handler.NewUserHandler,
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
