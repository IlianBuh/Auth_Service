package tests

import (
	"Service/internal/config"
	"Service/tests/suite"
	"fmt"
	"math/rand"
	"testing"

	authv1 "github.com/IlianBuh/SSO_Protobuf/gen/go/auth"
	followv1 "github.com/IlianBuh/SSO_Protobuf/gen/go/follow"
	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	cfg := config.New()
	ctx, s := suite.NewSuiteAuth(t, cfg)
	ctx, sf := suite.NewSuiteFollow(t, cfg)

	loginSet := make(map[string]bool)
	for range 200 {
		login := generateLoginWord()
		if loginSet[login] {
			continue
		}

		_, err := s.Client.SignUp(
			ctx,
			&authv1.SignUpRequest{
				Login:    login,
				Email:    fmt.Sprintf("%s@example.com", login),
				Password: "password",
			},
		)
		require.NoError(t, err)

	}

	followList := make(map[int]struct{}, 20)
	for i := range 200 {
		clear(followList)
		i++
		for range rand.Int()%20 + 5 {
			id := rand.Int()%200 + 1
			if _, ok := followList[id]; ok || i == id {
				continue
			}

			followList[id] = struct{}{}
			_, err := sf.Client.Follow(
				t.Context(),
				&followv1.FollowRequest{
					Src:    int32(i),
					Target: int32(id),
				},
			)
			require.NoError(t, err)
		}
	}

}

func generateLoginWord() string {
	adjectives := []string{
		"quick", "brave", "lazy", "happy", "clever", "noisy", "bright", "silent", "shiny", "rough",
		"calm", "wild", "kind", "fancy", "bold", "chilly", "graceful", "bitter", "zany", "quirky",
	}
	nouns := []string{
		"fox", "lion", "tiger", "panda", "owl", "wolf", "eagle", "shark", "bear", "hawk",
		"snake", "otter", "whale", "rhino", "camel", "gecko", "lemur", "sloth", "crab", "crow",
	}

	return fmt.Sprintf("%s%s%d", adjectives[rand.Intn(len(adjectives))], nouns[rand.Intn(len(nouns))], rand.Intn(10000))
}
