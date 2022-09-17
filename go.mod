module github.com/ssup2/kpexec

go 1.16

require (
	github.com/containerd/containerd v1.6.8
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.1 // indirect
	github.com/tidwall/gjson v1.14.3
	golang.org/x/net v0.0.0-20220722155237-a158d28d115b // indirect
	golang.org/x/sys v0.0.0-20220722155257-8c9f86f7a55f // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/api v0.22.5
	k8s.io/apimachinery v0.22.5
	k8s.io/client-go v0.22.5
)

replace github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
