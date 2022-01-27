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
