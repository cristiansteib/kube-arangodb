//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//

package inspector

import (
	"context"

	podv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/pod/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	typedCore "k8s.io/client-go/kubernetes/typed/core/v1"
)

func (i *inspectorState) PodsModInterface() podv1.ModInterface {
	return podsModInterface{
		i: i,
	}
}

type podsModInterface struct {
	i *inspectorState
}

func (s podsModInterface) svcs() typedCore.PodInterface {
	return s.i.client.Kubernetes().CoreV1().Pods(s.i.namespace)
}

func (s podsModInterface) Create(ctx context.Context, pod *core.Pod, opts meta.CreateOptions) (*core.Pod, error) {
	if svc, err := s.svcs().Create(ctx, pod, opts); err != nil {
		return svc, err
	} else {
		s.i.GetThrottles().Pod().Invalidate()
		return svc, err
	}
}

func (s podsModInterface) Update(ctx context.Context, pod *core.Pod, opts meta.UpdateOptions) (*core.Pod, error) {
	if svc, err := s.svcs().Update(ctx, pod, opts); err != nil {
		return svc, err
	} else {
		s.i.GetThrottles().Pod().Invalidate()
		return svc, err
	}
}

func (s podsModInterface) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts meta.PatchOptions, subresources ...string) (result *core.Pod, err error) {
	if svc, err := s.svcs().Patch(ctx, name, pt, data, opts, subresources...); err != nil {
		return svc, err
	} else {
		s.i.GetThrottles().Pod().Invalidate()
		return svc, err
	}
}

func (s podsModInterface) Delete(ctx context.Context, name string, opts meta.DeleteOptions) error {
	if err := s.svcs().Delete(ctx, name, opts); err != nil {
		return err
	} else {
		s.i.GetThrottles().Pod().Invalidate()
		return err
	}
}
