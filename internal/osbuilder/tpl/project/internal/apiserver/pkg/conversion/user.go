package conversion

import (
	"github.com/onexstack/onexstack/pkg/core"

    "{{.D.ModuleName}}/internal/{{.Web.Name}}/model"
	{{.Web.APIImportPath}}
)

// UserModelToUserV1 将模型层的 UserM（用户模型对象）转换为 Protobuf 层的 User（v1 用户对象）.
func UserModelToUserV1(userModel *model.UserM) *{{.D.APIAlias}}.User {
	var protoUser {{.D.APIAlias}}.User
    _ = core.CopyWithConverters(&protoUser, userModel)
	return &protoUser
}

// UserV1ToUserModel 将 Protobuf 层的 User（v1 用户对象）转换为模型层的 UserM（用户模型对象）.
func UserV1ToUserModel(protoUser *{{.D.APIAlias}}.User) *model.UserM {
	var userModel model.UserM
	_ = core.CopyWithConverters(&userModel, protoUser)
	return &userModel
}
