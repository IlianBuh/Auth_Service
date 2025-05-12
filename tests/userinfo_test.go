package tests

import (
	"Service/internal/config"
	"Service/tests/suite"
	"testing"

	authv1 "github.com/IlianBuh/SSO_Protobuf/gen/go/auth"
	userinfov1 "github.com/IlianBuh/SSO_Protobuf/gen/go/userinfo"
	"github.com/stretchr/testify/require"
)

func TestUsersByLogin(t *testing.T) {
	cfg := config.New()
	ctx, stu := suite.NewSuiteUserInfo(t, cfg)
	_, sta := suite.NewSuiteAuth(t, cfg)

	tt := []struct {
		login    string
		email    string
		password string
	}{
		{
			login:    "__login1",
			email:    "email01@email.com",
			password: "weriuowjfskl",
		},
		{
			login:    "__login2",
			email:    "email02@email.com",
			password: "weriuowjfskl",
		},
	}

	for _, test := range tt {
		_, err := sta.Client.SignUp(ctx, &authv1.SignUpRequest{
			Login:    test.login,
			Email:    test.email,
			Password: test.password,
		})
		require.NoError(t, err, "failed to sign up user")
	}

	users, err := stu.Client.UsersByLogin(ctx, &userinfov1.UsersByLoginRequest{
		Login: "_",
	})
	require.NoError(t, err)

	for i, user := range users.GetUsers() {
		require.Equal(t, tt[i].login, user.Login)
		require.Equal(t, tt[i].email, user.Email)
	}
}
