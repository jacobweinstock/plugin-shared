package shared

import (
	context "context"
	"fmt"
	"os"

	"github.com/hashicorp/go-plugin"
	grpc "google.golang.org/grpc"
)

// handshakeConfigs are used to just do a basic handshake between
// a plugin and host. If the handshake fails, a user friendly error is shown.
// This prevents users from executing bad plugins or executing a plugin
// directory. It is a UX feature, not a security feature.
var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

// This is the implementation of plugin.GRPCPlugin so we can serve/consume this.
type KCPGRPCPlugin struct {
	// GRPCPlugin must still implement the Plugin interface
	plugin.Plugin
	// Concrete implementation, written in Go. This is only used for plugins
	// that are written in Go.
	Impl KCP
}

type KCP interface {
	Start() error
}

// GRPCClient is an implementation of KV that talks over RPC.
type GRPCClient struct{ client KCPClient }

// Here is the gRPC server that GRPCClient talks to.
type GRPCServer struct {
	// This is the real implementation
	Impl KCP
}

func (p *KCPGRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	RegisterKCPServer(s, &GRPCServer{Impl: p.Impl})
	return nil
}

func (p *KCPGRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: NewKCPClient(c)}, nil
}

func (g *GRPCServer) Start(ctx context.Context, req *Empty) (*Empty, error) {
	fmt.Fprintln(os.Stderr, "in plugin: RPCServer.Start")
	return &Empty{}, g.Impl.Start()
}

func (g *GRPCClient) Start(ctx context.Context, req *Empty) (*Empty, error) {
	fmt.Fprintln(os.Stderr, "in plugin: RPCClient.Start")
	var err error
	req, e := g.client.Start(ctx, req)
	if e != nil {
		return req, e
	}

	return req, err
}
