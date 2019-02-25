package main

import (
	restful "github.com/emicklei/go-restful"
	authApi "github.com/kubernetes/dashboard/src/app/backend/auth/api"
	backendClient "github.com/kubernetes/dashboard/src/app/backend/client"
	clientApi "github.com/kubernetes/dashboard/src/app/backend/client/api"
	v1 "k8s.io/api/authorization/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type clientManager struct {
	clientConfig clientcmd.ClientConfig
}

func newClientManager(loadingRules *clientcmd.ClientConfigLoadingRules, overrides *clientcmd.ConfigOverrides) *clientManager {
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		overrides)

	return &clientManager{
		clientConfig: clientConfig,
	}
}

func (c *clientManager) Client(req *restful.Request) (kubernetes.Interface, error) {
	config, err := c.clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}
func (c *clientManager) InsecureClient() kubernetes.Interface {
	client, _ := c.Client(nil)
	return client
}
func (c *clientManager) CanI(req *restful.Request, ssar *v1.SelfSubjectAccessReview) bool {
	return true
}
func (c *clientManager) Config(req *restful.Request) (*rest.Config, error) {
	return c.clientConfig.ClientConfig()
}
func (c *clientManager) ClientCmdConfig(req *restful.Request) (clientcmd.ClientConfig, error) {
	return c.clientConfig, nil
}
func (c *clientManager) CSRFKey() string {
	return "123"
}
func (c *clientManager) HasAccess(authInfo api.AuthInfo) error {
	return nil
}
func (c *clientManager) VerberClient(req *restful.Request) (clientApi.ResourceVerber, error) {
	client, err := c.Client(req)
	if err != nil {
		return nil, err
	}

	return backendClient.NewResourceVerber(client.CoreV1().RESTClient(),
		client.ExtensionsV1beta1().RESTClient(), client.AppsV1beta2().RESTClient(),
		client.BatchV1().RESTClient(), client.BatchV1beta1().RESTClient(), client.AutoscalingV1().RESTClient(),
		client.StorageV1().RESTClient(), client.RbacV1().RESTClient()), nil
}

func (c *clientManager) SetTokenManager(manager authApi.TokenManager) {}

var _ clientApi.ClientManager = &clientManager{}
