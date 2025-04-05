package suite

import (
	"Service/internal/config"
	"context"
	authv1 "github.com/IlianBuh/Auth_Protobuf/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"strconv"
	"testing"
)

type Suite struct {
	*testing.T
	Cfg    *config.Config
	Client authv1.AuthClient
}

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()
	cfg := config.New()

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
	return ctx, &Suite{
		Client: client,
		Cfg:    cfg,
	}
}
