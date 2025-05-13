package suite

import (
	"Service/internal/config"
	"context"
	"net"
	"strconv"
	"testing"

	userinfov1 "github.com/IlianBuh/SSO_Protobuf/gen/go/userinfo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SuiteUserInfo struct {
	t      *testing.T
	cfg    *config.Config
	Client userinfov1.UserInfoClient
}

func NewSuiteUserInfo(t *testing.T, cfg *config.Config) (context.Context, *SuiteUserInfo) {
	t.Helper()

	ctx, cncl := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)
	t.Cleanup(func() {
		t.Helper()
		cncl()
	})

	addr := net.JoinHostPort("localhost", strconv.Itoa(cfg.GRPC.Port))
	cc, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to create grpc new client: %v", err)
	}

	userinfo := userinfov1.NewUserInfoClient(cc)
	return ctx, &SuiteUserInfo{
		cfg:    cfg,
		Client: userinfo,
	}
}
