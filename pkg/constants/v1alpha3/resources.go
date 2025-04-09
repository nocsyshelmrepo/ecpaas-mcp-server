package v1alpha3

import "k8s.io/apimachinery/pkg/runtime/schema"

const (
	ResourcesGroupName = "resources.kubesphere.io"
	Version            = "v1alpha3"
)

var ResourcesGroupVersion = schema.GroupVersion{Group: ResourcesGroupName, Version: Version}
