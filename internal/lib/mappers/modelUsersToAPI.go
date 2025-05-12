package mappers

import (
	"Service/internal/domain/models"
	userv1 "github.com/IlianBuh/SSO_Protobuf/gen/go/user"
)

func ModelUsersToAPI(users ...models.User) []*userv1.User {
	res := make([]*userv1.User, len(users))

	for i, user := range users {
		res[i] = &userv1.User{
			Uuid:  int32(user.UUID),
			Login: user.Login,
			Email: user.Email,
		}
	}

	return res
}
