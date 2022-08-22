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
)

func main() {
	// Initializing secret provider
	// Pre requisites and behavior are mentioned in READ me.
	// First argument is the provider type, which can be sp.VPC/sp.Bluemix/sp.Softlayer
	// Note: The first argument is of use, only when storage-secret-store is used (so that api key can be fetched either VPC or Bluemix or Softlayer section in slclient.toml)
	secretprovider, err := sp.NewSecretProvider(sp.VPC)
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
	fmt.Println(secretprovider.GetDefaultIAMToken("test", false))
}
