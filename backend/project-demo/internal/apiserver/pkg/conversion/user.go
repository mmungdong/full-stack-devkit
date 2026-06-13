package conversion

import (
	"github.com/onexstack/onexstack/pkg/core"

	"github.com/mungdong/devkit/internal/apiserver/model"
	v1 "github.com/mungdong/devkit/pkg/api/apiserver/v1"
)

// UserMToUserV1 converts a UserM object to a User object in the v1 API format.
func UserMToUserV1(userM *model.UserM) *v1.User {
	var user v1.User
	_ = core.CopyWithConverters(&user, userM)
	return &user
}

// UserV1ToUserM converts a User object from the v1 API format to UserM object.
func UserV1ToUserM(user *v1.User) *model.UserM {
	var userM model.UserM
	_ = core.CopyWithConverters(&userM, user)
	return &userM
}
