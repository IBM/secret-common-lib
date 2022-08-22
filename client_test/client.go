package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	sp "github.com/IBM/secret-common-lib/pkg/secret_provider"
	"github.com/IBM/secret-utils-lib/pkg/k8s_utils"
	"github.com/IBM/secret-utils-lib/pkg/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	logger := setUpLogger(false)
	//test1(logger)
	//test2(logger)
	//test3(logger)
	//test4(logger)
	//test5(logger)
	test6(logger)
}

// Possible test cases
// 1. ibm-cloud-credentials with no extra key and cloud-conf
// 2. ibm-cloud-credentials with key and cloud-conf
// 3. ibm-cloud-credentials with/without no extra key and cluster-info
// 4. ibm-cloud-credentials with/without no extra key and storage-secret-store
// 5. storage-secret-store with no extra key
// 6. storage-secret-store with no extra key and cluster-info
// 7. storage-secret-store with extra-key
// 8. storage-secret-store with extra-key and cluster-info
// 9. storage-secret-store with extra-key and cloud-conf

// test1 - ibm-cloud-credentials with no extra key and cloud-conf
func test1(logger *zap.Logger) {
	kc, _ := k8s_utils.FakeGetk8sClientSet(logger)
	cwd, err := os.Getwd()
	if err != nil {
		logger.Error("Error fetching current working directory")
		return
	}

	// Creating ibm-cloud-credentials secret
	secretFilePath := filepath.Join(cwd, "..", "test-fixtures/secrets/ibm-cloud-credentials/iam-cloud-provider.env")
	byteData, err := ioutil.ReadFile(secretFilePath)
	if err != nil {
		logger.Error("Error reading secret data", zap.Error(err))
		return
	}
	data := make(map[string][]byte)
	data[utils.CLOUD_PROVIDER_ENV] = byteData
	secret := new(v1.Secret)
	secret.Name = utils.IBMCLOUD_CREDENTIALS_SECRET
	secret.Data = data
	clientset := kc.GetClientSet()
	_, err = clientset.CoreV1().Secrets("kube-system").Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		logger.Error("Error creating secret", zap.Error(err))
		return
	}

	// Creating cloud-conf
	cloudConfFilePath := filepath.Join(cwd, "..", "test-fixtures/config-maps/cloud-conf.json")
	byteData, err = ioutil.ReadFile(cloudConfFilePath)
	if err != nil {
		logger.Error("Error reading cm", zap.Error(err))
		return
	}
	data2 := make(map[string]string)
	data2["cloud-conf.json"] = string(byteData)
	cm := new(v1.ConfigMap)
	cm.Data = data2
	cm.Name = "cloud-conf"
	_, err = clientset.CoreV1().ConfigMaps("kube-system").Create(context.TODO(), cm, metav1.CreateOptions{})
	if err != nil {
		logger.Error("Error creating config map", zap.Error(err))
		return
	}

	// Init unmanaged secret provider
	secretprovider, err := sp.InitUnmanagedSecretProvider(logger, kc, utils.VPC)
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

// test2 - ibm-cloud-credentials with key and cloud-conf
func test2(logger *zap.Logger) {
	kc, _ := k8s_utils.FakeGetk8sClientSet(logger)
	cwd, err := os.Getwd()
	if err != nil {
		logger.Error("Error fetching current working directory")
		return
	}

	// Creating ibm-cloud-credentials secret
	secretFilePath := filepath.Join(cwd, "..", "test-fixtures/secrets/ibm-cloud-credentials/new_key.env")
	byteData, err := ioutil.ReadFile(secretFilePath)
	if err != nil {
		logger.Error("Error reading secret data", zap.Error(err))
		return
	}
	data := make(map[string][]byte)
	data["new_key.env"] = byteData
	secret := new(v1.Secret)
	secret.Name = utils.IBMCLOUD_CREDENTIALS_SECRET
	secret.Data = data
	clientset := kc.GetClientSet()
	_, err = clientset.CoreV1().Secrets("kube-system").Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		logger.Error("Error creating secret", zap.Error(err))
		return
	}

	// Creating cloud-conf
	cloudConfFilePath := filepath.Join(cwd, "..", "test-fixtures/config-maps/cloud-conf.json")
	byteData, err = ioutil.ReadFile(cloudConfFilePath)
	if err != nil {
		logger.Error("Error reading cm", zap.Error(err))
		return
	}
	data2 := make(map[string]string)
	data2["cloud-conf.json"] = string(byteData)
	cm := new(v1.ConfigMap)
	cm.Data = data2
	cm.Name = "cloud-conf"
	_, err = clientset.CoreV1().ConfigMaps("kube-system").Create(context.TODO(), cm, metav1.CreateOptions{})
	if err != nil {
		logger.Error("Error creating config map", zap.Error(err))
		return
	}

	// Init unmanaged secret provider
	secretprovider, err := sp.InitUnmanagedSecretProvider(logger, kc, utils.VPC, "new_key.env")
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

// setUpLogger ...
func setUpLogger(managed bool) *zap.Logger {
	// Prepare a new logger
	atom := zap.NewAtomicLevel()
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	var secretProviderType string
	if managed {
		secretProviderType = "managed-secret-provider"
	} else {
		secretProviderType = "unmanaged-secret-provider"
	}
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		atom,
	), zap.AddCaller()).With(zap.String("name", "secret-provider")).With(zap.String("secret-provider-type", secretProviderType))

	atom.SetLevel(zap.InfoLevel)
	return logger
}

// test3 - ibm-cloud-credentials with no extra key and cluster-info
func test3(logger *zap.Logger) {
	kc, _ := k8s_utils.FakeGetk8sClientSet(logger)
	cwd, err := os.Getwd()
	if err != nil {
		logger.Error("Error fetching current working directory")
		return
	}

	// Creating ibm-cloud-credentials secret
	secretFilePath := filepath.Join(cwd, "..", "test-fixtures/secrets/ibm-cloud-credentials/new_key.env")
	byteData, err := ioutil.ReadFile(secretFilePath)
	if err != nil {
		logger.Error("Error reading secret data", zap.Error(err))
		return
	}
	data := make(map[string][]byte)
	data["new_key.env"] = byteData
	secret := new(v1.Secret)
	secret.Name = utils.IBMCLOUD_CREDENTIALS_SECRET
	secret.Data = data
	clientset := kc.GetClientSet()
	_, err = clientset.CoreV1().Secrets("kube-system").Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		logger.Error("Error creating secret", zap.Error(err))
		return
	}

	// Creating cluster-info
	cloudConfFilePath := filepath.Join(cwd, "..", "test-fixtures/config-maps/cluster-config.json")
	byteData, err = ioutil.ReadFile(cloudConfFilePath)
	if err != nil {
		logger.Error("Error reading cm", zap.Error(err))
		return
	}
	data2 := make(map[string]string)
	data2["cluster-config.json"] = string(byteData)
	cm := new(v1.ConfigMap)
	cm.Data = data2
	cm.Name = "cluster-info"
	_, err = clientset.CoreV1().ConfigMaps("kube-system").Create(context.TODO(), cm, metav1.CreateOptions{})
	if err != nil {
		logger.Error("Error creating config map", zap.Error(err))
		return
	}

	// Init unmanaged secret provider
	secretprovider, err := sp.InitUnmanagedSecretProvider(logger, kc, utils.VPC, "new_key.env")
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

// test4 - ibm-cloud-credentials with/without no extra key and storage-secret-store
func test4(logger *zap.Logger) {
	kc, _ := k8s_utils.FakeGetk8sClientSet(logger)
	cwd, err := os.Getwd()
	if err != nil {
		logger.Error("Error fetching current working directory")
		return
	}

	// Creating ibm-cloud-credentials secret
	secretFilePath := filepath.Join(cwd, "..", "test-fixtures/secrets/ibm-cloud-credentials/new_key.env")
	byteData, err := ioutil.ReadFile(secretFilePath)
	if err != nil {
		logger.Error("Error reading secret data", zap.Error(err))
		return
	}
	data := make(map[string][]byte)
	data["new_key.env"] = byteData
	secret := new(v1.Secret)
	secret.Name = utils.IBMCLOUD_CREDENTIALS_SECRET
	secret.Data = data
	clientset := kc.GetClientSet()
	_, err = clientset.CoreV1().Secrets("kube-system").Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		logger.Error("Error creating secret", zap.Error(err))
		return
	}

	// Creating storage-secret-store
	secretFilePath = filepath.Join(cwd, "..", "test-fixtures/secrets/storage-secret-store/slclient.toml")
	byteData, err = ioutil.ReadFile(secretFilePath)
	if err != nil {
		logger.Error("Error reading secret", zap.Error(err))
		return
	}
	data["slclient.toml"] = byteData
	secret.Name = utils.STORAGE_SECRET_STORE_SECRET
	secret.Data = data
	_, err = clientset.CoreV1().Secrets("kube-system").Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		logger.Error("Error creating secret", zap.Error(err))
		return
	}

	// Init unmanaged secret provider
	secretprovider, err := sp.InitUnmanagedSecretProvider(logger, kc, utils.VPC, "new_key.env")
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

// test5 - storage-secret-store with no extra key
func test5(logger *zap.Logger) {
	kc, _ := k8s_utils.FakeGetk8sClientSet(logger)
	cwd, err := os.Getwd()
	if err != nil {
		logger.Error("Error fetching current working directory")
		return
	}

	// Creating storage-secret-store
	secretFilePath := filepath.Join(cwd, "..", "test-fixtures/secrets/storage-secret-store/slclient.toml")
	byteData, err := ioutil.ReadFile(secretFilePath)
	if err != nil {
		logger.Error("Error reading secret", zap.Error(err))
		return
	}
	data := make(map[string][]byte)
	data["slclient.toml"] = byteData
	secret := new(v1.Secret)
	secret.Name = utils.STORAGE_SECRET_STORE_SECRET
	secret.Data = data
	clientset := kc.GetClientSet()
	_, err = clientset.CoreV1().Secrets("kube-system").Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		logger.Error("Error creating secret", zap.Error(err))
		return
	}

	// Init unmanaged secret provider
	secretprovider, err := sp.InitUnmanagedSecretProvider(logger, kc, utils.VPC)
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

// test6 - storage-secret-store extra key and cluster-info
func test6(logger *zap.Logger) {
	kc, _ := k8s_utils.FakeGetk8sClientSet(logger)
	cwd, err := os.Getwd()
	if err != nil {
		logger.Error("Error fetching current working directory")
		return
	}

	// Creating storage-secret-store
	secretFilePath := filepath.Join(cwd, "..", "test-fixtures/secrets/storage-secret-store/new_key.env")
	byteData, err := ioutil.ReadFile(secretFilePath)
	if err != nil {
		logger.Error("Error reading secret", zap.Error(err))
		return
	}
	data := make(map[string][]byte)
	data["new_key.env"] = byteData
	secret := new(v1.Secret)
	secret.Name = utils.STORAGE_SECRET_STORE_SECRET
	secret.Data = data
	clientset := kc.GetClientSet()
	_, err = clientset.CoreV1().Secrets("kube-system").Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		logger.Error("Error creating secret", zap.Error(err))
		return
	}

	// Creating cluster-info
	cloudConfFilePath := filepath.Join(cwd, "..", "test-fixtures/config-maps/cluster-config.json")
	byteData, err = ioutil.ReadFile(cloudConfFilePath)
	if err != nil {
		logger.Error("Error reading cm", zap.Error(err))
		return
	}
	data2 := make(map[string]string)
	data2["cluster-config.json"] = string(byteData)
	cm := new(v1.ConfigMap)
	cm.Data = data2
	cm.Name = "cluster-info"
	_, err = clientset.CoreV1().ConfigMaps("kube-system").Create(context.TODO(), cm, metav1.CreateOptions{})
	if err != nil {
		logger.Error("Error creating config map", zap.Error(err))
		return
	}

	// Init unmanaged secret provider
	secretprovider, err := sp.InitUnmanagedSecretProvider(logger, kc, utils.VPC, "new_key.env")
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
