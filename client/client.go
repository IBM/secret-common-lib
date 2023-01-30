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

package main

import (
	"fmt"

	sp "github.com/IBM/secret-common-lib/pkg/secret_provider"
	"github.com/IBM/secret-utils-lib/pkg/k8s_utils"
)

func main() {
	// Initializing secret provider
	// Pre requisites and behavior are mentioned in READ me.

	arg := map[string]string{
		sp.ProviderType: sp.VPC,
	}
	// OR
	//arg := map[string]string{
	//	sp.SecretKey: "iam_api_key",
	//}

	// For the client code to work, initialise a fake k8s client
	/*
		k8sClient, _ := k8s_utils.FakeGetk8sClientSet()
		pwd, _ := os.Getwd()
		secretDataPath := filepath.Join(pwd, "..", "test-fixtures/secrets/storage-secret-store", "slclient.toml")
		_ = k8s_utils.FakeCreateSecret(k8sClient, "DEFAULT", secretDataPath)
	*/

	// For real time scenarios, the following can be done
	k8sClient, _ := k8s_utils.Getk8sClientSet()
	secretprovider, err := sp.NewSecretProvider(&k8sClient, arg)
	// OR, you may also provide a key which is expected to be present the k8s secret - ibm-cloud-credentials/storage-secret-store
	// secretprovider, err := sp.NewSecretProvider(sp.VPC, "your-key")

	if err != nil {
		fmt.Println("Error initializing provider")
		fmt.Println(err)
		return
	}

	// API calls
	fmt.Println(secretprovider.GetContainerAPIRoute(false))
	fmt.Println(secretprovider.GetPrivateRIAASEndpoint(false))
	fmt.Println(secretprovider.GetRIAASEndpoint(false))
	fmt.Println(secretprovider.GetPrivateContainerAPIRoute(false))
	fmt.Println(secretprovider.GetDefaultIAMToken(false, "test"))
}
