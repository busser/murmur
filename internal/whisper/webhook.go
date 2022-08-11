package whisper

import "errors"

type K8sClient interface {
	GetConfigMapData(name, namespace string) (map[string]string, error)
	GetSecretData(name, namespace string) (map[string]string, error)
}

func Webhook(k8s K8sClient) error {
	return errors.New("not implemented")
}
