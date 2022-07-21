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
	"encoding/base64"
	"os"
	"strings"

	localutils "github.com/IBM/secret-common-lib/pkg/utils"
	auth "github.com/IBM/secret-utils-lib/pkg/authenticator"
	"github.com/IBM/secret-utils-lib/pkg/config"
	"github.com/IBM/secret-utils-lib/pkg/k8s_utils"
	"github.com/IBM/secret-utils-lib/pkg/utils"
	"go.uber.org/zap"
)

// UnmanagedSecretProvider ...
type UnmanagedSecretProvider struct {
	authenticator            auth.Authenticator
	logger                   *zap.Logger
	k8sClient                k8s_utils.KubernetesClient
	authType                 string
	tokenExchangeURL         string
	region                   string
	riaasEndpoint            string
	privateRIAASEndpoint     string
	containerAPIRoute        string
	privateContainerAPIRoute string
	resourceGroupID          string
}

// newUnmanagedSecretProvider ...
func newUnmanagedSecretProvider(logger *zap.Logger, providerType string, secretKey ...string) (*UnmanagedSecretProvider, error) {
	kc, err := k8s_utils.Getk8sClientSet(logger)
	if err != nil {
		logger.Info("Error fetching k8s client set", zap.Error(err))
		return nil, err
	}
	return initUnmanagedSecretProvider(logger, kc, providerType, secretKey...)
}

// initUnmanagedSecretProvider ...
func initUnmanagedSecretProvider(logger *zap.Logger, kc k8s_utils.KubernetesClient, providerType string, secretKey ...string) (*UnmanagedSecretProvider, error) {
	authenticator, authType, err := auth.NewAuthenticator(logger, kc, providerType, secretKey...)
	if err != nil {
		logger.Error("Error initializing unmanaged secret provider", zap.Error(err))
		return nil, err
	}

	tokenExchangeURL := config.FrameTokenExchangeURL(kc, logger)
	authenticator.SetURL(tokenExchangeURL)

	if authenticator.IsSecretEncrypted() {
		logger.Error("Secret is encrypted, decryption is only supported by sidecar container")
		return nil, utils.Error{Description: localutils.ErrDecryptionNotSupported}
	}

	// Checking if the secret(api key) needs to be decoded
	if authType == utils.DEFAULT && os.Getenv("IS_SATELLITE") == "True" {
		logger.Info("Decoding apiKey since it's a satellite cluster")
		decodedSecret, err := base64.StdEncoding.DecodeString(authenticator.GetSecret())
		if err != nil {
			logger.Error("Error decoding the secret", zap.Error(err))
			return nil, err
		}
		// In the decoded secret, newline could be present, trimming the same to extract a valid api key.
		authenticator.SetSecret(strings.TrimSuffix(string(decodedSecret), "\n"))
	}

	usp := new(UnmanagedSecretProvider)
	usp.authenticator = authenticator
	usp.logger = logger
	usp.authType = authType
	usp.tokenExchangeURL = tokenExchangeURL
	usp.k8sClient = kc

	cloudConf, err := getCloudConf(logger, kc)
	if err == nil && cloudConf.Region != "" {
		usp.region = cloudConf.Region
		if cloudConf.ContainerAPIRoute == "" {
			usp.containerAPIRoute = constructContainerAPIRoute(usp.region)
		}
		if cloudConf.PrivateContainerAPIRoute == "" {
			usp.privateContainerAPIRoute = constructPrivateContainerAPIRoute(usp.region)
		}
		if cloudConf.RiaasEndpoint == "" {
			usp.riaasEndpoint = constructRIAASEndpoint(usp.region)
		}
		if cloudConf.PrivateRIAASEndpoint == "" {
			usp.privateRIAASEndpoint = constructPrivateRIAASEndpoint(usp.region)
		}
		usp.resourceGroupID = cloudConf.ResourceGroupID
		logger.Info("Initliazed unmanaged secret provider")
		return usp, nil
	}

	data, err := k8s_utils.GetSecretData(kc, utils.STORAGE_SECRET_STORE_SECRET, utils.SECRET_STORE_FILE)
	if err != nil {
		logger.Error("Error initializing secret provider, unable to fetch storage-secret-store")
		return nil, utils.Error{Description: localutils.ErrInitSecretProvider, BackendError: err.Error()}
	}

	conf, err := config.ParseConfig(logger, data)
	if err != nil {
		logger.Error("Error initializing secret provider, unable to parse storage-secret-store")
		return nil, utils.Error{Description: localutils.ErrInitSecretProvider, BackendError: err.Error()}
	}

	usp.containerAPIRoute = conf.Bluemix.APIEndpointURL
	usp.privateContainerAPIRoute = conf.Bluemix.PrivateAPIRoute
	usp.riaasEndpoint = conf.VPC.G2EndpointURL
	usp.privateRIAASEndpoint = conf.VPC.G2EndpointPrivateURL
	usp.resourceGroupID = conf.VPC.G2ResourceGroupID
	logger.Info("Initliazed unmanaged secret provider")
	return usp, nil
}

// GetDefaultIAMToken ...
func (usp *UnmanagedSecretProvider) GetDefaultIAMToken(isFreshTokenRequired bool) (string, uint64, error) {
	usp.logger.Info("In GetDefaultIAMToken()")
	return usp.authenticator.GetToken(true)
}

// GetIAMToken ...
func (usp *UnmanagedSecretProvider) GetIAMToken(secret string, isFreshTokenRequired bool) (string, uint64, error) {
	usp.logger.Info("In GetIAMToken()")
	var authenticator auth.Authenticator
	switch usp.authType {
	case utils.IAM, utils.DEFAULT:
		authenticator = auth.NewIamAuthenticator(secret, usp.logger)
	case utils.PODIDENTITY:
		authenticator = auth.NewComputeIdentityAuthenticator(secret, usp.logger)
	}

	authenticator.SetURL(usp.tokenExchangeURL)
	token, tokenlifetime, err := authenticator.GetToken(true)
	if err != nil {
		usp.logger.Error("Error fetching IAM token", zap.Error(err))
		return token, tokenlifetime, err
	}
	return token, tokenlifetime, nil
}

// GetRIAASEndpoint ...
func (usp *UnmanagedSecretProvider) GetRIAASEndpoint() (string, error) {
	usp.logger.Info("In GetRIAASEndpoint()")
	endpoint, err := getEndpoint(localutils.RIAAS, usp.riaasEndpoint, usp.k8sClient, usp.logger)
	if err != nil {
		return "", err
	}

	usp.riaasEndpoint = endpoint
	return endpoint, nil
}

// GetPrivateRIAASEndpoint ...
func (usp *UnmanagedSecretProvider) GetPrivateRIAASEndpoint() (string, error) {
	usp.logger.Info("In GetPrivateRIAASEndpoint()")
	endpoint, err := getEndpoint(localutils.PrivateRIAAS, usp.privateRIAASEndpoint, usp.k8sClient, usp.logger)
	if err != nil {
		return "", err
	}

	usp.privateRIAASEndpoint = endpoint
	return endpoint, nil
}

// GetContainerAPIRoute ...
func (usp *UnmanagedSecretProvider) GetContainerAPIRoute() (string, error) {
	usp.logger.Info("In GetContainerAPIRoute()")
	endpoint, err := getEndpoint(localutils.ContainerAPIRoute, usp.containerAPIRoute, usp.k8sClient, usp.logger)
	if err != nil {
		return "", err
	}

	usp.containerAPIRoute = endpoint
	return endpoint, nil
}

// GetPrivateContainerAPIRoute ...
func (usp *UnmanagedSecretProvider) GetPrivateContainerAPIRoute() (string, error) {
	usp.logger.Info("In GetPrivateContainerAPIRoute()")
	endpoint, err := getEndpoint(localutils.PrivateContainerAPIRoute, usp.privateContainerAPIRoute, usp.k8sClient, usp.logger)
	if err != nil {
		return "", err
	}

	usp.privateContainerAPIRoute = endpoint
	return endpoint, nil
}

// GetResourceGroupID ...
func (usp *UnmanagedSecretProvider) GetResourceGroupID() string {
	return usp.resourceGroupID
}
