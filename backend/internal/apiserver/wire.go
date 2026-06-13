//go:build wireinject
// +build wireinject

package apiserver

import (
	"context"

	"github.com/google/wire"
	"github.com/onexstack/onexstack/pkg/authz"

	"github.com/mungdong/devkit/internal/apiserver/biz"
	"github.com/mungdong/devkit/internal/apiserver/handler"
	"github.com/mungdong/devkit/internal/apiserver/pkg/validation"
	"github.com/mungdong/devkit/internal/apiserver/store"
	mw "github.com/mungdong/devkit/internal/pkg/middleware/gin"
)

// infrastructureSet groups all infrastructure-related providers.
// This keeps the main wire.Build call clean.
var infrastructureSet = wire.NewSet(
	ProvideDB,
	wire.NewSet(
		wire.Struct(new(UserRetriever), "*"),
		wire.Bind(new(mw.UserRetriever), new(*UserRetriever)),
	),
	authz.ProviderSet,
)

// NewServer initializes and creates the web server with all necessary dependencies using Wire.
func NewServer(context.Context, *Config) (*Server, error) {
	wire.Build(
		// Server infrastructure
		NewWebServer,
		NewDependencies,
		wire.Struct(new(ServerConfig), "*"), // Inject all fields
		wire.Struct(new(Server), "*"),

		// Domain layers
		store.ProviderSet,
		biz.ProviderSet,
		validation.ProviderSet,
		handler.NewHandler,

		// Infrastructure dependencies
		infrastructureSet,
	)
	return nil, nil
}
