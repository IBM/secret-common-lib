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
	"fmt"
	"net/http"
	"strings"

	localutils "github.com/IBM/secret-common-lib/pkg/utils"
	"github.com/IBM/secret-utils-lib/pkg/config"
	"github.com/IBM/secret-utils-lib/pkg/k8s_utils"
	"github.com/IBM/secret-utils-lib/pkg/utils"
	"go.uber.org/zap"
)

const (
	// riaasEndpoint ...
	riaasEndpoint = "https://region.iaas.cloud.ibm.com"

	// privateRIAASEndpoint ...
	privateRIAASEndpoint = "https://region.private.iaas.cloud.ibm.com"

	// containerAPIRoute ...
	containerAPIRoute = "https://region.containers.cloud.ibm.com"

	// privateContainerAPIRoute ...
	privateContainerAPIRoute = "https://private.region.containers.cloud.ibm.com"
)

// getEndpoint ...
func getEndpoint(endpointName, endpointValue string, k8sClient k8s_utils.KubernetesClient, logger *zap.Logger) (string, error) {
	// Checking if the endpoint in cache is reachable
	if err := isURLReachable(endpointName, endpointValue, logger); err == nil {
		logger.Info(fmt.Sprintf("Returning %s endpoint", endpointName), zap.String("endpoint", endpointValue))
		return endpointValue, nil
	}

	// Fetching endpoint using Cloud conf
	cloudConf, err := getCloudConf(logger, k8sClient)

	// Check if the endpoint fetched from cloud-conf is reachable
	if err == nil {
		switch endpointName {
		case localutils.RIAAS:
			endpointValue = cloudConf.RiaasEndpoint
		case localutils.PrivateRIAAS:
			endpointValue = cloudConf.PrivateRIAASEndpoint
		case localutils.ContainerAPIRoute:
			endpointValue = cloudConf.ContainerAPIRoute
		case localutils.PrivateContainerAPIRoute:
			endpointValue = cloudConf.PrivateContainerAPIRoute
		}
		if err = isURLReachable(endpointName, endpointValue, logger); err == nil {
			logger.Info(fmt.Sprintf("Fetched %s endpoint from cloud-conf", endpointName), zap.String("endpoint", endpointValue))
			return endpointValue, nil
		}
	}

	// Fetching endpoint using storage-secret-store
	logger.Info(fmt.Sprintf("Fetching %s endpoint from storage-secret-store", endpointName))
	data, err := k8s_utils.GetSecretData(k8sClient, utils.STORAGE_SECRET_STORE_SECRET, utils.SECRET_STORE_FILE)
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to fetch %s endpoint from storage-secret-store", endpointName), zap.Error(err))
		return "", utils.Error{Description: fmt.Sprintf(localutils.ErrorFetchingEndpoint, endpointName), BackendError: err.Error()}
	}

	conf, err := config.ParseConfig(logger, data)
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to fetch %s endpoint from storage-secret-store", endpointName), zap.Error(err))
		return "", utils.Error{Description: fmt.Sprintf(localutils.ErrorFetchingEndpoint, endpointName), BackendError: err.Error()}
	}

	switch endpointName {
	case localutils.RIAAS:
		endpointValue = conf.VPC.G2EndpointURL
	case localutils.PrivateRIAAS:
		endpointValue = conf.VPC.G2EndpointPrivateURL
	case localutils.ContainerAPIRoute:
		endpointValue = conf.Bluemix.APIEndpointURL
	case localutils.PrivateContainerAPIRoute:
		endpointValue = conf.Bluemix.PrivateAPIRoute
	}

	if err = isURLReachable(endpointName, endpointValue, logger); err == nil {
		logger.Info(fmt.Sprintf("Fetched %s endpoint from storage-secret-store", endpointName), zap.String("endpoint", endpointValue))
		return endpointValue, nil
	}

	logger.Error(fmt.Sprintf(localutils.ErrorFetchingEndpoint, endpointName), zap.Error(err))
	return "", utils.Error{Description: fmt.Sprintf(localutils.ErrorFetchingEndpoint, endpointName), BackendError: err.Error()}
}

// isURLReachable checks if a given url is reachable
// if the provided url is unreachable, returns error
// urlName is something like - riaas, privateRIAAS, etc. Used for better logging and return message
func isURLReachable(urlName, urlValue string, logger *zap.Logger) error {
	if urlValue == "" {
		logger.Error("Empty URL", zap.String("URL Name", urlName))
		return utils.Error{Description: fmt.Sprintf(localutils.ErrEmptyURL, urlName)}
	}

	_, err := http.Get(urlValue)
	if err != nil {
		logger.Error("Endpoint unreachable", zap.String("URL", urlValue), zap.String("URL Name", urlName), zap.Error(err))
		return utils.Error{Description: fmt.Sprintf(localutils.ErrURLUnreachable, urlValue), BackendError: err.Error()}
	}

	return nil
}

// constructRIAASEndpoint ...
func constructRIAASEndpoint(region string) string {
	return strings.Replace(riaasEndpoint, "region", region, 1)
}

// constructPrivateRIAASEndpoint ...
func constructPrivateRIAASEndpoint(region string) string {
	return strings.Replace(privateRIAASEndpoint, "region", region, 1)
}

// constructContainerAPIRoute ...
func constructContainerAPIRoute(region string) string {
	return strings.Replace(containerAPIRoute, "region", region, 1)
}

// constructPrivateContainerAPIRoute ...
func constructPrivateContainerAPIRoute(region string) string {
	return strings.Replace(privateContainerAPIRoute, "region", region, 1)
}
