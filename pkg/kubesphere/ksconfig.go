package kubesphere

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/flowcontrol"
	kubespherescheme "kubesphere.io/client-go/kubesphere/scheme"

	"kubesphere.io/ks-mcp-server/pkg/constants"
)

var defaultKSConfig = clientcmdapi.Config{
	Clusters: map[string]*clientcmdapi.Cluster{
		"kubesphere": {
			Server: "https://kubesphere.com",
		},
	},
	Contexts: map[string]*clientcmdapi.Context{
		"kubesphere": {
			Cluster:  "kubesphere",
			AuthInfo: "kubesphere-admin",
		},
	},
	AuthInfos: map[string]*clientcmdapi.AuthInfo{
		"kubesphere-admin": {
			Username: "admin",
			Password: "P@88w0rd",
		},
	},
	CurrentContext: "kubesphere",
}

func LoadKSConfig(kubeconfig string) (config *restclient.Config, err error) {
	if kubeconfig == "" {
		return clientcmd.NewDefaultClientConfig(defaultKSConfig, nil).ClientConfig()
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		&clientcmd.ConfigOverrides{
			CurrentContext: constants.Getenv(constants.Kubesphere),
		}).ClientConfig()
}

type KSConfig struct {
	// httpClient is the http client to use for requests
	httpClient *http.Client

	// config is the rest.Config to talk to an apiserver
	config *restclient.Config

	// ksClientByType stores structured type metadata
	ksClientByType map[string]*ksClient

	mu sync.RWMutex
}

var _ restclient.Interface = &ksClient{}

func NewKSConfig(config *restclient.Config, apiserver string) (*KSConfig, error) {
	if config.BearerToken == "" && config.Password == "" {
		return nil, errors.New("[bearerToken] and [user, password] should not both be empty")
	}
	if apiserver != "" {
		config.Host = apiserver
	}
	httpclient, err := restclient.HTTPClientFor(config)
	if err != nil {
		return nil, err
	}

	return &KSConfig{
		httpClient:     httpclient,
		config:         config,
		ksClientByType: make(map[string]*ksClient),
	}, nil
}

func (c *KSConfig) RestClient(gv schema.GroupVersion, cluster string) (restclient.Interface, error) {
	var apiPath string
	switch gv.Group {
	// Core Kubernetes APIs like /api/v1/namespaces
	case "":
		apiPath = "api"
	// Another Kubernetes APIs
	case "apps", "batch", "networking.k8s.io", "apiextensions.k8s.io", "storage.k8s.io":
		apiPath = "apis"
	default:
		apiPath = "kapis"
	}
	if cluster != "" {
		apiPath = "clusters/" + cluster + "/kapis"
	}

	c.mu.RLock()
	r, known := c.ksClientByType[apiPath+"/"+gv.String()]
	c.mu.RUnlock()

	if known {
		return r, nil
	}

	// Initialize a new Client
	c.mu.Lock()
	defer c.mu.Unlock()
	restconfig := restclient.CopyConfig(c.config)
	restconfig.NegotiatedSerializer = kubespherescheme.Codecs
	restconfig.GroupVersion = &gv
	restconfig.APIPath = apiPath

	restclient, err := restclient.UnversionedRESTClientForConfigAndClient(restconfig, c.httpClient)
	if err != nil {
		return nil, err
	}
	ksclient := &ksClient{
		restclient: restclient,
		config:     restconfig,
		tokenCache: make(map[string]*token),
	}
	c.ksClientByType[apiPath+"/"+gv.String()] = ksclient

	return ksclient, nil
}

type ksClient struct {
	restclient restclient.Interface
	// config is the rest.Config to talk to an apiserver
	config *restclient.Config

	// store token cache to access kubesphere
	tokenCache map[string]*token
	mu         sync.RWMutex
}

type token struct {
	created      time.Time
	AccessToken  string `json:"access_Token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// APIVersion implements rest.Interface.
func (k *ksClient) APIVersion() schema.GroupVersion {
	return k.restclient.APIVersion()
}

// Delete implements rest.Interface.
func (k *ksClient) Delete() *restclient.Request {
	token, err := k.accessToken()
	if err != nil {
		return k.restclient.Delete()
	}

	return k.restclient.Delete().SetHeader("Authorization", "Bearer "+token)
}

// Get implements rest.Interface.
func (k *ksClient) Get() *restclient.Request {
	token, err := k.accessToken()
	if err != nil {
		return k.restclient.Get()
	}

	return k.restclient.Get().SetHeader("Authorization", "Bearer "+token)
}

// GetRateLimiter implements rest.Interface.
func (k *ksClient) GetRateLimiter() flowcontrol.RateLimiter {
	return k.restclient.GetRateLimiter()
}

// Patch implements rest.Interface.
func (k *ksClient) Patch(pt types.PatchType) *restclient.Request {
	token, err := k.accessToken()
	if err != nil {
		return k.restclient.Patch(pt)
	}

	return k.restclient.Patch(pt).SetHeader("Authorization", "Bearer "+token)
}

// Post implements rest.Interface.
func (k *ksClient) Post() *restclient.Request {
	token, err := k.accessToken()
	if err != nil {
		return k.restclient.Post()
	}

	return k.restclient.Post().SetHeader("Authorization", "Bearer "+token)
}

// Put implements rest.Interface.
func (k *ksClient) Put() *restclient.Request {
	token, err := k.accessToken()
	if err != nil {
		return k.restclient.Put()
	}

	return k.restclient.Put().SetHeader("Authorization", "Bearer "+token)
}

// Verb implements rest.Interface.
func (k *ksClient) Verb(verb string) *restclient.Request {
	token, err := k.accessToken()
	if err != nil {
		return k.restclient.Verb(verb)
	}

	return k.restclient.Verb(verb).SetHeader("Authorization", "Bearer "+token)
}

// AccessToken from tokenCache, if not exists in tokenCache, get it in ks-apiserver.
// refer to: https://kubesphere.io/zh/docs/v3.4/reference/api-docs/
func (k *ksClient) accessToken() (string, error) {
	if k.config.BearerToken != "" {
		return k.config.BearerToken, nil
	}
	k.mu.RLock()
	// `tokenReservedTime`: Specifies the buffer duration before a token expires prematurely.
	// This mechanism prevents scenarios where tokens might become invalid before reaching the server due to prolonged network delays during request processing.
	tokenReservedTime := time.Minute * 1
	ctk, ok := k.tokenCache[k.config.Username]
	if ok && ctk.created.Add(time.Second*time.Duration(ctk.ExpiresIn)).After(time.Now().Add(tokenReservedTime)) {
		return ctk.AccessToken, nil
	}
	k.mu.RUnlock()

	k.mu.Lock()
	defer k.mu.Unlock()
	// get from ks-apiserver
	data := url.Values{
		"grant_type":    []string{"password"},
		"client_id":     []string{"kubesphere"},
		"client_secret": []string{"kubesphere"},
		"username":      []string{k.config.Username},
		"password":      []string{k.config.Password},
	}
	path, err := url.JoinPath(k.config.Host, "oauth/token")
	if err != nil {
		return "", errors.Wrapf(err, "failed to join oauth/token path to %s", k.config.Host)
	}
	resp, err := http.Post(path, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return "", errors.Wrap(err, "failed to post oauth/token request")
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to ready http response body")
	}
	tk := &token{}
	if err := json.Unmarshal(respBody, tk); err != nil {
		return "", errors.Wrap(err, "faield to unmarshal http response body")
	}
	k.tokenCache[k.config.Username] = tk

	return tk.AccessToken, nil
}
