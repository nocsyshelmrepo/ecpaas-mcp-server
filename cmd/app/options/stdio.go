package options

import (
	cliflag "k8s.io/component-base/cli/flag"
)

type StdioOptions struct {
	KSConfig    string
	KSApiserver string
}

func NewStdioOptions() *StdioOptions {
	return &StdioOptions{}
}

func (o *StdioOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}
	gfs := fss.FlagSet("generic")
	gfs.StringVar(&o.KSApiserver, "ks-apiserver", o.KSApiserver, "the kubesphere apiserver address.")
	gfs.StringVar(&o.KSConfig, "ksconfig", o.KSConfig, "the KubeSphere config file to access KubeSphere.")

	return fss
}
