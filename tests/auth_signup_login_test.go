package tests

import (
	"Service/tests/suite"
	authv1 "github.com/IlianBuh/Auth_Protobuf/gen/go"
	"github.com/brianvoe/gofakeit"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSignUpLoginPositive(t *testing.T) {
	ctx, st := suite.New(t)

	login := gofakeit.FirstName()
	email := gofakeit.Email()
	pass := randomFakePassword()

	signUpTime := time.Now()
	respSignUp, err := st.Client.SignUp(ctx, &authv1.SignUpRequest{
		Login:    login,
		Email:    email,
		Password: pass,
	})
	require.NoError(t, err)
	tokenSignUp := respSignUp.GetToken()
	require.NotEmpty(t, tokenSignUp)

	loginTime := time.Now()
	respLogin, err := st.Client.Login(ctx, &authv1.LoginRequest{
		Login:    login,
		Password: pass,
	})
	require.NoError(t, err)
	tokenLogin := respLogin.GetToken()
	require.NotEmpty(t, tokenLogin)

	parsedSignUp, err := jwt.Parse(tokenSignUp, func(token *jwt.Token) (interface{}, error) {
		return []byte(st.Cfg.Secret), nil
	})
	require.NoError(t, err)

	parsedLogin, err := jwt.Parse(tokenLogin, func(token *jwt.Token) (interface{}, error) {
		return []byte(st.Cfg.Secret), nil
	})
	require.NoError(t, err)

	clSignUp, ok := parsedSignUp.Claims.(jwt.MapClaims)
	require.True(t, ok)
	clLogin, ok := parsedLogin.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, clSignUp["uuid"].(float64), clLogin["uuid"].(float64))
	assert.Equal(t, clSignUp["login"].(string), clLogin["login"].(string))

	const deltaSeconds = 1

	assert.InDelta(t, signUpTime.Add(st.Cfg.TokenTTL).Unix(), clSignUp["exp"].(float64), deltaSeconds)
	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), clLogin["exp"].(float64), deltaSeconds)
}

func TestDoubleSignUp(t *testing.T) {
	ctx, st := suite.New(t)

	login := gofakeit.FirstName()
	email := gofakeit.Email()
	pass := randomFakePassword()

	respSignUp, err := st.Client.SignUp(ctx, &authv1.SignUpRequest{
		Login:    login,
		Email:    email,
		Password: pass,
	})
	require.NoError(t, err)
	require.NotEmpty(t, respSignUp.GetToken())

	respSignUp, err = st.Client.SignUp(ctx, &authv1.SignUpRequest{
		Login:    login,
		Email:    email,
		Password: pass,
	})
	require.Error(t, err)
	assert.Empty(t, respSignUp.GetToken())
	assert.ErrorContains(t, err, "invalid arguments")

}
func randomFakePassword() string {
	const passwordDefaultLen = 20
	return gofakeit.Password(true, true, true, true, false, passwordDefaultLen)
}
