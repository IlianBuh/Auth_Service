package suite

import (
	"Service/internal/config"
	"context"
	followv1 "github.com/IlianBuh/SSO_Protobuf/gen/go/follow"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"strconv"
	"testing"
)

type SuiteFollow struct {
	*testing.T
	Cfg    *config.Config
	Client followv1.FollowClient
}

func NewSuiteFollow(t *testing.T, cfg *config.Config) (context.Context, *SuiteFollow) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() {
		t.Helper()
		cancel()
	})

	srvAddr := net.JoinHostPort("localhost", strconv.Itoa(cfg.GRPC.Port))
	cc, err := grpc.NewClient(
		srvAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to connect to grpc server: %v", err)
	}

	client := followv1.NewFollowClient(cc)
	return ctx, &SuiteFollow{
		Client: client,
		Cfg:    cfg,
	}
}
