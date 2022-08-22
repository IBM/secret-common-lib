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
	secretprovider, err := sp.NewSecretProvider(sp.VPC)
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
