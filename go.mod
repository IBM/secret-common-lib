module github.com/IBM/secret-common-lib

go 1.16

require (
	github.com/IBM/secret-utils-lib v0.0.0-20220516185248-2c54b70560d2
	go.uber.org/zap v1.20.0
	google.golang.org/grpc v1.27.1
)

replace (
	k8s.io/api => k8s.io/api v0.21.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.0
	k8s.io/client-go => k8s.io/client-go v0.21.0
)
