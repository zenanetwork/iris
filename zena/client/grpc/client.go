package grpc

import (
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/log"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"

	proto "github.com/zenanetwork/zenaproto/zena"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

type ZenaGRPCClient struct {
	conn   *grpc.ClientConn
	client proto.ZenaApiClient
}

func NewZenaGRPCClient(address string) *ZenaGRPCClient {
	address = removePrefix(address)

	opts := []grpc_retry.CallOption{
		grpc_retry.WithMax(5),
		grpc_retry.WithBackoff(grpc_retry.BackoffLinear(1 * time.Second)),
		grpc_retry.WithCodes(codes.Internal, codes.Unavailable, codes.Azenated, codes.NotFound),
	}

	conn, err := grpc.NewClient(address,
		grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(opts...)),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(opts...)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Crit("Failed to connect to Zena gRPC", "error", err)
	}

	log.Info("Connected to Zena gRPC server", "address", address)

	return &ZenaGRPCClient{
		conn:   conn,
		client: proto.NewZenaApiClient(conn),
	}
}

func (h *ZenaGRPCClient) Close() {
	log.Debug("Shutdown detected, Closing Zena gRPC client")
	h.conn.Close()
}

// removePrefix removes the http:// or https:// prefix from the address, if present.
func removePrefix(address string) string {
	if strings.HasPrefix(address, "http://") || strings.HasPrefix(address, "https://") {
		return address[strings.Index(address, "//")+2:]
	}
	return address
}
