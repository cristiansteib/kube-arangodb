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

package reconcile

import (
	"context"
	"sort"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/pod"
)

func secretKeysToListWithPrefix(s *core.Secret) []string {
	return util.PrefixStringArray(secretKeysToList(s), "sha256:")
}

func secretKeysToList(s *core.Secret) []string {
	keys := make([]string, 0, len(s.Data))

	for key := range s.Data {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

// getCluster returns the cluster connection.
func getCluster(ctx context.Context, planCtx PlanBuilderContext) (driver.Cluster, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	c, err := planCtx.GetDatabaseClient(ctxChild)
	if err != nil {
		return nil, errors.WithStack(errors.Wrapf(err, "Unable to get database client"))
	}

	ctxChild, cancel = globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	cluster, err := c.Cluster(ctxChild)
	if err != nil {
		return nil, errors.WithStack(errors.Wrapf(err, "Unable to get cluster client"))
	}

	return cluster, nil
}

func ifPodUIDMismatch(m api.MemberStatus, a api.Action, i pod.Inspector) bool {
	ut, ok := a.GetParam(api.ParamPodUID)
	if !ok || ut == "" {
		return false
	}

	u := types.UID(ut)

	if m.PodName == "" {
		return false
	}

	p, ok := i.Pod().V1().GetSimple(m.PodName)
	if !ok {
		return true
	}

	return u != p.GetUID()
}

func withPredefinedMember(id string) api.MemberStatus {
	return api.MemberStatus{
		ID: id,
	}
}
