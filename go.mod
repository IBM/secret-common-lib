module github.com/IBM/secret-common-lib

go 1.16

require (
	github.com/IBM/secret-utils-lib v1.0.3-0.20220822070217-4b66786cdf09
	go.uber.org/zap v1.20.0
	google.golang.org/grpc v1.47.0
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
)

replace (
	k8s.io/api => k8s.io/api v0.21.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.0
	k8s.io/client-go => k8s.io/client-go v0.21.0
)
