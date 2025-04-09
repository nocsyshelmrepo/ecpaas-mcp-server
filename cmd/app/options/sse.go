package options

import cliflag "k8s.io/component-base/cli/flag"

const (
	defaultPort = "3000"
)

func NewSSEOptions() *SSEOptions {
	return &SSEOptions{
		Port: defaultPort,
	}
}

type SSEOptions struct {
	KSConfig    string
	KSApiserver string
	Port        string
}

func (o *SSEOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}
	gfs := fss.FlagSet("generic")
	gfs.StringVar(&o.KSApiserver, "ks-apiserver", o.KSApiserver, "the kubesphere apiserver address.")
	gfs.StringVar(&o.KSConfig, "ksconfig", o.KSConfig, "the KubeSphere config file to access KubeSphere.")
	gfs.StringVar(&o.KSConfig, "Port", o.KSConfig, "the sse port for ks-mcp-server. default is "+defaultPort)

	return fss
}
