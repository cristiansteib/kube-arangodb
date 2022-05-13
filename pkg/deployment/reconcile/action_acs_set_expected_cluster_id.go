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
	"github.com/rs/zerolog"
	types "k8s.io/apimachinery/pkg/types"
)

func init() {
	registerAction(api.ActionTypeSetExpectedClusterID, newSetExpectedClusterIDAction, defaultTimeout)
}

const (
	ActionSetExpectedClusterID = "clusterID"
)

func newSetExpectedClusterIDAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &actionSetExpectedClusterID{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx)

	return a
}

type actionSetExpectedClusterID struct {
	// actionImpl implement timeout and member id functions
	actionImpl

	actionEmptyCheckProgress
}

func (a actionSetExpectedClusterID) Start(ctx context.Context) (bool, error) {
	log := a.log
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		log.Error().Msg("No such member")
		return true, nil
	}

	cid, ok := a.action.GetParam(ActionSetExpectedClusterID)
	if !ok {
		return true, nil
	}

	c, ok := a.actionCtx.ACS().Cluster(types.UID(cid))
	if !ok {
		return true, nil
	}

	if !c.Ready() {
		return true, nil
	}

	m.ExpectedClusterID = types.UID(cid)

	return true, a.actionCtx.UpdateMember(ctx, m)
}
