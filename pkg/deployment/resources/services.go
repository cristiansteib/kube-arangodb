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

package resources

import (
	"context"
	"strings"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/globals"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/acs/sutil"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"
	"github.com/arangodb/kube-arangodb/pkg/metrics"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	servicev1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/service/v1"
	"github.com/rs/zerolog"
	"k8s.io/apimachinery/pkg/types"
	"sort"
)

var (
	inspectedServicesCounters     = metrics.MustRegisterCounterVec(metricsComponent, "inspected_services", "Number of Service inspections per deployment", metrics.DeploymentName)
	inspectServicesDurationGauges = metrics.MustRegisterGaugeVec(metricsComponent, "inspect_services_duration", "Amount of time taken by a single inspection of all Services for a deployment (in sec)", metrics.DeploymentName)
)

type memberInfo struct {
	serving  bool
	ready    bool
	ip       string
	port     int32
	portName string
	cid      types.UID

	sg bool
}

type memberInfos map[string]memberInfo

func (m memberInfos) GetAllReadyForCluster(uid types.UID) []string {
	z := make([]string, 0, len(m))

	for k, v := range m {
		if v.cid == uid && v.ready && v.sg {
			z = append(z, k)
		}
	}

	sort.Strings(z)

	return z
}

func (m memberInfos) GetAllReady() []string {
	z := make([]string, 0, len(m))

	for k, v := range m {
		if v.ready && v.sg {
			z = append(z, k)
		}
	}

	sort.Strings(z)

	return z
}

// EnsureServices creates all services needed to service the deployment
func (r *Resources) EnsureServices(ctx context.Context, cachedStatus inspectorInterface.Inspector) error {
	log := r.log
	start := time.Now()
	apiObject := r.context.GetAPIObject()
	status, _ := r.context.GetStatus()
	deploymentName := apiObject.GetName()
	owner := apiObject.AsOwner()
	spec := r.context.GetSpec()
	defer metrics.SetDuration(inspectServicesDurationGauges.WithLabelValues(deploymentName), start)
	counterMetric := inspectedServicesCounters.WithLabelValues(deploymentName)

	reconcileRequired := k8sutil.NewReconcile(cachedStatus)

	mi := memberInfos{}

	// Get members details
	if err := status.Members.ForeachServerGroup(func(group api.ServerGroup, list api.MemberStatusList) error {
		for _, m := range list {
			var q memberInfo
			q.cid = m.ClusterID

			q.serving = m.Conditions.IsTrue(api.ConditionTypeServing)
			q.sg = group == spec.Mode.ServingGroup()

			if pod := m.PodName; pod != "" {
				item, ok := r.context.ACS().Cluster(m.ClusterID)
				if ok {
					cache := item.Cache()
					podStat, ok := cache.Pod().V1().GetSimple(pod)
					if ok {
						q.ip = podStat.Status.PodIP
						q.port = shared.ArangoPort
						q.portName = "server"
						q.ready = k8sutil.IsPodReady(podStat) && podStat.DeletionTimestamp == nil
					}
				}

				println("YYYYYYYYYYY", pod, q.ready)
			}

			mi[m.ID] = q
		}

		return nil
	}); err != nil {
		return err
	}

	// Ensure member services
	if err := status.Members.ForeachServerGroup(func(group api.ServerGroup, list api.MemberStatusList) error {
		var targetPort int32 = shared.ArangoPort

		switch group {
		case api.ServerGroupSyncMasters:
			targetPort = shared.ArangoSyncMasterPort
		case api.ServerGroupSyncWorkers:
			targetPort = shared.ArangoSyncWorkerPort
		}

		memberServiceEnsurer := func(item sutil.ACSItem) error {
			for _, m := range list {
				memberName := m.ArangoMemberName(r.context.GetAPIObject().GetName(), group)

				svcs := item.Cache().ServicesModInterface()

				member, ok := item.Cache().ArangoMember().V1().GetSimple(memberName)
				if !ok {
					return errors.Newf("Member %s not found", memberName)
				}

				if s, ok := item.Cache().Service().V1().GetSimple(member.GetName()); !ok {
					s = &core.Service{
						ObjectMeta: meta.ObjectMeta{
							Name:      member.GetName(),
							Namespace: member.GetNamespace(),
							OwnerReferences: []meta.OwnerReference{
								member.AsOwner(),
							},
						},
						Spec: core.ServiceSpec{
							Type: core.ServiceTypeClusterIP,
							Ports: []core.ServicePort{
								{
									Name:       "server",
									Protocol:   "TCP",
									Port:       shared.ArangoPort,
									TargetPort: intstr.IntOrString{IntVal: targetPort},
								},
							},
							PublishNotReadyAddresses: true,
						},
					}

					err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
						_, err := svcs.Create(ctxChild, s, meta.CreateOptions{})
						return err
					})
					if err != nil {
						if !k8sutil.IsConflict(err) {
							return err
						}
					}

					reconcileRequired.Required()
					continue
				} else {
					spec := s.Spec.DeepCopy()

					spec.Type = core.ServiceTypeClusterIP
					spec.Ports = []core.ServicePort{
						{
							Name:       "server",
							Protocol:   "TCP",
							Port:       shared.ArangoPort,
							TargetPort: intstr.IntOrString{IntVal: targetPort},
						},
					}
					spec.PublishNotReadyAddresses = true

					if item.IsOwnedBy(m.ClusterID) {
						spec.Selector = k8sutil.LabelsForMember(deploymentName, group.AsRole(), m.ID)
					} else {
						spec.Selector = nil
					}

					if !equality.Semantic.DeepDerivative(*spec, s.Spec) || (spec.Selector != nil && s.Spec.Selector == nil) || (spec.Selector == nil && s.Spec.Selector != nil) {
						s.Spec = *spec

						err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
							_, err := svcs.Update(ctxChild, s, meta.UpdateOptions{})
							return err
						})
						if err != nil {
							return err
						}

						reconcileRequired.Required()
						continue
					}

					if eread, err := item.Cache().Endpoints().V1(); err == nil {
						if !item.IsOwnedBy(m.ClusterID) {
							ends := item.Cache().EndpointsModInterface()

							if end, ok := eread.GetSimple(member.GetName()); !ok {
								// Create
								end = &core.Endpoints{
									ObjectMeta: meta.ObjectMeta{
										Name:         member.GetName(),
										GenerateName: "",
										Namespace:    item.Cache().Namespace(),
										OwnerReferences: []meta.OwnerReference{
											{
												APIVersion: inspector.ServiceVersionV1,
												Kind:       inspector.ServiceKind,
												Name:       member.GetName(),
												UID:        s.GetUID(),
											},
										},
									},
								}

								err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
									_, err := ends.Create(ctxChild, end, meta.CreateOptions{})
									return err
								})
								if err != nil {
									return err
								}

								reconcileRequired.Required()
								continue
							} else {
								var subsets *core.EndpointSubset
								if det, ok := mi[m.ID]; ok && det.ip != "" {
									subsets = &core.EndpointSubset{
										Addresses: []core.EndpointAddress{
											{
												IP: det.ip,
											},
										},
										Ports: []core.EndpointPort{
											{
												Name:     det.portName,
												Protocol: "TCP",
												Port:     det.port,
											},
										},
									}
								}

								changed := false
								if subsets == nil {
									// No IP behind
									if end.Subsets != nil {
										end.Subsets = nil
										changed = true
									}
								} else {
									if len(end.Subsets) != 1 || !equality.Semantic.DeepEqual(*subsets, end.Subsets[0]) {
										end.Subsets = []core.EndpointSubset{
											*subsets,
										}
										changed = true
									}
								}

								if changed {
									err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
										_, err := ends.Update(ctxChild, end, meta.UpdateOptions{})
										return err
									})
									if err != nil {
										return err
									}

									reconcileRequired.Required()
									continue
								}
							}
						}
					}
				}
			}

			return nil
		}

		return r.context.ACS().ForEachHealthyCluster(func(item sutil.ACSItem) error {
			if err := memberServiceEnsurer(item); err != nil {
				if item.IsMain() {
					return err
				}

				r.log.Warn().Str("cid", string(item.UID())).Err(err).Msg("Unable to ensure services")
			}

			return nil
		})
	}); err != nil {
		return err
	}

	svcs := cachedStatus.ServicesModInterface()

	// Headless service
	counterMetric.Inc()
	if _, exists := cachedStatus.Service().V1().GetSimple(k8sutil.CreateHeadlessServiceName(deploymentName)); !exists {
		ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
		defer cancel()
		svcName, newlyCreated, err := k8sutil.CreateHeadlessService(ctxChild, svcs, apiObject, owner)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to create headless service")
			return errors.WithStack(err)
		}
		if newlyCreated {
			log.Debug().Str("service", svcName).Msg("Created headless service")
		}
	}

	// Internal database client service
	single := spec.GetMode().HasSingleServers()
	counterMetric.Inc()

	role := "coordinator"
	if single {
		role = "single"
	}

	if err := r.context.ACS().ForEachHealthyCluster(func(item sutil.ACSItem) error {
		if svc, exists := item.Cache().Service().V1().GetSimple(k8sutil.CreateDatabaseClientServiceName(deploymentName)); !exists {
			svcs := item.Cache().ServicesModInterface()

			obj, err := item.Cache().GetCurrentArangoDeployment()
			if err != nil {
				return nil
			}

			ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
			defer cancel()
			svcName, newlyCreated, err := k8sutil.CreateDatabaseClientService(ctxChild, svcs, apiObject, single, obj.AsOwner())
			if err != nil {
				log.Debug().Err(err).Msg("Failed to create database client service")
				return errors.WithStack(err)
			}
			if newlyCreated {
				log.Debug().Str("service", svcName).Msg("Created database client service")
			}
		} else {
			return manageAdvertisedArangoDServiceEndpoints(ctx, item, deploymentName, role, svc, mi)
		}

		return nil
	}); err != nil {
		return err
	}

	// Database external access service
	eaServiceName := k8sutil.CreateDatabaseExternalAccessServiceName(deploymentName)

	if err := r.context.ACS().ForEachHealthyCluster(func(item sutil.ACSItem) error {
		obj, err := item.Cache().GetCurrentArangoDeployment()
		if err != nil {
			return nil
		}

		svcs := item.Cache().ServicesModInterface()

		if err := r.ensureExternalAccessServices(ctx, item.Cache(), svcs, eaServiceName, role, "database", shared.ArangoPort, false, spec.ExternalAccess, obj, log, false); err != nil {
			if item.IsMain() {
				return errors.WithStack(err)
			}

			r.log.Warn().Err(err).Msgf("Unable to create EQ service")
		}

		svc, ok := item.Cache().Service().V1().GetSimple(eaServiceName)
		if !ok {
			return nil
		}

		return manageAdvertisedArangoDServiceEndpoints(ctx, item, deploymentName, role, svc, mi)
	}); err != nil {
		return errors.WithStack(err)
	}

	if spec.Sync.IsEnabled() {
		// External (and internal) Sync master service
		counterMetric.Inc()
		eaServiceName := k8sutil.CreateSyncMasterClientServiceName(deploymentName)
		role := "syncmaster"
		if err := r.ensureExternalAccessServices(ctx, cachedStatus, svcs, eaServiceName, role, "sync", shared.ArangoSyncMasterPort, true, spec.Sync.ExternalAccess.ExternalAccessSpec, apiObject, log, true); err != nil {
			return errors.WithStack(err)
		}
		status, lastVersion := r.context.GetStatus()
		if status.SyncServiceName != eaServiceName {
			status.SyncServiceName = eaServiceName
			if err := r.context.UpdateStatus(ctx, status, lastVersion); err != nil {
				return errors.WithStack(err)
			}
		}
	}

	if spec.Metrics.IsEnabled() {
		ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
		defer cancel()
		name, _, err := k8sutil.CreateExporterService(ctxChild, cachedStatus, svcs, apiObject, apiObject.AsOwner())
		if err != nil {
			log.Debug().Err(err).Msgf("Failed to create %s exporter service", name)
			return errors.WithStack(err)
		}
		status, lastVersion := r.context.GetStatus()
		if status.ExporterServiceName != name {
			status.ExporterServiceName = name
			if err := r.context.UpdateStatus(ctx, status, lastVersion); err != nil {
				return errors.WithStack(err)
			}
		}
	}

	return reconcileRequired.Reconcile(ctx)
}

// EnsureServices creates all services needed to service the deployment
func (r *Resources) ensureExternalAccessServices(ctx context.Context, cachedStatus inspectorInterface.Inspector,
	svcs servicev1.ModInterface, eaServiceName, svcRole, title string, port int, noneIsClusterIP bool,
	spec api.ExternalAccessSpec, apiObject k8sutil.APIObject, log zerolog.Logger, attachSelector bool) error {
	// Database external access service
	createExternalAccessService := false
	deleteExternalAccessService := false
	eaServiceType := spec.GetType().AsServiceType() // Note: Type auto defaults to ServiceTypeLoadBalancer
	if existing, exists := cachedStatus.Service().V1().GetSimple(eaServiceName); exists {
		// External access service exists
		updateExternalAccessService := false
		loadBalancerIP := spec.GetLoadBalancerIP()
		loadBalancerSourceRanges := spec.LoadBalancerSourceRanges
		nodePort := spec.GetNodePort()
		if spec.GetType().IsNone() {
			if noneIsClusterIP {
				eaServiceType = core.ServiceTypeClusterIP
				if existing.Spec.Type != core.ServiceTypeClusterIP {
					deleteExternalAccessService = true // Remove the current and replace with proper one
					createExternalAccessService = true
				}
			} else {
				// Should not be there, remove it
				deleteExternalAccessService = true
			}
		} else if spec.GetType().IsAuto() {
			// Inspect existing service.
			if existing.Spec.Type == core.ServiceTypeLoadBalancer {
				// See if LoadBalancer has been configured & the service is "old enough"
				oldEnoughTimestamp := time.Now().Add(-1 * time.Minute) // How long does the load-balancer provisioner have to act.
				if len(existing.Status.LoadBalancer.Ingress) == 0 && existing.GetObjectMeta().GetCreationTimestamp().Time.Before(oldEnoughTimestamp) {
					log.Info().Str("service", eaServiceName).Msgf("LoadBalancerIP of %s external access service is not set, switching to NodePort", title)
					createExternalAccessService = true
					eaServiceType = core.ServiceTypeNodePort
					deleteExternalAccessService = true // Remove the LoadBalancer ex service, then add the NodePort one
				} else if existing.Spec.Type == core.ServiceTypeLoadBalancer && (loadBalancerIP != "" && existing.Spec.LoadBalancerIP != loadBalancerIP) {
					deleteExternalAccessService = true // LoadBalancerIP is wrong, remove the current and replace with proper one
					createExternalAccessService = true
				} else if existing.Spec.Type == core.ServiceTypeNodePort && len(existing.Spec.Ports) == 1 && (nodePort != 0 && existing.Spec.Ports[0].NodePort != int32(nodePort)) {
					deleteExternalAccessService = true // NodePort is wrong, remove the current and replace with proper one
					createExternalAccessService = true
				}
			}
		} else if spec.GetType().IsLoadBalancer() {
			if existing.Spec.Type != core.ServiceTypeLoadBalancer || (loadBalancerIP != "" && existing.Spec.LoadBalancerIP != loadBalancerIP) {
				deleteExternalAccessService = true // Remove the current and replace with proper one
				createExternalAccessService = true
			}
			if strings.Join(existing.Spec.LoadBalancerSourceRanges, ",") != strings.Join(loadBalancerSourceRanges, ",") {
				updateExternalAccessService = true
				existing.Spec.LoadBalancerSourceRanges = loadBalancerSourceRanges
			}
		} else if spec.GetType().IsNodePort() {
			if existing.Spec.Type != core.ServiceTypeNodePort || len(existing.Spec.Ports) != 1 || (nodePort != 0 && existing.Spec.Ports[0].NodePort != int32(nodePort)) {
				deleteExternalAccessService = true // Remove the current and replace with proper one
				createExternalAccessService = true
			}
		}
		if updateExternalAccessService && !createExternalAccessService && !deleteExternalAccessService {
			err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
				_, err := svcs.Update(ctxChild, existing, meta.UpdateOptions{})
				return err
			})
			if err != nil {
				log.Debug().Err(err).Msgf("Failed to update %s external access service", title)
				return errors.WithStack(err)
			}
		}
	} else {
		// External access service does not exist
		if !spec.GetType().IsNone() || noneIsClusterIP {
			createExternalAccessService = true
		}
	}

	if deleteExternalAccessService {
		log.Info().Str("service", eaServiceName).Msgf("Removing obsolete %s external access service", title)
		err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			return svcs.Delete(ctxChild, eaServiceName, meta.DeleteOptions{})
		})
		if err != nil {
			log.Debug().Err(err).Msgf("Failed to remove %s external access service", title)
			return errors.WithStack(err)
		}
	}
	if createExternalAccessService {
		// Let's create or update the database external access service
		nodePort := spec.GetNodePort()
		loadBalancerIP := spec.GetLoadBalancerIP()
		loadBalancerSourceRanges := spec.LoadBalancerSourceRanges
		ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
		defer cancel()
		_, newlyCreated, err := k8sutil.CreateExternalAccessService(ctxChild, svcs, eaServiceName, svcRole, apiObject, eaServiceType, port, nodePort, loadBalancerIP, loadBalancerSourceRanges, apiObject.AsOwner(), attachSelector)
		if err != nil {
			log.Debug().Err(err).Msgf("Failed to create %s external access service", title)
			return errors.WithStack(err)
		}
		if newlyCreated {
			log.Debug().Str("service", eaServiceName).Msgf("Created %s external access service", title)
		}
	}
	return nil
}

func manageAdvertisedArangoDServiceEndpoints(ctx context.Context, item sutil.ACSItem, deploymentName, role string, svc *core.Service, mi memberInfos) error {
	// Manage endpoints
	if z := mi.GetAllReadyForCluster(item.UID()); len(z) > 0 {
		if svc.Spec.Selector == nil {
			svc.Spec.Selector = k8sutil.LabelsForDeployment(deploymentName, role)
			if err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
				_, err := item.Cache().ServicesModInterface().Update(ctxChild, svc, meta.UpdateOptions{})
				return err
			}); err != nil {
				return err
			}

			return nil
		}
	} else {
		if svc.Spec.Selector != nil {
			svc.Spec.Selector = nil
			if err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
				_, err := item.Cache().ServicesModInterface().Update(ctxChild, svc, meta.UpdateOptions{})
				return err
			}); err != nil {
				return err
			}

			return nil
		}

		if ends, err := item.Cache().Endpoints().V1(); err == nil {
			end, ok := ends.GetSimple(svc.GetName())
			if !ok {
				end = &core.Endpoints{
					ObjectMeta: meta.ObjectMeta{
						Name:   svc.GetName(),
						Labels: svc.Labels,
						OwnerReferences: []meta.OwnerReference{
							{
								APIVersion: inspector.ServiceVersionV1,
								Kind:       inspector.ServiceKind,
								Name:       svc.GetName(),
								UID:        svc.GetUID(),
							},
						},
					},
				}
				if err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
					_, err := item.Cache().EndpointsModInterface().Create(ctxChild, end, meta.CreateOptions{})
					return err
				}); err != nil {
					return err
				}
				return nil
			}

			if n := mi.GetAllReady(); len(n) == 0 {
				// No endpoints, push as nil
				if end.Subsets != nil {
					end.Subsets = nil
					if err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
						_, err := item.Cache().EndpointsModInterface().Update(ctxChild, end, meta.UpdateOptions{})
						return err
					}); err != nil {
						return err
					}
					return nil
				}
			} else {
				ends := make([]core.EndpointSubset, len(n))

				for id := range ends {
					ep := mi[n[id]]

					ends[id] = core.EndpointSubset{
						Ports: []core.EndpointPort{
							{
								Name:     ep.portName,
								Port:     ep.port,
								Protocol: "TCP",
							},
						},
						Addresses: []core.EndpointAddress{
							{
								IP: ep.ip,
							},
						},
					}
				}

				if !equality.Semantic.DeepEqual(end.Subsets, ends) {
					end.Subsets = ends
					if err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
						_, err := item.Cache().EndpointsModInterface().Update(ctxChild, end, meta.UpdateOptions{})
						return err
					}); err != nil {
						return err
					}
					return nil
				}
			}
		}
	}

	return nil
}
