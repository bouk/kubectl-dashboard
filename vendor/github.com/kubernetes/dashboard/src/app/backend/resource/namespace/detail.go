// Copyright 2017 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package namespace

import (
	"log"

	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/errors"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	"github.com/kubernetes/dashboard/src/app/backend/resource/event"
	"github.com/kubernetes/dashboard/src/app/backend/resource/limitrange"
	rq "github.com/kubernetes/dashboard/src/app/backend/resource/resourcequota"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sClient "k8s.io/client-go/kubernetes"
)

// NamespaceDetail is a presentation layer view of Kubernetes Namespace resource. This means it is Namespace plus
// additional augmented data we can get from other sources.
type NamespaceDetail struct {
	ObjectMeta api.ObjectMeta `json:"objectMeta"`
	TypeMeta   api.TypeMeta   `json:"typeMeta"`

	// Phase is the current lifecycle phase of the namespace.
	Phase v1.NamespacePhase `json:"phase"`

	// Events is list of events associated to the namespace.
	EventList common.EventList `json:"eventList"`

	// ResourceQuotaList is list of resource quotas associated to the namespace
	ResourceQuotaList *rq.ResourceQuotaDetailList `json:"resourceQuotaList"`

	// ResourceLimits is list of limit ranges associated to the namespace
	ResourceLimits []limitrange.LimitRangeItem `json:"resourceLimits"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

// GetNamespaceDetail gets namespace details.
func GetNamespaceDetail(client k8sClient.Interface, name string) (*NamespaceDetail, error) {
	log.Printf("Getting details of %s namespace\n", name)

	namespace, err := client.CoreV1().Namespaces().Get(name, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	events, err := event.GetNamespaceEvents(client, dataselect.DefaultDataSelect, namespace.Name)
	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}

	resourceQuotaList, err := getResourceQuotas(client, *namespace)
	nonCriticalErrors, criticalError = errors.AppendError(err, nonCriticalErrors)
	if criticalError != nil {
		return nil, criticalError
	}

	resourceLimits, err := getLimitRanges(client, *namespace)
	nonCriticalErrors, criticalError = errors.AppendError(err, nonCriticalErrors)
	if criticalError != nil {
		return nil, criticalError
	}

	namespaceDetails := toNamespaceDetail(*namespace, events, resourceQuotaList, resourceLimits, nonCriticalErrors)
	return &namespaceDetails, nil
}

func toNamespaceDetail(namespace v1.Namespace, events common.EventList, resourceQuotaList *rq.ResourceQuotaDetailList,
	resourceLimits []limitrange.LimitRangeItem, nonCriticalErrors []error) NamespaceDetail {

	return NamespaceDetail{
		ObjectMeta:        api.NewObjectMeta(namespace.ObjectMeta),
		TypeMeta:          api.NewTypeMeta(api.ResourceKindNamespace),
		Phase:             namespace.Status.Phase,
		EventList:         events,
		ResourceQuotaList: resourceQuotaList,
		ResourceLimits:    resourceLimits,
		Errors:            nonCriticalErrors,
	}
}

func getResourceQuotas(client k8sClient.Interface, namespace v1.Namespace) (*rq.ResourceQuotaDetailList, error) {
	list, err := client.CoreV1().ResourceQuotas(namespace.Name).List(api.ListEverything)

	result := &rq.ResourceQuotaDetailList{
		Items:    make([]rq.ResourceQuotaDetail, 0),
		ListMeta: api.ListMeta{TotalItems: len(list.Items)},
	}

	for _, item := range list.Items {
		detail := rq.ToResourceQuotaDetail(&item)
		result.Items = append(result.Items, *detail)
	}

	return result, err
}

func getLimitRanges(client k8sClient.Interface, namespace v1.Namespace) ([]limitrange.LimitRangeItem, error) {
	list, err := client.CoreV1().LimitRanges(namespace.Name).List(api.ListEverything)
	if err != nil {
		return nil, err
	}

	resourceLimits := make([]limitrange.LimitRangeItem, 0)
	for _, item := range list.Items {
		list := limitrange.ToLimitRanges(&item)
		resourceLimits = append(resourceLimits, list...)
	}

	return resourceLimits, nil
}
