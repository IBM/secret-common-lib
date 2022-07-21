/**
 * Copyright 2022 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package secret_provider

import (
	"context"
	"flag"
	"net"
	"time"

	localutils "github.com/IBM/secret-common-lib/pkg/utils"
	"github.com/IBM/secret-utils-lib/pkg/k8s_utils"
	"github.com/IBM/secret-utils-lib/pkg/utils"
	sp "github.com/IBM/secret-utils-lib/secretprovider"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	endpoint = flag.String("sidecarEndpoint", "/csi/provider.sock", "Storage secret sidecar endpoint")
)

// ManagedSecretProvider ...
type ManagedSecretProvider struct {
	logger                   *zap.Logger
	k8sClient                k8s_utils.KubernetesClient
	region                   string
	riaasEndpoint            string
	privateRIAASEndpoint     string
	containerAPIRoute        string
	privateContainerAPIRoute string
	resourceGroupID          string
}

// newManagedSecretProvider ...
func newManagedSecretProvider(logger *zap.Logger, providerType string) (*ManagedSecretProvider, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	_, err := grpc.DialContext(ctx, *endpoint, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithContextDialer(unixConnect))
	if err != nil {
		logger.Error("Error establishing grpc connection to secret sidecar", zap.Error(err))
		return nil, utils.Error{Description: "Error establishing grpc connection", BackendError: err.Error()}
	}

	logger.Info("Initialized managed secret provider")
	return &ManagedSecretProvider{logger: logger}, nil
}

// GetDefaultIAMToken ...
func (msp *ManagedSecretProvider) GetDefaultIAMToken(freshTokenRequired bool) (string, uint64, error) {
	var tokenlifetime uint64
	// Connecting to sidecar
	msp.logger.Info("Connecting to sidecar")
	conn, err := grpc.Dial(*endpoint, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithContextDialer(unixConnect))
	if err != nil {
		msp.logger.Error("Error establishing grpc connection to secret sidecar", zap.Error(err))
		return "", tokenlifetime, utils.Error{Description: "Error establishing grpc connection to secret sidecar", BackendError: err.Error()}
	}

	c := sp.NewSecretProviderClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	defer conn.Close()

	response, err := c.GetDefaultIAMToken(ctx, &sp.Request{IsFreshTokenRequired: freshTokenRequired})
	if err != nil {
		msp.logger.Error("Error fetching IAM token", zap.Error(err))
		return "", tokenlifetime, err
	}

	msp.logger.Info("Fetched IAM token for default secret")
	return response.Iamtoken, response.Tokenlifetime, nil
}

// GetIAMToken ...
func (msp *ManagedSecretProvider) GetIAMToken(secret string, freshTokenRequired bool) (string, uint64, error) {
	var tokenlifetime uint64

	msp.logger.Info("Connecting to sidecar")
	conn, err := grpc.Dial(*endpoint, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithContextDialer(unixConnect))
	if err != nil {
		msp.logger.Error("Error establishing grpc connection to secret sidecar", zap.Error(err))
		return "", tokenlifetime, utils.Error{Description: "Error establishing grpc connection to secret sidecar", BackendError: err.Error()}
	}

	c := sp.NewSecretProviderClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	defer conn.Close()

	response, err := c.GetIAMToken(ctx, &sp.Request{Secret: secret, IsFreshTokenRequired: freshTokenRequired})
	if err != nil {
		msp.logger.Error("Error fetching IAM token", zap.Error(err))
		return "", tokenlifetime, err
	}

	msp.logger.Info("Fetched IAM token for the provided secret")
	return response.Iamtoken, response.Tokenlifetime, nil
}

// unixConnect ...
func unixConnect(ctx context.Context, addr string) (net.Conn, error) {
	unixAddr, err := net.ResolveUnixAddr("unix", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUnix("unix", nil, unixAddr)
	return conn, err
}

// GetRIAASEndpoint ...
func (msp *ManagedSecretProvider) GetRIAASEndpoint() (string, error) {
	msp.logger.Info("In GetRIAASEndpoint()")
	endpoint, err := getEndpoint(localutils.RIAAS, msp.riaasEndpoint, msp.k8sClient, msp.logger)
	if err != nil {
		return "", err
	}

	msp.riaasEndpoint = endpoint
	return endpoint, nil
}

// GetPrivateRIAASEndpoint ...
func (msp *ManagedSecretProvider) GetPrivateRIAASEndpoint() (string, error) {
	msp.logger.Info("In GetPrivateRIAASEndpoint()")
	endpoint, err := getEndpoint(localutils.PrivateRIAAS, msp.privateRIAASEndpoint, msp.k8sClient, msp.logger)
	if err != nil {
		return "", err
	}

	msp.privateRIAASEndpoint = endpoint
	return endpoint, nil
}

// GetContainerAPIRoute ...
func (msp *ManagedSecretProvider) GetContainerAPIRoute() (string, error) {
	msp.logger.Info("In GetContainerAPIRoute()")
	endpoint, err := getEndpoint(localutils.ContainerAPIRoute, msp.containerAPIRoute, msp.k8sClient, msp.logger)
	if err != nil {
		return "", err
	}

	msp.containerAPIRoute = endpoint
	return endpoint, nil
}

// GetPrivateContainerAPIRoute ...
func (msp *ManagedSecretProvider) GetPrivateContainerAPIRoute() (string, error) {
	msp.logger.Info("In GetPrivateContainerAPIRoute()")
	endpoint, err := getEndpoint(localutils.PrivateContainerAPIRoute, msp.privateContainerAPIRoute, msp.k8sClient, msp.logger)
	if err != nil {
		return "", err
	}

	msp.privateContainerAPIRoute = endpoint
	return endpoint, nil
}

// GetResourceGroupID ...
func (msp *ManagedSecretProvider) GetResourceGroupID() string {
	return msp.resourceGroupID
}
