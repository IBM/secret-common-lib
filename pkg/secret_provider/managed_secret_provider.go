package secret_provider

import (
	"context"
	"flag"
	"net"
	"time"

	sp "IBM/secret-utils-lib/secretprovider"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	endpoint = flag.String("managed-secret-provider", "/csi/provider.sock", "Storage secret sidecar endpoint")
)

// ManagedSecretProvider ...
type ManagedSecretProvider struct {
	logger *zap.Logger
}

// newManagedSecretProvider ...
func newManagedSecretProvider(logger *zap.Logger) (*ManagedSecretProvider, error) {
	logger.Info("Initializing managed secret provider, Checking if connection can be established to secret sidecar")
	_, err := grpc.Dial(*endpoint, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithDialer(unixConnect))
	if err != nil {
		logger.Error("Error establishing grpc connection to secret sidecar", zap.Error(err))
		return nil, err
	}
	logger.Info("Initialized managed secret provider")
	return &ManagedSecretProvider{logger: logger}, nil
}

// GetDefaultIAMToken ...
func (msp *ManagedSecretProvider) GetDefaultIAMToken(freshTokenRequired bool) (string, uint64, error) {
	msp.logger.Info("Fetching IAM token for default secret")
	conn, err := grpc.Dial(*endpoint, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithDialer(unixConnect))
	var token string
	var tokenlifetime uint64
	if err != nil {
		msp.logger.Error("Error establishing grpc connection to secret sidecar", zap.Error(err))
		return "", tokenlifetime, err
	}
	c := sp.NewSecretProviderClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	defer conn.Close()
	response, err := c.GetDefaultIAMToken(ctx, &sp.Request{IsFreshTokenRequired: freshTokenRequired})
	if err != nil {
		msp.logger.Error("Error fetching IAM token", zap.Error(err))
		return token, tokenlifetime, err
	}
	msp.logger.Info("Successfully fetched IAM token for default secret")
	return response.Iamtoken, response.Tokenlifetime, nil
}

// GetIAMToken ...
func (msp *ManagedSecretProvider) GetIAMToken(secret string, freshTokenRequired bool) (string, uint64, error) {
	msp.logger.Info("Fetching IAM token for the provided secret")
	conn, err := grpc.Dial(*endpoint, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithDialer(unixConnect))
	var token string
	var tokenlifetime uint64
	if err != nil {
		msp.logger.Error("Error establishing grpc connection to secret sidecar", zap.Error(err))
		return "", tokenlifetime, err
	}
	c := sp.NewSecretProviderClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	defer conn.Close()
	response, err := c.GetIAMToken(ctx, &sp.Request{Secret: secret, IsFreshTokenRequired: freshTokenRequired})
	if err != nil {
		msp.logger.Error("Error fetching IAM token", zap.Error(err))
		return token, tokenlifetime, err
	}
	msp.logger.Info("Successfully fetched IAM token for the provided secret")
	return response.Iamtoken, response.Tokenlifetime, nil
}

// unixConnect ...
func unixConnect(addr string, t time.Duration) (net.Conn, error) {
	unixAddr, err := net.ResolveUnixAddr("unix", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUnix("unix", nil, unixAddr)
	return conn, err
}
