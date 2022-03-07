# secret-common-lib
secret common lib library handles authentication for access to ibmcloud resources by implementing secret providers.

## secret providers

The following interface can be seen defined in the secret-utils-lib library
```
// SecretProviderInterface ...
type SecretProviderInterface interface {

	// GetIAMToken ...
	GetIAMToken(secret string, freshTokenRequired bool) (string, uint64, error)

	// GetDefaultIAMToken ...
	GetDefaultIAMToken(freshTokenRequired bool) (string, uint64, error)
}
```

## Pre requisites
- An environment variable IKS_ENABLED must to be set to true or false.
- A k8s secret must be present in the same namespace where the pod (the application in which this code is used) is deployed.
- The secrets format are present in secrets folder of secret-utils-lib library - ibm-cloud-credentials.yaml or storage-secret-store.yaml

## Description
- There are two types of secret providers - managed and unmanaged.
- A secret provider can be initialized like shown in the client code, in the client directory.
- If IKS_ENABLED is set to true, a managed secret provider is initialized which basically makes a call to another application (deployed in a different container), which has additional benefits compared unmanaged secret provider.
- If IKS_ENABLED is set to false, an unmanaged secret provider is initialized, which does not connect to any other application and supports very basic functionality.
- Both secret providers first look for ibm-cloud-credentials k8s secret, if it is not present, storage-secret-store k8s secret is considered.
- For more details - check concept doc.

### Managed secret provider
- Managed secret provider supports more functionalities than unmanaged secret provider.
- Secret-watcher (automatic update) - Once the secret provider is initialized successfully, a secret watcher is also initialized. If the secret is updated with a new API key or trusted profile, the same is automatically updated in the cache.
- Secret-watcher (fallback mechanism) - Provided that the secret provider is initialized with ibm-cloud-credentials, if the same is deleted, automatically secret sidecar switcher to storage-secret-store. If storage-secret-store is deleted, sidecar looks for ibm-cloud-credentials and uses the same. (This process happens without restarting the pod).
- Support for multiple secrets - In a pod, with one secret sidecar, multiple other containers can use it, for different profiles or api keys (A limit needs to be mentioned in the deployment).
- Deleting LRU secret - Given multiple applications are using secret sidecar, always the least recently used secret will not be stored in the cache. (Eg: If the limit for number of secrets is set to 3, and 4 different applications are using secret sidecar, with every call for fetching token, the least recently used secret is removed from cache). Always, the default secret fetched from ibm-cloud-credentials or storage-secret-store is always there in the cache.
- A TOKEN_EXPIRY_DIFF can be set at the time of deployment. Usage - Given that it is set to 20m, always managed secret provider makes sure, the token provided has atleast 20 minutes of validity.


### Unmanaged secret provider
- Unmanaged secret provider does not need a different container (like secret-sidecar in the case of managed secret provider).
- This is initialized as a part of the application which is using it.
- Does not support any secret watcher - with any update in secret watcher, pod needs to be restarted to pick the updated secret.
- Does not support multiple secrets.


