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

package sync

import (
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AddOwner(toCache inspectorInterface.Inspector, from, to meta.Object) (added bool) {
	for _, owner := range from.GetOwnerReferences() {
		switch owner.Kind {
		case inspector.ArangoDeploymentKind:
			d, err := toCache.GetCurrentArangoDeployment()
			if err != nil {
				continue
			}
			to.SetOwnerReferences(append(to.GetOwnerReferences(), d.AsOwner()))
			added = true
		case inspector.ArangoMemberKind:
			d, ok := toCache.ArangoMember().V1().GetSimple(owner.Name)
			if !ok {
				continue
			}
			to.SetOwnerReferences(append(to.GetOwnerReferences(), d.AsOwner()))
			added = true
		}
	}
	return
}
