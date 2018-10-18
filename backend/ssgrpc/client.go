package ssgrpc

import (
	"github.com/susilolab/sacrificial-socket/backend/ssgrpc/transport"
	"google.golang.org/grpc"
	//"sync"
)

type propagateClient struct {
	conn   *grpc.ClientConn
	client transport.PropagateClient
}
