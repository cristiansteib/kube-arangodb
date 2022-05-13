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

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconcile/sync"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/rs/zerolog"
	types "k8s.io/apimachinery/pkg/types"
)

func init() {
	registerAction(api.ActionTypeResourceSync, newResourceSyncAction, defaultTimeout)
}

func newResourceSyncAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &actionResourceSync{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx)

	return a
}

type actionResourceSync struct {
	// actionImpl implement timeout and member id functions
	actionImpl

	actionEmptyCheckProgress
}

func (a actionResourceSync) Start(ctx context.Context) (bool, error) {
	fromID, ok := a.action.GetParam(ActionResourcesParamSource)
	if !ok {
		return false, errors.Newf("Unable to find param %s", ActionResourcesParamSource)
	}
	toID, ok := a.action.GetParam(ActionResourcesParamDestination)
	if !ok {
		return false, errors.Newf("Unable to find param %s", ActionResourcesParamDestination)
	}
	kind, ok := a.action.GetParam(ActionResourcesParamKind)
	if !ok {
		return false, errors.Newf("Unable to find param %s", ActionResourcesParamKind)
	}
	name, ok := a.action.GetParam(ActionResourcesParamName)
	if !ok {
		return false, errors.Newf("Unable to find param %s", ActionResourcesParamName)
	}

	from, ok := a.actionCtx.ACS().Cluster(types.UID(fromID))
	if !ok {
		return false, errors.Newf("Unable to find cluster %s", fromID)
	} else if !from.Ready() {
		return false, errors.Newf("Cluster %s not ready", fromID)
	}

	to, ok := a.actionCtx.ACS().Cluster(types.UID(toID))
	if !ok {
		return false, errors.Newf("Unable to find cluster %s", toID)
	} else if !to.Ready() {
		return false, errors.Newf("Cluster %s not ready", toID)
	}

	if from.Cache() == to.Cache() {
		return false, errors.Newf("Unable to sync within same cluster")
	}

	switch kind {
	case inspector.SecretKind:
		return true, sync.EnsureSecretAction(ctx, from.Cache(), to.Cache(), name)
	case inspector.ArangoMemberKind:
		return true, sync.EnsureArangoMemberAction(ctx, from.Cache(), to.Cache(), name)
	}

	return true, nil
}
