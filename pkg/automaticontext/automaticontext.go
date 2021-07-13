package automaticontext

// TODO should merge this package into the main

import (
	"os"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func Config(kubeconfigPath string) (*rest.Config, error) {
	var config *rest.Config
	var err error
	// check whether we are running out or inside a cluster
	// (ab)using the environment vars to do that
	// create the clientset

	// if you don't provide a kubeconfigPath, this will fallback to
	// InClusterConfig but with an annoying warning log
	if isInCluster() {
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, err
		}
	}
	return config, nil
}

func Client(kubeconfigPath string) (kubernetes.Interface, error) {
	config, err := Config(kubeconfigPath)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func CurrentNamespace() (string, error) {
	if isInCluster() {
		b, err := os.ReadFile("/run/secrets/kubernetes.io/serviceaccount/namespace")
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	clientCfg, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	if err != nil {
		return "", err
	}
	return clientCfg.Contexts[clientCfg.CurrentContext].Namespace, nil
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
func isInCluster() bool {
	for _, env := range os.Environ() {
		if strings.Contains(env, "KUBERNETES_SERVICE_HOST") {
			return true
		}
	}
	return false
}
