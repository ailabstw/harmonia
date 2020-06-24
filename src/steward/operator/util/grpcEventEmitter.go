package util

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"google.golang.org/grpc"
)

func EmitEvent(
	clientURI string, 
	newClient func(*grpc.ClientConn) interface{},
	emitEvent func(context.Context, interface{}) (interface{}, error),
) {
	opts := []grpc.DialOption {
		grpc.WithInsecure(),
		grpc.WithBlock(),
	}

	conn, err := grpc.Dial(clientURI, opts...)
	if err != nil {
		zap.L().Fatal("fail to dail grpc", zap.Error(err))
	}
	defer conn.Close()

	client := newClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	response, err := emitEvent(ctx, client)
	if err != nil {
		if err == context.DeadlineExceeded {
			zap.L().Fatal("Deadline exceeded")
		} else {
			zap.L().Fatal("emitEvent get error", zap.Error(err))
		}
	}

	zap.L().Debug("received response", zap.String("response", fmt.Sprintf("%v", response)))
}
