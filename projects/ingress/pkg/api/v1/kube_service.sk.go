// Code generated by solo-kit. DO NOT EDIT.

package v1

import (
	"sort"

	"github.com/gogo/protobuf/proto"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/hashutils"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// TODO: modify as needed to populate additional fields
func NewKubeService(namespace, name string) *KubeService {
	return &KubeService{
		Metadata: core.Metadata{
			Name:      name,
			Namespace: namespace,
		},
	}
}

func (r *KubeService) SetMetadata(meta core.Metadata) {
	r.Metadata = meta
}

func (r *KubeService) Hash() uint64 {
	metaCopy := r.GetMetadata()
	metaCopy.ResourceVersion = ""
	return hashutils.HashAll(
		metaCopy,
		r.KubeServiceSpec,
		r.KubeServiceStatus,
	)
}

type KubeServiceList []*KubeService
type ServicesByNamespace map[string]KubeServiceList

// namespace is optional, if left empty, names can collide if the list contains more than one with the same name
func (list KubeServiceList) Find(namespace, name string) (*KubeService, error) {
	for _, kubeService := range list {
		if kubeService.Metadata.Name == name {
			if namespace == "" || kubeService.Metadata.Namespace == namespace {
				return kubeService, nil
			}
		}
	}
	return nil, errors.Errorf("list did not find kubeService %v.%v", namespace, name)
}

func (list KubeServiceList) AsResources() resources.ResourceList {
	var ress resources.ResourceList
	for _, kubeService := range list {
		ress = append(ress, kubeService)
	}
	return ress
}

func (list KubeServiceList) Names() []string {
	var names []string
	for _, kubeService := range list {
		names = append(names, kubeService.Metadata.Name)
	}
	return names
}

func (list KubeServiceList) NamespacesDotNames() []string {
	var names []string
	for _, kubeService := range list {
		names = append(names, kubeService.Metadata.Namespace+"."+kubeService.Metadata.Name)
	}
	return names
}

func (list KubeServiceList) Sort() KubeServiceList {
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].Metadata.Less(list[j].Metadata)
	})
	return list
}

func (list KubeServiceList) Clone() KubeServiceList {
	var kubeServiceList KubeServiceList
	for _, kubeService := range list {
		kubeServiceList = append(kubeServiceList, proto.Clone(kubeService).(*KubeService))
	}
	return kubeServiceList
}

func (list KubeServiceList) Each(f func(element *KubeService)) {
	for _, kubeService := range list {
		f(kubeService)
	}
}

func (list KubeServiceList) AsInterfaces() []interface{} {
	var asInterfaces []interface{}
	list.Each(func(element *KubeService) {
		asInterfaces = append(asInterfaces, element)
	})
	return asInterfaces
}

func (byNamespace ServicesByNamespace) Add(kubeService ...*KubeService) {
	for _, item := range kubeService {
		byNamespace[item.Metadata.Namespace] = append(byNamespace[item.Metadata.Namespace], item)
	}
}

func (byNamespace ServicesByNamespace) Clear(namespace string) {
	delete(byNamespace, namespace)
}

func (byNamespace ServicesByNamespace) List() KubeServiceList {
	var list KubeServiceList
	for _, kubeServiceList := range byNamespace {
		list = append(list, kubeServiceList...)
	}
	return list.Sort()
}

func (byNamespace ServicesByNamespace) Clone() ServicesByNamespace {
	cloned := make(ServicesByNamespace)
	for ns, list := range byNamespace {
		cloned[ns] = list.Clone()
	}
	return cloned
}

var _ resources.Resource = &KubeService{}

// Kubernetes Adapter for KubeService

func (o *KubeService) GetObjectKind() schema.ObjectKind {
	t := KubeServiceCrd.TypeMeta()
	return &t
}

func (o *KubeService) DeepCopyObject() runtime.Object {
	return resources.Clone(o).(*KubeService)
}

var KubeServiceCrd = crd.NewCrd("ingress.solo.io",
	"services",
	"ingress.solo.io",
	"v1",
	"KubeService",
	"sv",
	false,
	&KubeService{})
