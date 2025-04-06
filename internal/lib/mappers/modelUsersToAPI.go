package mappers

import (
	"Service/internal/domain/models"
	userinfov1 "github.com/IlianBuh/SSO_Protobuf/gen/go/userinfo"
)

func ModelUsersToAPI(users ...models.User) []*userinfov1.User {
	res := make([]*userinfov1.User, len(users))

	for i, user := range users {
		res[i] = &userinfov1.User{
			Uuid:  int32(user.UUID),
			Login: user.Login,
			Email: user.Email,
		}
	}

	return res
}
