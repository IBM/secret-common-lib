module github.com/IBM/secret-common-lib

go 1.16

require (
	github.com/IBM/secret-utils-lib v0.0.0-20220405171320-9d9e137f5038
	go.uber.org/zap v1.20.0
	google.golang.org/grpc v1.27.1
	k8s.io/client-go v0.0.0-00010101000000-000000000000
)

replace (
	k8s.io/api => k8s.io/api v0.21.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.0
	k8s.io/client-go => k8s.io/client-go v0.21.0
)
