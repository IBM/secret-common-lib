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
	"io/ioutil"
	"net"
	"time"

	"github.com/IBM/secret-utils-lib/pkg/token"
	"github.com/IBM/secret-utils-lib/pkg/utils"
	sp "github.com/IBM/secret-utils-lib/secretprovider"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	endpoint = flag.String("sidecarEndpoint", "/csi/provider.sock", "Storage secret sidecar endpoint")
)

// ManagedSecretProvider ...
type ManagedSecretProvider struct {
	logger             *zap.Logger
	defaultSecretToken string
	watcher            *secretWatcher
}

type secretWatcher struct {
	watcher          watch.Interface
	secretname       string
	isUpdateRequired bool
}

// newManagedSecretProvider ...
func newManagedSecretProvider(logger *zap.Logger) (*ManagedSecretProvider, error) {
	logger.Info("Initializing managed secret provider, Checking if connection can be established to secret sidecar")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := grpc.DialContext(ctx, *endpoint, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithDialer(unixConnect))
	if err != nil {
		logger.Error("Error establishing grpc connection to secret sidecar", zap.Error(err))
		return nil, utils.Error{Description: "Error establishing grpc connection", BackendError: err.Error()}
	}

	watcher, err := initSecretWatcher(logger)
	if err != nil {
		logger.Error("Error initializing secret watcher")
		return nil, err
	}
	msp := &ManagedSecretProvider{logger: logger, watcher: watcher}
	go checkUpdate(msp)
	logger.Info("Initialized managed secret provider")
	return msp, nil
}

// GetDefaultIAMToken ...
func (msp *ManagedSecretProvider) GetDefaultIAMToken(freshTokenRequired bool) (string, uint64, error) {
	msp.logger.Info("Fetching IAM token for default secret")

	var tokenlifetime uint64

	if !msp.watcher.isUpdateRequired {
		// If the token in cache is valid, secret sidecar will not be called
		tokenlifetime, err := token.CheckTokenLifeTime(msp.defaultSecretToken)
		if err == nil {
			msp.logger.Info("Successfully fetched iam token")
			return msp.defaultSecretToken, tokenlifetime, nil
		}
	}

	// token in the cache isn't valid, hence sidecar needs to be called
	// Connecting to sidecar
	msp.logger.Info("Connecting to sidecar")
	conn, err := grpc.Dial(*endpoint, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithDialer(unixConnect))
	if err != nil {
		msp.logger.Error("Error establishing grpc connection to secret sidecar", zap.Error(err))
		return "", tokenlifetime, utils.Error{Description: "Error establishing grpc connection to secret sidecar", BackendError: err.Error()}
	}

	c := sp.NewSecretProviderClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	defer conn.Close()

	response, err := c.GetDefaultIAMToken(ctx, &sp.Request{IsFreshTokenRequired: true, ReadSecret: msp.watcher.isUpdateRequired})
	if err != nil {
		msp.logger.Error("Error fetching IAM token", zap.Error(err))
		return "", tokenlifetime, err
	}
	msp.watcher.isUpdateRequired = false
	msp.logger.Info("Successfully fetched IAM token for default secret")
	// Updating the cache with the new token received from sidecar
	msp.defaultSecretToken = response.Iamtoken
	return response.Iamtoken, response.Tokenlifetime, nil
}

// GetIAMToken ...
func (msp *ManagedSecretProvider) GetIAMToken(secret string, freshTokenRequired bool) (string, uint64, error) {
	msp.logger.Info("Fetching IAM token for the provided secret")

	var tokenlifetime uint64

	msp.logger.Info("Connecting to secret sidecar")
	conn, err := grpc.Dial(*endpoint, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithDialer(unixConnect))
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

	msp.logger.Info("Successfully fetched IAM token for the provided secret")
	return response.Iamtoken, response.Tokenlifetime, nil
}

// unixConnect ...
func unixConnect(addr string, t time.Duration) (net.Conn, error) {
	unixAddr, err := net.ResolveUnixAddr("unix", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUnix("unix", nil, unixAddr)
	return conn, err
}

// initSecretWatcher ...
func initSecretWatcher(logger *zap.Logger) (*secretWatcher, error) {
	logger.Info("Initializing secret watcher")

	// Fetching cluster config used to create k8s client
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		logger.Error("Error fetching in cluster config", zap.Error(err))
		return nil, utils.Error{Description: "Error initiliazing secret watcher", BackendError: err.Error()}
	}

	// Creating k8s client used to read secret
	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		logger.Error("Error creating k8s client", zap.Error(err))
		return nil, utils.Error{Description: "Error initiliazing secret watcher", BackendError: err.Error()}
	}

	// Reading the namespace in which the pod is deployed
	byteData, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		logger.Error("Error fetching namespace", zap.Error(err))
		return nil, utils.Error{Description: "Error initiliazing secret watcher - unable to read namespace", BackendError: err.Error()}
	}

	namespace := string(byteData)
	if namespace == "" {
		logger.Error("Unable to fetch namespace", zap.Error(err))
		return nil, utils.Error{Description: "Error initiliazing secret watcher - unable to read namespace"}
	}

	secretname := utils.IBMCLOUD_CREDENTIALS_SECRET
	watcher, err := clientset.CoreV1().Secrets(namespace).Watch(context.TODO(), metav1.SingleObject(metav1.ObjectMeta{Name: utils.IBMCLOUD_CREDENTIALS_SECRET, Namespace: namespace}))
	if err != nil {
		logger.Error("Error initializing watcher for ibm-cloud-credentials", zap.Error(err))
		logger.Info("Initializing watcher for storage-secret-store")
		watcher, err = clientset.CoreV1().Secrets(namespace).Watch(context.TODO(), metav1.SingleObject(metav1.ObjectMeta{Name: utils.STORAGE_SECRET_STORE_SECRET, Namespace: namespace}))
		if err != nil {
			logger.Error("Error initializing watcher for storage-secret-store", zap.Error(err))
			return nil, utils.Error{Description: "Error initiliazing secret watcher", BackendError: err.Error()}
		}
		secretname = utils.STORAGE_SECRET_STORE_SECRET
	}

	logger.Info("Initialized secret watcher")
	return &secretWatcher{watcher: watcher, secretname: secretname}, nil
}

// checkUpdate ...
func checkUpdate(msp *ManagedSecretProvider) {
	var err error
	for {
		event, open := <-msp.watcher.watcher.ResultChan()
		if open {
			if event.Type == watch.Deleted {
				msp.logger.Info("Secret is deleted", zap.String("secret-name", msp.watcher.secretname))
				msp.watcher, err = initSecretWatcher(msp.logger)
				if err != nil {
					msp.logger.Error("Error reinitializing secret, exiting watcher", zap.Error(err))
					return
				}
			} else if event.Type == watch.Modified {
				msp.logger.Info("Secret is modified", zap.String("secret-name", msp.watcher.secretname))
			}
			msp.watcher.isUpdateRequired = true
		}
	}
}
