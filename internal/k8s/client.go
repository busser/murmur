package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	k8sConfig "sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Client struct {
	k8sClient kubernetes.Interface
}

func NewClient() (*Client, error) {
	kubeConfig, err := k8sConfig.GetConfig()
	if err != nil {
		return nil, err
	}

	k8sClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	return &Client{
		k8sClient: k8sClient,
	}, nil
}

func (c *Client) GetConfigMapData(ctx context.Context, name, namespace string) (map[string]string, error) {
	configMap, err := c.k8sClient.CoreV1().ConfigMaps(namespace).Get(ctx, name, corev1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting ConfigMap: %w", err)
	}

	return configMap.Data, nil
}

func (c *Client) GetSecretData(ctx context.Context, name, namespace string) (map[string]string, error) {
	secret, err := c.k8sClient.CoreV1().Secrets(namespace).Get(ctx, name, corev1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting Secret: %w", err)
	}

	return secret.StringData, nil
}
