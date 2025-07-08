package client

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"os"
)

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return ""
}

func getKubeconfig() (*api.Config, error) {
	kubeconfig := os.Getenv("KUBECONFIG")

	if kubeconfig == "" {
		// echo $HOME => Output: /Users/rahulxf
		// Mac specific things (will not work for window)
		if home := homeDir(); home != "" {
			kubeconfig = fmt.Sprintf("%s/.kube/config", home)
		}
	}
	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func GetClientSetWithContext(contextName string) (*kubernetes.Clientset, error) {
	var (
		config    *rest.Config
		clientset *kubernetes.Clientset
		err       error
	)
	kubeconfig := os.Getenv("KUBECONFIG")

	if kubeconfig == "" {
		// echo $HOME => Output: /Users/rahulxf
		// Mac specific things (will not work for window)
		if home := homeDir(); home != "" {
			kubeconfig = fmt.Sprintf("%s/.kube/config", home)
		}
	}

	if _, err := os.Stat(kubeconfig); err == nil {
		rawConfig, err := clientcmd.LoadFromFile(kubeconfig)
		if err != nil {
			return nil, err
		}
		if contextName == "" {
			contextName = rawConfig.CurrentContext
		}
		ctxContext := rawConfig.Contexts[contextName]
		if ctxContext == nil {
			return nil, fmt.Errorf("failed to find context '%s'", contextName)
		}
		clientConfig := clientcmd.NewDefaultClientConfig(
			*rawConfig,
			&clientcmd.ConfigOverrides{
				CurrentContext: contextName,
			},
		)
		config, err = clientConfig.ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create restconfig: %v", err)
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get in-cluster config: %v", err)
		}
	}
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}
	return clientset, nil
}
