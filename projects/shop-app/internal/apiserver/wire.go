//go:build wireinject
// +build wireinject

package apiserver

import (
	"github.com/google/wire"
	"github.com/onexstack/onexstack/pkg/authz"

	"github.com/onexstack/shop-app/internal/apiserver/biz"
	"github.com/onexstack/shop-app/internal/apiserver/pkg/validation"
	"github.com/onexstack/shop-app/internal/apiserver/store"
	mw "github.com/onexstack/shop-app/internal/pkg/middleware/gin"
)

// NewServer sets up and create the web server with all necessary dependencies.
func NewServer(*Config) (*Server, error) {
	wire.Build(
		NewWebServer,
		wire.Struct(new(ServerConfig), "*"), // * 表示注入全部字段
		wire.Struct(new(Server), "*"),
		wire.NewSet(store.ProviderSet, biz.ProviderSet),
		ProvideDB, // 提供数据库实例
		validation.ProviderSet,
		wire.NewSet(
			wire.Struct(new(UserRetriever), "*"),
			wire.Bind(new(mw.UserRetriever), new(*UserRetriever)),
		),
		authz.ProviderSet,
	)
	return nil, nil
}
