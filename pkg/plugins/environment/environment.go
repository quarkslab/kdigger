package environment

import (
	"log"
	"os"
	"strings"

	"github.com/mtardy/kdigger/pkg/bucket"
)

const (
	bucketName        = "environment"
	bucketDescription = "Environment checks the presence of kubernetes related environment variables and shows them."

	KubernetesHostEnv = "KUBERNETES_SERVICE_HOST"
)

var bucketAliases = []string{"environments", "environ", "env"}

type EnvironmentBucket struct{}

func (n EnvironmentBucket) Run() (bucket.Results, error) {
	res := bucket.NewResults(bucketName)
	res.SetHeaders([]string{"name", "value"})
	for name, value := range kubeEnviron() {
		res.AddContent([]interface{}{name, value})
	}
	if IsTypicalKubernetesEnv() {
		res.SetComment("Typical Kubernetes API service env var was found, we might be running inside a pod.")
	} else {
		res.SetComment("Typical Kubernetes API service env var was not found, we might not be running inside a pod.")
	}
	return *res, nil
}

// Register registers a plugin
func Register(b *bucket.Buckets) {
	b.Register(bucketName, bucketAliases, bucketDescription, false, func(config bucket.Config) (bucket.Interface, error) {
		return NewEnvironmentBucket(config)
	})
}

func NewEnvironmentBucket(config bucket.Config) (*EnvironmentBucket, error) {
	return &EnvironmentBucket{}, nil
}

func environ() map[string]string {
	envs := make(map[string]string)
	for _, env := range os.Environ() {
		e := strings.SplitN(env, "=", 2)
		if len(e) < 2 {
			log.Panicln("environ strings should be in the form \"key=value\"")
		}
		envs[e[0]] = e[1]
	}
	return envs
}

func kubeEnviron() map[string]string {
	envs := make(map[string]string)
	for name, value := range environ() {
		if strings.Contains(name, "KUBE") {
			envs[name] = value
		}
	}
	return envs
}

// Since 1.13 you can disable service environment variables [1] thanks to the
// pull request to add the enableServiceLinks setting [2] but default
// kubernetes API server ones [3] are still always exported in container
// kubelet. This is why it's a rather stable way to determine quickly if you
// are running inside a kubernetes cluster. Of course it's possible to remove
// that variable from the environment to create a false negative and to add
// such variable in any environment to create a false positive but it's a bit
// unlikely.
//
// [1] https://kubernetes.io/docs/concepts/services-networking/service/#discovering-services
// [2] https://github.com/kubernetes/kubernetes/pull/68754
// [3] https://kubernetes.io/docs/concepts/services-networking/connect-applications-service/#environment-variables
func IsTypicalKubernetesEnv() bool {
	for name := range environ() {
		if name == KubernetesHostEnv {
			return true
		}
	}
	return false
}
