package biz

import (
	"github.com/google/wire"
	"github.com/onexstack/onexstack/pkg/authz"

	userv1 "github.com/mungdong/devkit/internal/apiserver/biz/v1/user"
	"github.com/mungdong/devkit/internal/apiserver/store"
)

// ProviderSet declares dependency injection rules for the business logic layer.
var ProviderSet = wire.NewSet(NewBiz, wire.Bind(new(IBiz), new(*biz)))

// IBiz defines the access points for various business logic modules.
type IBiz interface {
	// UserV1 gets the user business interface.
	UserV1() userv1.UserBiz
}

// biz is the concrete implementation of the business logic IBiz.
type biz struct {
	store store.IStore
	authz *authz.Authz
}

// Ensure biz implements IBiz at compile time.
var _ IBiz = (*biz)(nil)

// NewBiz creates and returns a new instance of the business logic layer.
func NewBiz(store store.IStore, authz *authz.Authz) *biz {
	return &biz{store: store, authz: authz}
}

// UserV1 returns an instance that implements the UserBiz interface.
func (b *biz) UserV1() userv1.UserBiz {
	return userv1.New(b.store, b.authz)
}
