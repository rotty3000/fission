/*
Copyright The Fission Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1 "github.com/fission/fission/pkg/apis/core/v1"
	corev1 "github.com/fission/fission/pkg/generated/applyconfiguration/core/v1"
	typedcorev1 "github.com/fission/fission/pkg/generated/clientset/versioned/typed/core/v1"
	gentype "k8s.io/client-go/gentype"
)

// fakeKubernetesWatchTriggers implements KubernetesWatchTriggerInterface
type fakeKubernetesWatchTriggers struct {
	*gentype.FakeClientWithListAndApply[*v1.KubernetesWatchTrigger, *v1.KubernetesWatchTriggerList, *corev1.KubernetesWatchTriggerApplyConfiguration]
	Fake *FakeCoreV1
}

func newFakeKubernetesWatchTriggers(fake *FakeCoreV1, namespace string) typedcorev1.KubernetesWatchTriggerInterface {
	return &fakeKubernetesWatchTriggers{
		gentype.NewFakeClientWithListAndApply[*v1.KubernetesWatchTrigger, *v1.KubernetesWatchTriggerList, *corev1.KubernetesWatchTriggerApplyConfiguration](
			fake.Fake,
			namespace,
			v1.SchemeGroupVersion.WithResource("kuberneteswatchtriggers"),
			v1.SchemeGroupVersion.WithKind("KubernetesWatchTrigger"),
			func() *v1.KubernetesWatchTrigger { return &v1.KubernetesWatchTrigger{} },
			func() *v1.KubernetesWatchTriggerList { return &v1.KubernetesWatchTriggerList{} },
			func(dst, src *v1.KubernetesWatchTriggerList) { dst.ListMeta = src.ListMeta },
			func(list *v1.KubernetesWatchTriggerList) []*v1.KubernetesWatchTrigger {
				return gentype.ToPointerSlice(list.Items)
			},
			func(list *v1.KubernetesWatchTriggerList, items []*v1.KubernetesWatchTrigger) {
				list.Items = gentype.FromPointerSlice(items)
			},
		),
		fake,
	}
}
