package suite

import (
	"Service/internal/config"
	"context"
	authv1 "github.com/IlianBuh/SSO_Protobuf/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"strconv"
	"testing"
)

type SuiteAuth struct {
	*testing.T
	Cfg    *config.Config
	Client authv1.AuthClient
}

func NewSuiteAuth(t *testing.T, cfg *config.Config) (context.Context, *SuiteAuth) {
	t.Helper()
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)
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

	client := authv1.NewAuthClient(cc)
	return ctx, &SuiteAuth{
		Client: client,
		Cfg:    cfg,
	}
}
