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

	persistentvolumeclaimv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/persistentvolumeclaim/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	typedCore "k8s.io/client-go/kubernetes/typed/core/v1"
)

func (i *inspectorState) PersistentVolumeClaimsModInterface() persistentvolumeclaimv1.ModInterface {
	return persistentvolumeclaimsModInterface{
		i: i,
	}
}

type persistentvolumeclaimsModInterface struct {
	i *inspectorState
}

func (s persistentvolumeclaimsModInterface) svcs() typedCore.PersistentVolumeClaimInterface {
	return s.i.client.Kubernetes().CoreV1().PersistentVolumeClaims(s.i.namespace)
}

func (s persistentvolumeclaimsModInterface) Create(ctx context.Context, persistentvolumeclaim *core.PersistentVolumeClaim, opts meta.CreateOptions) (*core.PersistentVolumeClaim, error) {
	if svc, err := s.svcs().Create(ctx, persistentvolumeclaim, opts); err != nil {
		return svc, err
	} else {
		s.i.GetThrottles().PersistentVolumeClaim().Invalidate()
		return svc, err
	}
}

func (s persistentvolumeclaimsModInterface) Update(ctx context.Context, persistentvolumeclaim *core.PersistentVolumeClaim, opts meta.UpdateOptions) (*core.PersistentVolumeClaim, error) {
	if svc, err := s.svcs().Update(ctx, persistentvolumeclaim, opts); err != nil {
		return svc, err
	} else {
		s.i.GetThrottles().PersistentVolumeClaim().Invalidate()
		return svc, err
	}
}

func (s persistentvolumeclaimsModInterface) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts meta.PatchOptions, subresources ...string) (result *core.PersistentVolumeClaim, err error) {
	if svc, err := s.svcs().Patch(ctx, name, pt, data, opts, subresources...); err != nil {
		return svc, err
	} else {
		s.i.GetThrottles().PersistentVolumeClaim().Invalidate()
		return svc, err
	}
}

func (s persistentvolumeclaimsModInterface) Delete(ctx context.Context, name string, opts meta.DeleteOptions) error {
	if err := s.svcs().Delete(ctx, name, opts); err != nil {
		return err
	} else {
		s.i.GetThrottles().PersistentVolumeClaim().Invalidate()
		return err
	}
}
