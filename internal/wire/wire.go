//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/erewhile/iam/internal/handler"
	"github.com/google/wire"
)

type App struct {
	Cert *handler.CertHandler
}

var providerSet = wire.NewSet(
	handler.NewCertHandler,
	wire.Struct(new(App), "*"),
)

func InitApp() *App {
	wire.Build(providerSet)
	return &App{}
}
