/**
 * Copyright 2021 IBM Corp.
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
	auth "IBM/secret-utils-lib/pkg/authenticator"

	"go.uber.org/zap"
)

// UnmanagedSecretProvider ...
type UnmanagedSecretProvider struct {
	authenticator auth.Authenticator
	logger        *zap.Logger
	authType      string
}

// newUnmanagedSecretProvider ...
func newUnmanagedSecretProvider(logger *zap.Logger) (*UnmanagedSecretProvider, error) {
	logger.Info("Initliazing unmanaged secret provider")
	authenticator, authType, err := auth.NewAuthenticator(logger)
	if err != nil {
		logger.Error("Error initializing unmanaged secret provider", zap.Error(err))
		return nil, err
	}
	logger.Info("Initliazed unmanaged secret provider")
	return &UnmanagedSecretProvider{authenticator: authenticator, logger: logger, authType: authType}, nil
}

// GetDefaultIAMToken ...
func (usp *UnmanagedSecretProvider) GetDefaultIAMToken(isFreshTokenRequired bool) (string, uint64, error) {
	usp.logger.Info("Fetching IAM token for default secret")
	return usp.authenticator.GetToken(isFreshTokenRequired)
}

// GetIAMToken ...
func (usp *UnmanagedSecretProvider) GetIAMToken(secret string, isFreshTokenRequired bool) (string, uint64, error) {
	usp.logger.Info("Fetching IAM token the provided secret")
	var authenticator auth.Authenticator
	switch usp.authType {
	case auth.IAM:
		authenticator = auth.NewIamAuthenticator(secret, usp.logger)
	case auth.PODIDENTITY:
		authenticator = auth.NewComputeIdentityAuthenticator(secret, usp.logger)
	}

	token, tokenlifetime, err := authenticator.GetToken(isFreshTokenRequired)
	if err != nil {
		usp.logger.Error("Error fetching IAM token", zap.Error(err))
		return token, tokenlifetime, err
	}
	usp.logger.Info("Successfully fetched IAM token for the provided secret")
	return token, tokenlifetime, nil
}
