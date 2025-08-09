package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Dial(ctx context.Context, endpoint string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			if err = conn.Close(); err != nil {
				// TODO do something when logging is a thing
			}
			return
		}
		go func() {
			<-ctx.Done()
			if err = conn.Close(); err != nil {
				// TODO do something when logging is a thing
			}
		}()
	}()

	return conn, nil
}
