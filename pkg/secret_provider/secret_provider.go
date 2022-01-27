package secret_provider

import (
	"errors"
	"os"

	utils "IBM/secret-common-lib/pkg/utils"
	sp "IBM/secret-utils-lib/pkg/secret_provider"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewSecretProvider ...
func NewSecretProvider(managed bool) (sp.SecretProviderInterface, error) {
	logger := setUpLogger()
	logger.Info("Initializing secret provider")
	var secretprovider sp.SecretProviderInterface
	var err error
	if managed {
		secretprovider, err = newManagedSecretProvider(logger)
	} else {
		secretprovider, err = newUnmanagedSecretProvider(logger)
	}

	if err != nil {
		logger.Error("Error initializing secret provider", zap.Error(err))
		return nil, errors.New(utils.ErrInitializingSecretProvider)
	}

	return secretprovider, nil
}

// setUpLogger ...
func setUpLogger() *zap.Logger {
	// Prepare a new logger
	atom := zap.NewAtomicLevel()
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		atom,
	), zap.AddCaller()).With(zap.String("name", "secret-provider")).With(zap.String("secret-provider-type", "unmanaged-secret-provider"))

	atom.SetLevel(zap.InfoLevel)
	return logger
}
