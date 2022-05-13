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
	"context"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func EnsureArangoMemberAction(ctx context.Context, from, to inspectorInterface.Inspector, name string) error {
	switch f, _, a := GetArangoMemberActionType(from, to, name); a {
	case ActionCreate:
		n := api.ArangoMember{
			ObjectMeta: meta.ObjectMeta{
				Name:      name,
				Namespace: to.Namespace(),
			},
		}
		if !AddOwner(to, f, &n) {
			return errors.Newf("Unable to find owner")
		}

		ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
		defer cancel()

		if _, err := to.Client().Arango().DatabaseV1().ArangoMembers(to.Namespace()).Create(ctxChild, &n, meta.CreateOptions{}); err != nil {
			return err
		}
	case ActionDelete:
		ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
		defer cancel()

		if err := to.Client().Arango().DatabaseV1().ArangoMembers(to.Namespace()).Delete(ctxChild, name, meta.DeleteOptions{}); err != nil {
			return err
		}
	}

	return nil
}

func GetArangoMemberActionType(from, to inspectorInterface.Inspector, name string) (*api.ArangoMember, *api.ArangoMember, ActionType) {
	fromObj, fromOk := from.ArangoMember().V1().GetSimple(name)
	toObj, toOk := to.ArangoMember().V1().GetSimple(name)

	if fromOk && toOk {
		// Both exists
		return fromObj, toObj, ActionNone
	} else if !fromOk && toOk {
		// Remote exists
		return nil, toObj, ActionDelete
	} else if fromOk && !toOk {
		// local exists
		return fromObj, nil, ActionCreate
	}

	return nil, nil, ActionNone
}
