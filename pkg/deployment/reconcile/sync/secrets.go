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

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func EnsureSecretAction(ctx context.Context, from, to inspectorInterface.Inspector, name string) error {
	switch f, t, a := GetSecretActionType(from, to, name); a {
	case ActionCreate:
		n := core.Secret{
			ObjectMeta: meta.ObjectMeta{
				Name:      name,
				Namespace: to.Namespace(),
			},
		}

		getSecretFields(&n, f)

		if !AddOwner(to, f, &n) {
			return errors.Newf("Unable to find owner")
		}

		ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
		defer cancel()

		if _, err := to.Client().Kubernetes().CoreV1().Secrets(to.Namespace()).Create(ctxChild, &n, meta.CreateOptions{}); err != nil {
			return err
		}
	case ActionUpdate:
		getSecretFields(t, f)

		ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
		defer cancel()

		if _, err := to.Client().Kubernetes().CoreV1().Secrets(to.Namespace()).Update(ctxChild, t, meta.UpdateOptions{}); err != nil {
			return err
		}
	case ActionDelete:
		ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
		defer cancel()

		if err := to.Client().Kubernetes().CoreV1().Secrets(to.Namespace()).Delete(ctxChild, name, meta.DeleteOptions{}); err != nil {
			return err
		}
	}

	return nil
}

func GetSecretActionType(from, to inspectorInterface.Inspector, name string) (*core.Secret, *core.Secret, ActionType) {
	fromObj, fromOk := from.Secret().V1().GetSimple(name)
	toObj, toOk := to.Secret().V1().GetSimple(name)

	if fromOk && toOk {
		// Both exists
		if !AreSecretsEqual(fromObj, toObj) {
			return fromObj, toObj, ActionUpdate
		}
	} else if !fromOk && toOk {
		// Remote exists
		return nil, toObj, ActionDelete
	} else if fromOk && !toOk {
		// local exists
		return fromObj, nil, ActionCreate
	}

	return nil, nil, ActionNone
}

func AreSecretsEqual(a, b *core.Secret) bool {
	ac, err := getSecretSha(a)
	if err != nil {
		return false
	}
	bc, err := getSecretSha(b)
	if err != nil {
		return false
	}

	return ac == bc
}

func getSecretSha(s *core.Secret) (string, error) {
	q := core.Secret{}

	getSecretFields(&q, s)

	return util.SHA256FromJSON(q)
}

func getSecretFields(in, from *core.Secret) {
	in.Labels = from.Labels
	in.Data = from.Data
	in.Type = from.Type
}
