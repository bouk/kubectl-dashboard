package main

import (
	restful "github.com/emicklei/go-restful"
	authApi "github.com/kubernetes/dashboard/src/app/backend/auth/api"
	backendClient "github.com/kubernetes/dashboard/src/app/backend/client"
	clientApi "github.com/kubernetes/dashboard/src/app/backend/client/api"
	pluginclientset "github.com/kubernetes/dashboard/src/app/backend/plugin/client/clientset/versioned"
	"github.com/kubernetes/dashboard/src/app/backend/resource/customresourcedefinition"
	v1 "k8s.io/api/authorization/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type clientManager struct {
	clientConfig        clientcmd.ClientConfig
	apiextensionsclient *apiextensionsclientset.Clientset
	pluginclient        *pluginclientset.Clientset
	client              *kubernetes.Clientset
}

func newClientManager(loadingRules *clientcmd.ClientConfigLoadingRules, overrides *clientcmd.ConfigOverrides) *clientManager {
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		overrides)

	config, _ := clientConfig.ClientConfig()
	apiextensionsclient, _ := apiextensionsclientset.NewForConfig(config)
	pluginclient, _ := pluginclientset.NewForConfig(config)
	client, _ := kubernetes.NewForConfig(config)

	return &clientManager{
		clientConfig:        clientConfig,
		apiextensionsclient: apiextensionsclient,
		pluginclient:        pluginclient,
		client:              client,
	}
}

func (c *clientManager) Client(req *restful.Request) (kubernetes.Interface, error) {
	return c.client, nil
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
func (c *clientManager) VerberClient(req *restful.Request, config *rest.Config) (clientApi.ResourceVerber, error) {
	client, err := c.Client(req)
	if err != nil {
		return nil, err
	}

	apiextensionsclient, err := c.APIExtensionsClient(req)
	if err != nil {
		return nil, err
	}

	pluginsclient, err := c.PluginClient(req)
	if err != nil {
		return nil, err
	}

	apiextensionsRestClient, err := customresourcedefinition.GetExtensionsAPIRestClient(apiextensionsclient)
	if err != nil {
		return nil, err
	}

	return backendClient.NewResourceVerber(
		client.CoreV1().RESTClient(),
		client.ExtensionsV1beta1().RESTClient(),
		client.AppsV1beta2().RESTClient(),
		client.BatchV1().RESTClient(),
		client.BatchV1beta1().RESTClient(),
		client.AutoscalingV1().RESTClient(),
		client.StorageV1().RESTClient(),
		client.RbacV1().RESTClient(),
		apiextensionsRestClient,
		pluginsclient.DashboardV1alpha1().RESTClient(),
		config,
	), nil
}

func (c *clientManager) SetTokenManager(manager authApi.TokenManager) {}

// APIExtensionsClient returns an API Extensions client. In case dashboard login is enabled and
// option to skip login page is disabled only secure client will be returned, otherwise insecure
// client will be used.
func (self *clientManager) APIExtensionsClient(req *restful.Request) (apiextensionsclientset.Interface, error) {
	return self.InsecureAPIExtensionsClient(), nil
}

// PluginClient returns a plugin client. In case dashboard login is enabled and
// option to skip login page is disabled only secure client will be returned, otherwise insecure
// client will be used.
func (self *clientManager) PluginClient(req *restful.Request) (pluginclientset.Interface, error) {
	return self.InsecurePluginClient(), nil
}

func (self *clientManager) InsecureAPIExtensionsClient() apiextensionsclientset.Interface {
	return self.apiextensionsclient
}

func (self *clientManager) InsecurePluginClient() pluginclientset.Interface {
	return self.pluginclient
}

var _ clientApi.ClientManager = &clientManager{}
