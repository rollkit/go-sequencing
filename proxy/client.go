package proxy

import (
	"fmt"
	"net/url"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/rollkit/go-sequencing"
	proxygrpc "github.com/rollkit/go-sequencing/proxy/grpc"
)

// NewClient creates a new Sequencer client.
func NewClient(uri string) (sequencing.Sequencer, error) {
	addr, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	switch addr.Scheme {
	case "grpc":
		grpcClient := proxygrpc.NewClient()
		if err := grpcClient.Start(addr.Host, grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
			return nil, err
		}
		return grpcClient, nil
	default:
		return nil, fmt.Errorf("unknown url scheme '%s'", addr.Scheme)
	}
}
