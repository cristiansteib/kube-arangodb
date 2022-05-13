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

	endpointv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/endpoints/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	typedCore "k8s.io/client-go/kubernetes/typed/core/v1"
)

func (i *inspectorState) EndpointsModInterface() endpointv1.ModInterface {
	return endpointsModInterface{
		i: i,
	}
}

type endpointsModInterface struct {
	i *inspectorState
}

func (s endpointsModInterface) svcs() typedCore.EndpointsInterface {
	return s.i.client.Kubernetes().CoreV1().Endpoints(s.i.namespace)
}

func (s endpointsModInterface) Create(ctx context.Context, endpoint *core.Endpoints, opts meta.CreateOptions) (*core.Endpoints, error) {
	if svc, err := s.svcs().Create(ctx, endpoint, opts); err != nil {
		return svc, err
	} else {
		s.i.GetThrottles().Endpoints().Invalidate()
		return svc, err
	}
}

func (s endpointsModInterface) Update(ctx context.Context, endpoint *core.Endpoints, opts meta.UpdateOptions) (*core.Endpoints, error) {
	if svc, err := s.svcs().Update(ctx, endpoint, opts); err != nil {
		return svc, err
	} else {
		s.i.GetThrottles().Endpoints().Invalidate()
		return svc, err
	}
}

func (s endpointsModInterface) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts meta.PatchOptions, subresources ...string) (result *core.Endpoints, err error) {
	if svc, err := s.svcs().Patch(ctx, name, pt, data, opts, subresources...); err != nil {
		return svc, err
	} else {
		s.i.GetThrottles().Endpoints().Invalidate()
		return svc, err
	}
}

func (s endpointsModInterface) Delete(ctx context.Context, name string, opts meta.DeleteOptions) error {
	if err := s.svcs().Delete(ctx, name, opts); err != nil {
		return err
	} else {
		s.i.GetThrottles().Endpoints().Invalidate()
		return err
	}
}
