# secret-common-lib
This library handles authentication for access to ibmcloud resources by implementing secret providers.

## secret providers

The following interface can be seen defined in the [secret-utils-lib](https://github.com/IBM/secret-utils-lib/blob/master/pkg/secret_provider/secret_provider_inf.go) library

```
// SecretProviderInterface ...
type SecretProviderInterface interface {

	// GetIAMToken ...
	GetIAMToken(reasonForCall, secret string, freshTokenRequired bool) (string, uint64, error)

	// GetDefaultIAMToken ...
	GetDefaultIAMToken(reasonForCall, freshTokenRequired bool) (string, uint64, error)

	// GetRIAASEndpoint ...
	GetRIAASEndpoint(readConfig bool) (string, error)

	// GetPrivateRIAASEndpoint ...
	GetPrivateRIAASEndpoint(readConfig bool) (string, error)

	// GetContainerAPIRoute ...
	GetContainerAPIRoute(readConfig bool) (string, error)

	// GetPrivateContainerAPIRoute ...
	GetPrivateContainerAPIRoute(readConfig bool) (string, error)

	// GetResourceGroupID ...
	GetResourceGroupID() string
}
```

## Pre requisites
- An environment variable IKS_ENABLED can to be set to true or false. If it is not set, it will be considered false.
- A k8s secret must be present in the same namespace where the pod (the application in which this code is used) is deployed.
- The format of the k8s secret should be either of one as mentioned in the following
1. ibm-cloud-credentials
```
apiVersion: v1
data:
  ibm-credentials.env: <base-64-encoded-value>
  <your-own-key>: <base-64-encoded-value>
kind: Secret
metadata:
  labels:
    addonmanager.kubernetes.io/mode: Reconcile
    app: ibm-vpc-block-csi-driver
    kubernetes.io/cluster-service: "true"
  name: ibm-cloud-credentials
  namespace: <namespace>
type: Opaque
```
The base-64-encoded-value can be obtained in either of the following ways:
```
echo -n "IBMCLOUD_AUTHTYPE=pod-identity
IBMCLOUD_PROFILEID=profile-id” | base64
```
OR
```
echo -n "IBMCLOUD_AUTHTYPE=iam
IBMCLOUD_APIKEY=api-key” | base64
```

2. storage-secret-store
```
apiVersion: v1
data:
  slclient.toml: <base-64-encoded-value>
  <your-own-key>: <base-64-encoded-value-of-api-key>
kind: Secret
metadata:
  labels:
    addonmanager.kubernetes.io/mode: Reconcile
    app: ibm-vpc-block-csi-driver
    kubernetes.io/cluster-service: "true"
  name: storage-secret-store
  namespace: kube-system
type: Opaque
```
The base-64-encoded-value can be obtained by base64 encoding the following:
```
[Bluemix]
  iam_url = "<iam_url>"
  iam_api_key = "<api-key>"
  containers_api_route = "<containers_api_route>"
  encryption = true
  containers_api_route_private = "<containers_api_route_private>"

[Softlayer]
  encryption = true
  softlayer_api_key = "<api-key>"
  softlayer_token_exchange_endpoint_url = "<softlayer_token_exchange_endpoint_url>"

[VPC]
  g2_token_exchange_endpoint_url = "<g2_token_exchange_endpoint_url>"
  g2_riaas_endpoint_url = "<g2_riaas_endpoint_url>"
  g2_riaas_endpoint_private_url = "<g2_riaas_endpoint_private_url>"
  g2_resource_group_id = "<g2_resource_group_id>"
  g2_api_key = "<api-key>"
  encryption = true
  iks_token_exchange_endpoint_private_url = "<iks_token_exchange_endpoint_private_url>"
```
- For the RIAAS endpoint, container api route, resource group ID, the library first looks for a k8s config map (in the same namespace where the application is deployed) which is of the following format
```
apiVersion: v1
data:
 cloud-conf.json: |
  {
   "region": "region",
   "riaas_endpoint": "",
   "riaas_private_endpoint": "",
   "containers_api_route": "",
   "containers_api_route_private": "",
   "token_exchange_url": "",
   "resource_group_id": ""
  }
kind: ConfigMap
metadata:
 name: cloud-conf
 namespace: <namespace>
```
- If the `cloud-conf` config map is not present, the same endpoints will be read from the k8s secret `storage-secret-store` whose format is shown above. 
- Note: As of now, in IKS/ROKS clusters, `cloud-conf` config map and `ibm-cloud-credentials` secret are not present by default, this needs to be created manually. This will be automated in the future. As of now, even if `cloud-conf` or `ibm-cloud-credentials` is not created, the library uses `storage-secret-store` for reading `api-key`,`endpoints` and `resource group id`, hence supporting backward compatibility. Going forward `storage-secret-store` will be completely deprecated.
- The following changes needs to be done in deployment file of the application that is using this library:
1. Under the volumes, the following entity needs to be added
```
volumes:
  - name: vault-token
    projected:
      sources:
      - serviceAccountToken:
          path: vault-token
          expirationSeconds: 600
```
2. If you are using unmanaged secret provider (more about the same, is described later in this READme), the volume mentioned above needs to be mounted inside container volumeMounts as shown below.
```
volumeMounts:
- mountPath: /var/run/secrets/tokens
  name: vault-token
```
3. If you are using managed secret provider, the following container config needs to be added in the deployment file
```
- name: storage-secret-sidecar
  image: icr.io/obs/armada-storage-secret:<release-tag>
  imagePullPolicy: Always
  args:
  - "--endpoint=$(ENDPOINT)"
  resources:
    limits:
      cpu: 40m
      memory: 80Mi
    requests:
      cpu: 10m
      memory: 20Mi
  env:
    - name: ENDPOINT
      value: "unix:/sidecardir/provider.sock"
    - name: TOKEN_EXPIRY_DIFF
      value: "40m"
    - name: PROFILE_CAPACITY
      value: "1"
  volumeMounts:
  - mountPath: /sidecardir
    name: socket-dir
  - mountPath: /var/run/secrets/tokens
    name: vault-token
```
Following needs to be added in application container which is using the managed secret provider
```
- name: <container name>
  image: <image>
  args:
    - "--sidecarEndpoint=$(SIDECAREP)"
  env:
    - name: SIDECAREP
      value: "/csi/provider.sock"
  ...
  ...
  ...
  volumeMounts:
    - name: socket-dir
      mountPath: /csi
  ...
  ...
```
Following needs to be added in volumes:
```
volumes:
- name: socket-dir
  emptyDir: {}
```


## Description
- There are two types of secret providers - managed and unmanaged.
- A secret provider can be initialized as shown in the [client](https://github.com/IBM/secret-common-lib/blob/Different-key-support/client/client.go) code.
- If IKS_ENABLED is set to true, a managed secret provider is initialized which basically makes a call to another application (deployed as a different container), which has additional benefits compared to unmanaged secret provider.
- If IKS_ENABLED is set to false, an unmanaged secret provider is initialized, which does not connect to any other application and supports very basic functionality.
- Both secret providers first look for `ibm-credentials.env` in `ibm-cloud-credentials` k8s secret, if it is not present, `slclient.toml` in `storage-secret-store` is considered.
- In the client code, you can pass an optional argument `your-own-key`. This option can be used, when you want to use a different key other than `ibm-credentials.env` in `ibm-cloud-credentials` or `slclient.toml` in `storage-secret-store`.
- Note: Going forward, since storage-secret-store will be completely deprecated, only ibm-cloud-credentials will be used.

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