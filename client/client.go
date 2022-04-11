package main

import (
	"fmt"

	sp "github.com/IBM/secret-common-lib/pkg/secret_provider"
)

func main() {
	secretprovider, _ := sp.NewSecretProvider()

	// set freshTokenRequired to true is a fresh token is required
	// else, set it to false, if the existing token is valid, the same will be returned, else a new token will be fetched from iam.
	freshTokenRequired := true

	// Call to get IAM token for the default secret
	fmt.Println(secretprovider.GetDefaultIAMToken(freshTokenRequired))

	// Call to get IAM token for a different secret
	secret := "valid-secret"
	fmt.Println(secretprovider.GetIAMToken(secret, freshTokenRequired))
}
