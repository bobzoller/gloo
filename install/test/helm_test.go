package test

import (
	"encoding/json"
	"fmt"

	"github.com/solo-io/go-utils/installutils/kuberesource"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	v2 "github.com/solo-io/gloo/projects/gateway/pkg/api/v2"
	"github.com/solo-io/gloo/projects/gateway/pkg/defaults"
	skres "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd/solo.io/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	skprotoutils "github.com/solo-io/solo-kit/pkg/utils/protoutils"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/pointer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	. "github.com/solo-io/go-utils/manifesttestutils"
)

func GetPodNamespaceStats() v1.EnvVar {
	return v1.EnvVar{
		Name:  "START_STATS_SERVER",
		Value: "true",
	}
}

func GetPodNameEnvVar() v1.EnvVar {
	return v1.EnvVar{
		Name: "POD_NAME",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath: "metadata.name",
			},
		},
	}
}

func ConvertKubeResource(unst *unstructured.Unstructured, res resources.Resource) {
	byt, err := unst.MarshalJSON()
	Expect(err).NotTo(HaveOccurred())
	var skRes *skres.Resource
	Expect(json.Unmarshal(byt, &skRes)).NotTo(HaveOccurred())
	res.SetMetadata(kubeutils.FromKubeMeta(skRes.ObjectMeta))
	if withStatus, ok := res.(resources.InputResource); ok {
		resources.UpdateStatus(withStatus, func(status *core.Status) {
			*status = skRes.Status
		})
	}
	if skRes.Spec != nil {
		if err := skprotoutils.UnmarshalMap(*skRes.Spec, res); err != nil {
			Expect(err).NotTo(HaveOccurred())
		}
	}
}

var _ = Describe("Helm Test", func() {
	var (
		glooPorts = []v1.ContainerPort{
			{Name: "grpc-xds", ContainerPort: 9977, Protocol: "TCP"},
			{Name: "grpc-validation", ContainerPort: 9988, Protocol: "TCP"},
		}
		helmTestInstallId = "helm-unit-test-install-id"
	)

	Describe("gateway proxy extra annotations and crds", func() {
		var (
			globalLabels = map[string]string{
				"installationId": helmTestInstallId,
			}
			setGlobalLabels = func(testLabels map[string]string) {
				for k, v := range globalLabels {
					testLabels[k] = v
				}
			}
			selector         map[string]string
			testManifest     TestManifest
			statsAnnotations map[string]string
		)

		// adds the install ID into the provided helm flags- no need to provide it yourself
		prepareMakefile := func(helmFlags string) {
			testManifest = renderManifest(helmFlags + " --set installConfig.installationId=" + helmTestInstallId)
		}

		// helper for passing a values file
		prepareMakefileFromValuesFile := func(valuesFile string) {
			helmFlags := "--namespace " + namespace +
				" --set namespace.create=true" +
				" --set gatewayProxies.gatewayProxyV2.service.extraAnnotations.test=test" +
				" --values " + valuesFile
			prepareMakefile(helmFlags)
		}
		BeforeEach(func() {
			statsAnnotations = map[string]string{
				"prometheus.io/path":   "/metrics",
				"prometheus.io/port":   "9091",
				"prometheus.io/scrape": "true",
			}
		})

		Context("installation", func() {
			installIdLabel := "installationId"

			It("attaches a unique installation ID label to all top-level kubernetes resources if install ID is omitted", func() {
				testManifest = renderManifest("--namespace " + namespace)

				Expect(testManifest.NumResources()).NotTo(BeZero(), "Test manifest should have a nonzero number of resources")
				var uniqueInstallationId string
				testManifest.ExpectAll(func(resource *unstructured.Unstructured) {
					installationId, ok := resource.GetLabels()[installIdLabel]
					Expect(ok).To(BeTrue(), fmt.Sprintf("The installation ID key should be present, but is not present on %s %s in namespace %s",
						resource.GetKind(),
						resource.GetName(),
						resource.GetNamespace()))

					if uniqueInstallationId == "" {
						uniqueInstallationId = installationId
					}

					Expect(installationId).To(Equal(uniqueInstallationId),
						fmt.Sprintf("Should not have generated several installation IDs, but found %s on %s %s in namespace %s",
							installationId,
							resource.GetKind(),
							resource.GetNamespace(),
							resource.GetNamespace()))
				})

				Expect(uniqueInstallationId).NotTo(Equal(helmTestInstallId), "Make sure we didn't accidentally set our install ID to the helm test ID")

				haveNonzeroDeployments := false
				testManifest.SelectResources(func(resource *unstructured.Unstructured) bool {
					return resource.GetKind() == "Deployment"
				}).ExpectAll(func(unstructuredDeployment *unstructured.Unstructured) {
					haveNonzeroDeployments = true

					converted, err := kuberesource.ConvertUnstructured(unstructuredDeployment)
					Expect(err).NotTo(HaveOccurred(), "Should be able to convert from unstructured")
					deployment, ok := converted.(*appsv1.Deployment)
					Expect(ok).To(BeTrue(), "Should be castable to a deployment")

					Expect(len(deployment.Spec.Template.Labels)).NotTo(BeZero(), "The deployment's pod spec had no labels")

					podInstallId, ok := deployment.Spec.Template.Labels[installIdLabel]
					Expect(ok).To(BeTrue(), "Should have the install id label set")
					Expect(podInstallId).To(Equal(uniqueInstallationId), "Pods from deployments should have the same install IDs as everything else")
				})

				Expect(haveNonzeroDeployments).To(BeTrue())
			})

			It("can assign a custom installation ID", func() {
				installId := "custom-install-id"
				testManifest = renderManifest("--namespace " + namespace + " --set installConfig.installationId=" + installId)

				Expect(testManifest.NumResources()).NotTo(BeZero())
				testManifest.ExpectAll(func(resource *unstructured.Unstructured) {
					installationId, ok := resource.GetLabels()[installIdLabel]
					Expect(ok).To(BeTrue(), fmt.Sprintf("The installation ID key should be present, but is not present on %s %s in namespace %s",
						resource.GetKind(),
						resource.GetName(),
						resource.GetNamespace()))

					Expect(installationId).To(Equal(installId),
						fmt.Sprintf("Should not have generated several installation IDs, but found %s on %s %s in namespace %s",
							installationId,
							resource.GetKind(),
							resource.GetNamespace(),
							resource.GetNamespace()))
				})
			})
		})

		Context("gateway", func() {
			var labels map[string]string
			BeforeEach(func() {
				labels = map[string]string{
					"app":              "gloo",
					"gloo":             "gateway-proxy",
					"gateway-proxy-id": "gateway-proxy-v2",
				}
				setGlobalLabels(labels)
				selector = map[string]string{
					"gateway-proxy":    "live",
					"gateway-proxy-id": "gateway-proxy-v2",
				}
				setGlobalLabels(selector)
			})

			It("has a namespace", func() {
				helmFlags := "--namespace " + namespace + " --set namespace.create=true  --set gatewayProxies.gatewayProxyV2.service.extraAnnotations.test=test"
				prepareMakefile(helmFlags)
				rb := ResourceBuilder{
					Namespace: namespace,
					Name:      defaults.GatewayProxyName,
					Labels:    labels,
					Service: ServiceSpec{
						Ports: []PortSpec{
							{
								Name: "http",
								Port: 80,
							},
							{
								Name: "https",
								Port: 443,
							},
						},
					},
				}
				svc := rb.GetService()
				svc.Spec.Selector = selector
				svc.Spec.Type = v1.ServiceTypeLoadBalancer
				svc.Spec.Ports[0].TargetPort = intstr.FromInt(8080)
				svc.Spec.Ports[1].TargetPort = intstr.FromInt(8443)
				svc.Annotations = map[string]string{"test": "test"}
				testManifest.ExpectService(svc)
			})

			Context("access logging service", func() {
				var (
					accessLoggerName          = "gateway-proxy-v2-access-logger"
					gatewayProxyConfigMapName = "gateway-proxy-v2-envoy-config"
				)
				BeforeEach(func() {
					labels = map[string]string{
						"app":  "gloo",
						"gloo": "gateway-proxy-v2-access-logger",
					}
					setGlobalLabels(labels)
				})

				It("can create an access logging deployment/service", func() {
					prepareMakefileFromValuesFile("install/test/val_access_logger.yaml")
					container := GetQuayContainerSpec("access-logger", version, GetPodNamespaceEnvVar(), GetPodNameEnvVar(),
						v1.EnvVar{
							Name:  "SERVICE_NAME",
							Value: "AccessLog",
						},
						v1.EnvVar{
							Name:  "SERVER_PORT",
							Value: "8083",
						},
					)
					container.PullPolicy = "Always"
					rb := &ResourceBuilder{
						Namespace:  namespace,
						Name:       accessLoggerName,
						Labels:     labels,
						Containers: []ContainerSpec{container},
						Service: ServiceSpec{
							Ports: []PortSpec{
								{
									Name: "http",
									Port: 8083,
								},
							},
						},
					}
					svc := rb.GetService()
					svc.Spec.Selector = labels
					svc.Spec.Type = ""
					svc.Spec.Ports[0].TargetPort = intstr.FromInt(8083)
					dep := rb.GetDeploymentAppsv1()
					dep.Spec.Template.Spec.Containers[0].Ports = []v1.ContainerPort{
						{Name: "http", ContainerPort: 8083, Protocol: "TCP"},
					}
					dep.Spec.Template.Spec.ServiceAccountName = "gateway-proxy"
					testManifest.ExpectDeploymentAppsV1(dep)
					testManifest.ExpectService(svc)
				})

				It("has a proxy with access logging cluster", func() {
					prepareMakefileFromValuesFile("install/test/val_access_logger.yaml")
					proxySpec := make(map[string]string)
					labels = map[string]string{
						"gloo":             "gateway-proxy",
						"app":              "gloo",
						"gateway-proxy-id": "gateway-proxy-v2",
					}
					setGlobalLabels(labels)
					proxySpec["envoy.yaml"] = confWithAccessLogger
					cmRb := ResourceBuilder{
						Namespace: namespace,
						Name:      gatewayProxyConfigMapName,
						Labels:    labels,
						Data:      proxySpec,
					}
					proxy := cmRb.GetConfigMap()
					testManifest.ExpectConfigMapWithYamlData(proxy)
				})
			})

			Context("default gateways", func() {

				var (
					proxyNames = []string{defaults.GatewayProxyName}
				)

				It("renders with http/https gateways by default", func() {
					prepareMakefile("--namespace " + namespace)
					gatewayUns := testManifest.ExpectCustomResource("Gateway", namespace, defaults.GatewayProxyName)
					var gateway1 v2.Gateway
					ConvertKubeResource(gatewayUns, &gateway1)
					Expect(gateway1.Ssl).To(BeFalse())
					Expect(gateway1.BindPort).To(Equal(uint32(8080)))
					Expect(gateway1.ProxyNames).To(Equal(proxyNames))
					Expect(gateway1.UseProxyProto).To(Equal(&types.BoolValue{Value: false}))
					Expect(gateway1.BindAddress).To(Equal(defaults.GatewayBindAddress))
					gatewayUns = testManifest.ExpectCustomResource("Gateway", namespace, defaults.GatewayProxyName+"-ssl")
					ConvertKubeResource(gatewayUns, &gateway1)
					Expect(gateway1.Ssl).To(BeTrue())
					Expect(gateway1.BindPort).To(Equal(uint32(8443)))
					Expect(gateway1.ProxyNames).To(Equal(proxyNames))
					Expect(gateway1.UseProxyProto).To(Equal(&types.BoolValue{Value: false}))
					Expect(gateway1.BindAddress).To(Equal(defaults.GatewayBindAddress))
				})

				It("can disable rendering http/https gateways", func() {
					prepareMakefile("--namespace " + namespace + " --set namespace.create=true  --set gatewayProxies.gatewayProxyV2.gatewaySettings.disableGeneratedGateways=true")
					testManifest.ExpectUnstructured("Gateway", namespace, defaults.GatewayProxyName).To(BeNil())
					testManifest.ExpectUnstructured("Gateway", namespace, defaults.GatewayProxyName+"-ssl").To(BeNil())
				})

				It("can render with custom listener yaml", func() {
					newGatewayProxyName := "test-name"
					vsList := []core.ResourceRef{
						{
							Name:      "one",
							Namespace: "one",
						},
					}
					prepareMakefileFromValuesFile("install/test/val_custom_gateways.yaml")
					for _, name := range []string{newGatewayProxyName, defaults.GatewayProxyName} {
						name := name
						gatewayUns := testManifest.ExpectCustomResource("Gateway", namespace, name)
						var gateway1 v2.Gateway
						ConvertKubeResource(gatewayUns, &gateway1)
						Expect(gateway1.UseProxyProto).To(Equal(&types.BoolValue{
							Value: true,
						}))
						httpGateway := gateway1.GetHttpGateway()
						Expect(httpGateway).NotTo(BeNil())
						Expect(httpGateway.VirtualServices).To(Equal(vsList))
						gatewayUns = testManifest.ExpectCustomResource("Gateway", namespace, name+"-ssl")
						ConvertKubeResource(gatewayUns, &gateway1)
						Expect(gateway1.UseProxyProto).To(Equal(&types.BoolValue{
							Value: true,
						}))
						Expect(httpGateway.VirtualServices).To(Equal(vsList))
					}

				})
			})

			Context("gateway conversion job", func() {
				var (
					job *batchv1.Job
				)
				BeforeEach(func() {
					jobLabels := map[string]string{
						"app":  "gloo",
						"gloo": "gateway",
					}
					setGlobalLabels(jobLabels)
					job = &batchv1.Job{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Job",
							APIVersion: "batch/v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Labels:    jobLabels,
							Name:      "gateway-conversion",
							Namespace: namespace,
						},
						Spec: batchv1.JobSpec{
							Template: v1.PodTemplateSpec{
								ObjectMeta: metav1.ObjectMeta{
									Labels: map[string]string{
										"gloo": "gateway",
									},
								},
								Spec: v1.PodSpec{
									RestartPolicy:      v1.RestartPolicyNever,
									ServiceAccountName: "gateway",
									Containers: []v1.Container{
										{
											Name:            "gateway-conversion",
											Image:           "quay.io/solo-io/gateway-conversion:" + version,
											ImagePullPolicy: v1.PullAlways,
											Env: []v1.EnvVar{
												GetPodNamespaceEnvVar(),
											},
										},
									},
								},
							},
						},
					}
				})

				It("doesn't creates a deployment", func() {
					prepareMakefile("--namespace " + namespace + " --set gateway.upgrade=false")
					testManifest.Expect(job.Kind, job.Namespace, job.Name).To(BeNil())
				})

				It("creates a deployment", func() {
					prepareMakefile("--namespace " + namespace + " --set gateway.upgrade=true")
					testManifest.Expect(job.Kind, job.Namespace, job.Name).To(BeEquivalentTo(job))
				})
			})

			Context("gateway-proxy service", func() {
				var gatewayProxyService *v1.Service

				BeforeEach(func() {
					serviceLabels := map[string]string{
						"app":              "gloo",
						"gloo":             "gateway-proxy",
						"gateway-proxy-id": "gateway-proxy-v2",
					}
					setGlobalLabels(serviceLabels)
					rb := ResourceBuilder{
						Namespace: namespace,
						Name:      "gateway-proxy-v2",
						Args:      nil,
						Labels:    serviceLabels,
					}
					gatewayProxyService = rb.GetService()
					selectorLabels := map[string]string{
						"gateway-proxy-id": "gateway-proxy-v2",
						"gateway-proxy":    "live",
					}
					setGlobalLabels(selectorLabels)
					gatewayProxyService.Spec.Selector = selectorLabels
					gatewayProxyService.Spec.Ports = []v1.ServicePort{
						{
							Name:       "http",
							Protocol:   "TCP",
							Port:       80,
							TargetPort: intstr.IntOrString{IntVal: 8080},
						},
						{
							Name:       "https",
							Protocol:   "TCP",
							Port:       443,
							TargetPort: intstr.IntOrString{IntVal: 8443},
						},
					}
					gatewayProxyService.Spec.Type = v1.ServiceTypeLoadBalancer
				})

				It("sets extra annotations", func() {
					gatewayProxyService.ObjectMeta.Annotations = map[string]string{"foo": "bar", "bar": "baz"}
					prepareMakefile("--namespace " + namespace +
						" --set gatewayProxies.gatewayProxyV2.service.extraAnnotations.foo=bar" +
						" --set gatewayProxies.gatewayProxyV2.service.extraAnnotations.bar=baz")
					testManifest.ExpectService(gatewayProxyService)
				})

				It("sets external traffic policy", func() {
					gatewayProxyService.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeLocal
					prepareMakefile("--namespace " + namespace +
						" --set gatewayProxies.gatewayProxyV2.service.externalTrafficPolicy=Local")
					testManifest.ExpectService(gatewayProxyService)
				})

				It("sets cluster IP", func() {
					gatewayProxyService.Spec.Type = v1.ServiceTypeClusterIP
					gatewayProxyService.Spec.ClusterIP = "test-ip"
					prepareMakefile("--namespace " + namespace +
						" --set gatewayProxies.gatewayProxyV2.service.type=ClusterIP" +
						" --set gatewayProxies.gatewayProxyV2.service.clusterIP=test-ip")
					testManifest.ExpectService(gatewayProxyService)
				})

				It("sets load balancer IP", func() {
					gatewayProxyService.Spec.Type = v1.ServiceTypeLoadBalancer
					gatewayProxyService.Spec.LoadBalancerIP = "test-lb-ip"
					prepareMakefile("--namespace " + namespace +
						" --set gatewayProxies.gatewayProxyV2.service.type=LoadBalancer" +
						" --set gatewayProxies.gatewayProxyV2.service.loadBalancerIP=test-lb-ip")
					testManifest.ExpectService(gatewayProxyService)
				})
			})

			Context("gateway-proxy deployment", func() {
				var (
					gatewayProxyDeployment *appsv1.Deployment
				)

				BeforeEach(func() {
					selector = map[string]string{
						"gloo":             "gateway-proxy",
						"gateway-proxy-id": "gateway-proxy-v2",
					}
					setGlobalLabels(selector)
					podLabels := map[string]string{
						"gloo":             "gateway-proxy",
						"gateway-proxy":    "live",
						"gateway-proxy-id": "gateway-proxy-v2",
					}
					setGlobalLabels(podLabels)
					podname := v1.EnvVar{
						Name: "POD_NAME",
						ValueFrom: &v1.EnvVarSource{
							FieldRef: &v1.ObjectFieldSelector{
								FieldPath: "metadata.name",
							},
						},
					}
					container := GetQuayContainerSpec("gloo-envoy-wrapper", version, GetPodNamespaceEnvVar(), podname)
					container.Name = "gateway-proxy-v2"
					container.Args = []string{"--disable-hot-restart"}

					rb := ResourceBuilder{
						Namespace:  namespace,
						Name:       "gateway-proxy-v2",
						Labels:     labels,
						Containers: []ContainerSpec{container},
					}
					deploy := rb.GetDeploymentAppsv1()
					deploy.Spec.Selector = &metav1.LabelSelector{
						MatchLabels: selector,
					}
					deploy.Spec.Template.ObjectMeta.Labels = podLabels
					deploy.Spec.Template.ObjectMeta.Annotations = map[string]string{
						"prometheus.io/path":   "/metrics",
						"prometheus.io/port":   "8081",
						"prometheus.io/scrape": "true",
					}
					deploy.Spec.Template.Spec.Volumes = []v1.Volume{{
						Name: "envoy-config",
						VolumeSource: v1.VolumeSource{
							ConfigMap: &v1.ConfigMapVolumeSource{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "gateway-proxy-v2-envoy-config",
								},
							},
						},
					}}
					deploy.Spec.Template.Spec.Containers[0].ImagePullPolicy = pullPolicy
					deploy.Spec.Template.Spec.Containers[0].Ports = []v1.ContainerPort{
						{Name: "http", ContainerPort: 8080, Protocol: "TCP"},
						{Name: "https", ContainerPort: 8443, Protocol: "TCP"},
					}
					deploy.Spec.Template.Spec.Containers[0].VolumeMounts = []v1.VolumeMount{{
						Name:      "envoy-config",
						ReadOnly:  false,
						MountPath: "/etc/envoy",
						SubPath:   "",
					}}
					truez := true
					falsez := false
					deploy.Spec.Template.Spec.Containers[0].SecurityContext = &v1.SecurityContext{
						Capabilities: &v1.Capabilities{
							Add:  []v1.Capability{"NET_BIND_SERVICE"},
							Drop: []v1.Capability{"ALL"},
						},
						ReadOnlyRootFilesystem:   &truez,
						AllowPrivilegeEscalation: &falsez,
					}
					deploy.Spec.Template.Spec.ServiceAccountName = "gateway-proxy"
					gatewayProxyDeployment = deploy
				})

				Context("gateway-proxy daemonset", func() {
					var (
						daemonSet *appsv1.DaemonSet
					)
					BeforeEach(func() {
						daemonSet = &appsv1.DaemonSet{
							TypeMeta: metav1.TypeMeta{
								Kind:       "DaemonSet",
								APIVersion: "apps/v1",
							},
							ObjectMeta: gatewayProxyDeployment.ObjectMeta,
							Spec: appsv1.DaemonSetSpec{
								Selector: gatewayProxyDeployment.Spec.Selector,
								Template: gatewayProxyDeployment.Spec.Template,
							},
						}
						for i, port := range daemonSet.Spec.Template.Spec.Containers[0].Ports {
							port.HostPort = port.ContainerPort
							daemonSet.Spec.Template.Spec.Containers[0].Ports[i] = port
						}
						daemonSet.Spec.Template.Spec.DNSPolicy = v1.DNSClusterFirstWithHostNet
						daemonSet.Spec.Template.Spec.HostNetwork = true

					})

					It("creates a daemonset", func() {
						helmFlags := "--namespace " + namespace + " --set gatewayProxies.gatewayProxyV2.kind.deployment=null --set gatewayProxies.gatewayProxyV2.kind.daemonSet.hostPort=true"
						prepareMakefile(helmFlags)
						testManifest.Expect("DaemonSet", gatewayProxyDeployment.Namespace, gatewayProxyDeployment.Name).To(BeEquivalentTo(daemonSet))
					})
				})

				It("creates a deployment", func() {
					helmFlags := "--namespace " + namespace + " --set namespace.create=true"
					prepareMakefile(helmFlags)
					testManifest.ExpectDeploymentAppsV1(gatewayProxyDeployment)
				})

				It("disables net bind", func() {
					helmFlags := "--namespace " + namespace + " --set gatewayProxies.gatewayProxyV2.podTemplate.disableNetBind=true"
					prepareMakefile(helmFlags)
					gatewayProxyDeployment.Spec.Template.Spec.Containers[0].SecurityContext.Capabilities.Add = nil
					testManifest.ExpectDeploymentAppsV1(gatewayProxyDeployment)
				})

				It("unprivelged user", func() {
					helmFlags := "--namespace " + namespace + " --set gatewayProxies.gatewayProxyV2.podTemplate.runUnprivileged=true"
					prepareMakefile(helmFlags)
					truez := true
					uid := int64(10101)
					gatewayProxyDeployment.Spec.Template.Spec.Containers[0].SecurityContext.RunAsNonRoot = &truez
					gatewayProxyDeployment.Spec.Template.Spec.Containers[0].SecurityContext.RunAsUser = &uid
					testManifest.ExpectDeploymentAppsV1(gatewayProxyDeployment)
				})

				It("enables anti affinity ", func() {
					helmFlags := "--namespace " + namespace + " --set gatewayProxies.gatewayProxyV2.kind.deployment.antiAffinity=true"
					prepareMakefile(helmFlags)
					gatewayProxyDeployment.Spec.Template.Spec.Affinity = &v1.Affinity{
						PodAntiAffinity: &v1.PodAntiAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{{
								Weight: 100,
								PodAffinityTerm: v1.PodAffinityTerm{
									TopologyKey: "kubernetes.io/hostname",
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{"gloo": "gateway-proxy"},
									},
								},
							}},
						},
					}
					testManifest.ExpectDeploymentAppsV1(gatewayProxyDeployment)
				})

				It("enables probes", func() {
					helmFlags := "--namespace " + namespace + " --set namespace.create=true --set gatewayProxies.gatewayProxyV2.podTemplate.probes=true"

					gatewayProxyDeployment.Spec.Template.Spec.Containers[0].ReadinessProbe = &v1.Probe{
						Handler: v1.Handler{
							Exec: &v1.ExecAction{
								Command: []string{
									"wget", "-O", "/dev/null", "localhost:19000/ready",
								},
							},
						},
						InitialDelaySeconds: 1,
						PeriodSeconds:       10,
						FailureThreshold:    10,
					}
					gatewayProxyDeployment.Spec.Template.Spec.Containers[0].LivenessProbe = &v1.Probe{
						Handler: v1.Handler{
							Exec: &v1.ExecAction{
								Command: []string{
									"wget", "-O", "/dev/null", "localhost:19000/server_info",
								},
							},
						},
						InitialDelaySeconds: 1,
						PeriodSeconds:       10,
						FailureThreshold:    10,
					}
					prepareMakefile(helmFlags)
					testManifest.ExpectDeploymentAppsV1(gatewayProxyDeployment)
				})

				It("has limits", func() {
					helmFlags := "--namespace " + namespace + " --set namespace.create=true --set gatewayProxies.gatewayProxyV2.podTemplate.resources.limits.memory=2  --set gatewayProxies.gatewayProxyV2.podTemplate.resources.limits.cpu=3 --set gatewayProxies.gatewayProxyV2.podTemplate.resources.requests.memory=4  --set gatewayProxies.gatewayProxyV2.podTemplate.resources.requests.cpu=5"
					prepareMakefile(helmFlags)

					// Add the limits we are testing:
					gatewayProxyDeployment.Spec.Template.Spec.Containers[0].Resources = v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("2"),
							v1.ResourceCPU:    resource.MustParse("3"),
						},
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("4"),
							v1.ResourceCPU:    resource.MustParse("5"),
						},
					}
					testManifest.ExpectDeploymentAppsV1(gatewayProxyDeployment)
				})

				It("can overwrite the container image information", func() {
					gatewayProxyDeployment.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("gcr.io/solo-public/gloo-envoy-wrapper:%s", version)
					gatewayProxyDeployment.Spec.Template.Spec.Containers[0].ImagePullPolicy = "Always"
					helmFlags := "--namespace " + namespace + " --set namespace.create=true --set gatewayProxies.gatewayProxyV2.podTemplate.image.pullPolicy=Always --set gatewayProxies.gatewayProxyV2.podTemplate.image.registry=gcr.io/solo-public"
					prepareMakefile(helmFlags)

					testManifest.ExpectDeploymentAppsV1(gatewayProxyDeployment)
				})

				It("adds readConfig annotations", func() {
					gatewayProxyDeployment.Spec.Template.Annotations["readconfig-stats"] = "/stats"
					gatewayProxyDeployment.Spec.Template.Annotations["readconfig-ready"] = "/ready"
					gatewayProxyDeployment.Spec.Template.Annotations["readconfig-config_dump"] = "/config_dump"
					gatewayProxyDeployment.Spec.Template.Annotations["readconfig-port"] = "8082"

					helmFlags := "--namespace " + namespace + " --set namespace.create=true --set gatewayProxies.gatewayProxyV2.readConfig=true"
					prepareMakefile(helmFlags)
					testManifest.ExpectDeploymentAppsV1(gatewayProxyDeployment)
				})

				It("can add extra sidecar containers to the gateway-proxy deployment", func() {
					gatewayProxyDeployment.Spec.Template.Spec.Containers = append(
						gatewayProxyDeployment.Spec.Template.Spec.Containers,
						v1.Container{
							Name:  "nginx",
							Image: "nginx:1.7.9",
							Ports: []v1.ContainerPort{{ContainerPort: 80}},
						})

					gatewayProxyDeployment.Spec.Template.Spec.Containers[0].VolumeMounts = append(
						gatewayProxyDeployment.Spec.Template.Spec.Containers[0].VolumeMounts,
						v1.VolumeMount{
							Name:      "shared-data",
							MountPath: "/usr/share/shared-data",
						})

					gatewayProxyDeployment.Spec.Template.Spec.Volumes = append(
						gatewayProxyDeployment.Spec.Template.Spec.Volumes,
						v1.Volume{
							Name: "shared-data",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{},
							},
						})

					helmFlags := "--namespace " + namespace + " --set namespace.create=true --set gatewayProxies.gatewayProxyV2.extraContainersHelper=gloo.testcontainer"
					prepareMakefile(helmFlags)
					testManifest.ExpectDeploymentAppsV1(gatewayProxyDeployment)
				})
			})

			Context("gateway validation resources", func() {
				It("creates a service for the gateway validation port", func() {
					gwService := makeUnstructured(`
apiVersion: v1
kind: Service
metadata:
  labels:
    app: gloo
    gloo: gateway
    installationId: ` + helmTestInstallId + `
  name: gateway
  namespace: ` + namespace + `
spec:
  ports:
  - name: https
    port: 443
    protocol: TCP
    targetPort: 8443
  selector:
    gloo: gateway
    installationId: ` + helmTestInstallId + `
`)

					prepareMakefile("--namespace " + namespace)
					testManifest.ExpectUnstructured(gwService.GetKind(), gwService.GetNamespace(), gwService.GetName()).To(BeEquivalentTo(gwService))

				})

				It("creates settings with the gateway config", func() {
					settings := makeUnstructured(`
apiVersion: gloo.solo.io/v1
kind: Settings
metadata:
  annotations:
    helm.sh/hook: pre-install
    helm.sh/hook-weight: "5"
  labels:
    app: gloo
    installationId: ` + helmTestInstallId + `
  name: default
  namespace: ` + namespace + `
spec:
  gateway:
    validation:
      alwaysAccept: true
      proxyValidationServerAddr: gloo:9988
  gloo:
    xdsBindAddr: 0.0.0.0:9977
  kubernetesArtifactSource: {}
  kubernetesConfigSource: {}
  kubernetesSecretSource: {}
  refreshRate: 60s
  discoveryNamespace: ` + namespace + `
`)

					prepareMakefile("--namespace " + namespace)
					testManifest.ExpectUnstructured(settings.GetKind(), settings.GetNamespace(), settings.GetName()).To(BeEquivalentTo(settings))
				})

				It("creates the validating webhook configuration", func() {
					vwc := makeUnstructured(`

apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingWebhookConfiguration
metadata:
  name: gloo-gateway-validation-webhook-` + namespace + `
  labels:
    app: gloo
    gloo: gateway
    installationId: ` + helmTestInstallId + `
  annotations:
    "helm.sh/hook": pre-install
    "helm.sh/hook-weight": "5" # should come before cert-gen job
webhooks:
  - name: gateway.` + namespace + `.svc  # must be a domain with at least three segments separated by dots
    clientConfig:
      service:
        name: gateway
        namespace: ` + namespace + `
        path: "/validation"
      caBundle: "" # update manually or use certgen job
    rules:
      - operations: [ "CREATE", "UPDATE", "DELETE" ]
        apiGroups: ["gateway.solo.io", "gateway.solo.io.v2"]
        apiVersions: ["v1", "v2"]
        resources: ["*"]
    failurePolicy: Ignore

`)
					prepareMakefile("--namespace " + namespace)
					testManifest.ExpectUnstructured(vwc.GetKind(), vwc.GetNamespace(), vwc.GetName()).To(BeEquivalentTo(vwc))
				})

				It("adds the validation port and mounts the certgen secret to the gateway deployment", func() {

					gwDeployment := makeUnstructured(`
# Source: gloo/templates/5-gateway-deployment.yaml

apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: gloo
    gloo: gateway
    installationId: ` + helmTestInstallId + `
  name: gateway-v2
  namespace: ` + namespace + `
spec:
  replicas: 1
  selector:
    matchLabels:
      gloo: gateway
      installationId: ` + helmTestInstallId + `
  template:
    metadata:
      labels:
        gloo: gateway
        installationId: ` + helmTestInstallId + `
      annotations:
        prometheus.io/path: /metrics
        prometheus.io/port: "9091"
        prometheus.io/scrape: "true"
    spec:
      serviceAccountName: gateway
      containers:
      - image: quay.io/solo-io/gateway:` + version + `
        imagePullPolicy: Always
        name: gateway
        ports:
          - containerPort: 8443
            name: https
            protocol: TCP

        securityContext:
          readOnlyRootFilesystem: true
          allowPrivilegeEscalation: false
          runAsNonRoot: true
          runAsUser: 10101
          capabilities:
            drop:
            - ALL
        env:
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          - name: START_STATS_SERVER
            value: "true"
          - name: VALIDATION_MUST_START
            value: "true"
        volumeMounts:
          - mountPath: /etc/gateway/validation-certs
            name: validation-certs
        readinessProbe:
          tcpSocket:
            port: 8443
          initialDelaySeconds: 1
          periodSeconds: 2
          failureThreshold: 10
      volumes:
        - name: validation-certs
          secret:
            defaultMode: 420
            secretName: gateway-validation-certs
`)
					prepareMakefile("--namespace " + namespace)
					testManifest.ExpectUnstructured(gwDeployment.GetKind(), gwDeployment.GetNamespace(), gwDeployment.GetName()).To(BeEquivalentTo(gwDeployment))
				})

				It("creates the certgen job, rbac, and service account", func() {
					job := makeUnstructured(`
apiVersion: batch/v1
kind: Job
metadata:
  labels:
    app: gloo
    gloo: gateway-certgen
    installationId: ` + helmTestInstallId + `
  name: gateway-certgen
  namespace: ` + namespace + `
  annotations:
    "helm.sh/hook": pre-install
    "helm.sh/hook-weight": "10"
spec:
  template:
    metadata:
      labels:
        gloo: gateway-certgen
    spec:
      serviceAccountName: gateway-certgen
      containers:
        - image: quay.io/solo-io/certgen:` + version + `
          imagePullPolicy: Always
          name: certgen
          env:
            - name: POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
          args:
            - "--secret-name=gateway-validation-certs"
            - "--svc-name=gateway"
            - "--validating-webhook-configuration-name=gloo-gateway-validation-webhook-` + namespace + `"
      restartPolicy: OnFailure

`)
					testManifest.ExpectUnstructured(job.GetKind(), job.GetNamespace(), job.GetName()).To(BeEquivalentTo(job))

					clusterRole := makeUnstructured(`

# this role requires access to cluster-scoped resources
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
    name: gloo-gateway-secret-create-vwc-update-` + namespace + `
    labels:
        app: gloo
        gloo: rbac
        installationId: ` + helmTestInstallId + `
    annotations:
      "helm.sh/hook": "pre-install"
      "helm.sh/hook-weight": "5"
rules:
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["create", "get", "update"]
- apiGroups: ["admissionregistration.k8s.io"]
  resources: ["validatingwebhookconfigurations"]
  verbs: ["get", "update"]
`)
					testManifest.ExpectUnstructured(clusterRole.GetKind(), clusterRole.GetNamespace(), clusterRole.GetName()).To(BeEquivalentTo(clusterRole))

					clusterRoleBinding := makeUnstructured(`
# this role requires access to cluster-scoped resources
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: gloo-gateway-secret-create-vwc-update-` + namespace + `
  labels:
    app: gloo
    gloo: rbac
    installationId: ` + helmTestInstallId + `
  annotations:
    "helm.sh/hook": "pre-install"
    "helm.sh/hook-weight": "5"
subjects:
- kind: ServiceAccount
  name: gateway-certgen
  namespace: ` + namespace + `
roleRef:
  kind: ClusterRole
  name: gloo-gateway-secret-create-vwc-update-` + namespace + `
  apiGroup: rbac.authorization.k8s.io
---
`)
					testManifest.ExpectUnstructured(clusterRoleBinding.GetKind(), clusterRoleBinding.GetNamespace(), clusterRoleBinding.GetName()).To(BeEquivalentTo(clusterRoleBinding))

					serviceAccount := makeUnstructured(`

apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: gloo
    gloo: gateway
    installationId: ` + helmTestInstallId + `
  annotations:
    "helm.sh/hook": "pre-install"
    "helm.sh/hook-weight": "5"
  name: gateway-certgen
  namespace: ` + namespace + `

`)
					testManifest.ExpectUnstructured(serviceAccount.GetKind(), serviceAccount.GetNamespace(), serviceAccount.GetName()).To(BeEquivalentTo(serviceAccount))

				})
			})
		})
		Context("control plane deployments", func() {
			updateDeployment := func(deploy *appsv1.Deployment) {
				deploy.Spec.Selector = &metav1.LabelSelector{
					MatchLabels: selector,
				}
				deploy.Spec.Template.ObjectMeta.Labels = map[string]string{}
				deploy.Spec.Template.ObjectMeta.Labels["installationId"] = helmTestInstallId
				for k, v := range selector {
					deploy.Spec.Template.ObjectMeta.Labels[k] = v
				}

				truez := true
				falsez := false
				user := int64(10101)
				deploy.Spec.Template.Spec.Containers[0].SecurityContext = &v1.SecurityContext{
					Capabilities: &v1.Capabilities{
						Drop: []v1.Capability{"ALL"},
					},
					RunAsNonRoot:             &truez,
					RunAsUser:                &user,
					ReadOnlyRootFilesystem:   &truez,
					AllowPrivilegeEscalation: &falsez,
				}
				deploy.Spec.Template.Spec.Containers[0].ImagePullPolicy = pullPolicy
			}
			Context("gloo deployment", func() {
				var (
					glooDeployment *appsv1.Deployment
					labels         map[string]string
				)
				BeforeEach(func() {
					labels = map[string]string{
						"gloo": "gloo",
						"app":  "gloo",
					}
					setGlobalLabels(labels)
					selector = map[string]string{
						"gloo": "gloo",
					}
					setGlobalLabels(selector)
					container := GetQuayContainerSpec("gloo", version, GetPodNamespaceEnvVar(), GetPodNamespaceStats())

					rb := ResourceBuilder{
						Namespace:   namespace,
						Name:        "gloo",
						Labels:      labels,
						Annotations: statsAnnotations,
						Containers:  []ContainerSpec{container},
					}
					deploy := rb.GetDeploymentAppsv1()
					updateDeployment(deploy)
					deploy.Spec.Template.Spec.Volumes = []v1.Volume{{
						Name: "labels-volume",
						VolumeSource: v1.VolumeSource{
							DownwardAPI: &v1.DownwardAPIVolumeSource{
								Items: []v1.DownwardAPIVolumeFile{{
									Path: "labels",
									FieldRef: &v1.ObjectFieldSelector{
										FieldPath: "metadata.labels",
									},
								}},
							},
						},
					}}
					deploy.Spec.Template.Spec.Containers[0].VolumeMounts = []v1.VolumeMount{{
						Name:      "labels-volume",
						MountPath: "/etc/gloo",
						ReadOnly:  true,
					}}
					deploy.Spec.Template.Spec.Containers[0].Ports = glooPorts
					deploy.Spec.Template.Spec.Containers[0].Resources = v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("256Mi"),
							v1.ResourceCPU:    resource.MustParse("500m"),
						},
					}
					deploy.Spec.Template.Spec.Containers[0].ReadinessProbe = &v1.Probe{
						Handler: v1.Handler{
							TCPSocket: &v1.TCPSocketAction{
								Port: intstr.FromInt(9977),
							},
						},
						InitialDelaySeconds: 1,
						PeriodSeconds:       2,
						FailureThreshold:    10,
					}
					deploy.Spec.Template.Spec.ServiceAccountName = "gloo"
					glooDeployment = deploy
				})

				It("should create a deployment", func() {
					helmFlags := "--namespace " + namespace + " --set namespace.create=true"
					prepareMakefile(helmFlags)
					testManifest.ExpectDeploymentAppsV1(glooDeployment)
				})

				It("has limits", func() {
					helmFlags := "--namespace " + namespace + " --set namespace.create=true --set gloo.deployment.resources.limits.memory=2  --set gloo.deployment.resources.limits.cpu=3 --set gloo.deployment.resources.requests.memory=4  --set gloo.deployment.resources.requests.cpu=5"
					prepareMakefile(helmFlags)

					// Add the limits we are testing:
					glooDeployment.Spec.Template.Spec.Containers[0].Resources = v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("2"),
							v1.ResourceCPU:    resource.MustParse("3"),
						},
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("4"),
							v1.ResourceCPU:    resource.MustParse("5"),
						},
					}
					testManifest.ExpectDeploymentAppsV1(glooDeployment)
				})

				It("can overwrite the container image information", func() {
					container := GetContainerSpec("gcr.io/solo-public", "gloo", version, GetPodNamespaceEnvVar(), GetPodNamespaceStats())
					container.PullPolicy = "Always"
					rb := ResourceBuilder{
						Namespace:   namespace,
						Name:        "gloo",
						Labels:      labels,
						Annotations: statsAnnotations,
						Containers:  []ContainerSpec{container},
					}
					deploy := rb.GetDeploymentAppsv1()
					updateDeployment(deploy)
					deploy.Spec.Template.Spec.Containers[0].Ports = glooPorts
					deploy.Spec.Template.Spec.ServiceAccountName = "gloo"

					glooDeployment = deploy
					helmFlags := "--namespace " + namespace + " --set namespace.create=true --set gloo.deployment.image.pullPolicy=Always --set gloo.deployment.image.registry=gcr.io/solo-public"
					prepareMakefile(helmFlags)

				})
			})

			Context("gateway deployment", func() {
				var (
					gatewayDeployment *appsv1.Deployment
					labels            map[string]string
				)
				BeforeEach(func() {
					labels = map[string]string{
						"gloo": "gateway",
						"app":  "gloo",
					}
					setGlobalLabels(labels)
					selector = map[string]string{
						"gloo": "gateway",
					}
					setGlobalLabels(selector)
					container := GetQuayContainerSpec("gateway", version, GetPodNamespaceEnvVar(), GetPodNamespaceStats())

					rb := ResourceBuilder{
						Namespace:   namespace,
						Name:        "gateway-v2",
						Labels:      labels,
						Annotations: statsAnnotations,
						Containers:  []ContainerSpec{container},
					}
					deploy := rb.GetDeploymentAppsv1()
					updateDeployment(deploy)
					deploy.Spec.Template.Spec.ServiceAccountName = "gateway"

					deploy.Spec.Template.Spec.Volumes = []v1.Volume{{
						Name: "validation-certs",
						VolumeSource: v1.VolumeSource{
							Secret: &v1.SecretVolumeSource{
								SecretName:  "gateway-validation-certs",
								DefaultMode: proto.Int(420),
							},
						},
					}}
					deploy.Spec.Template.Spec.Containers[0].VolumeMounts = []v1.VolumeMount{{
						Name:      "validation-certs",
						MountPath: "/etc/gateway/validation-certs",
					}}
					deploy.Spec.Template.Spec.Containers[0].Ports = []v1.ContainerPort{{
						Name:          "https",
						ContainerPort: 8443,
						Protocol:      "TCP",
					}}
					deploy.Spec.Template.Spec.Containers[0].Env = append(deploy.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{
						Name:  "VALIDATION_MUST_START",
						Value: "true",
					})

					deploy.Spec.Template.Spec.Containers[0].ReadinessProbe = &v1.Probe{
						Handler: v1.Handler{
							TCPSocket: &v1.TCPSocketAction{
								Port: intstr.FromInt(8443),
							},
						},
						InitialDelaySeconds: 1,
						PeriodSeconds:       2,
						FailureThreshold:    10,
					}

					gatewayDeployment = deploy
				})

				It("has a creates a deployment", func() {
					helmFlags := "--namespace " + namespace + " --set namespace.create=true"
					prepareMakefile(helmFlags)
					testManifest.ExpectDeploymentAppsV1(gatewayDeployment)
				})

				It("has limits", func() {
					helmFlags := "--namespace " + namespace + " --set namespace.create=true --set gateway.deployment.resources.limits.memory=2  --set gateway.deployment.resources.limits.cpu=3 --set gateway.deployment.resources.requests.memory=4  --set gateway.deployment.resources.requests.cpu=5"
					prepareMakefile(helmFlags)

					// Add the limits we are testing:
					gatewayDeployment.Spec.Template.Spec.Containers[0].Resources = v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("2"),
							v1.ResourceCPU:    resource.MustParse("3"),
						},
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("4"),
							v1.ResourceCPU:    resource.MustParse("5"),
						},
					}
					testManifest.ExpectDeploymentAppsV1(gatewayDeployment)
				})

				It("can overwrite the container image information", func() {
					container := GetContainerSpec("gcr.io/solo-public", "gateway", version, GetPodNamespaceEnvVar(), GetPodNamespaceStats())
					container.PullPolicy = "Always"
					rb := ResourceBuilder{
						Namespace:   namespace,
						Name:        "gateway",
						Labels:      labels,
						Annotations: statsAnnotations,
						Containers:  []ContainerSpec{container},
					}
					deploy := rb.GetDeploymentAppsv1()
					updateDeployment(deploy)

					gatewayDeployment = deploy
					helmFlags := "--namespace " + namespace + " --set namespace.create=true --set gateway.deployment.image.pullPolicy=Always --set gateway.deployment.image.registry=gcr.io/solo-public"
					prepareMakefile(helmFlags)

				})
			})

			Context("discovery deployment", func() {
				var (
					discoveryDeployment *appsv1.Deployment
					labels              map[string]string
				)
				BeforeEach(func() {
					labels = map[string]string{
						"gloo": "discovery",
						"app":  "gloo",
					}
					setGlobalLabels(labels)
					selector = map[string]string{
						"gloo": "discovery",
					}
					setGlobalLabels(selector)
					container := GetQuayContainerSpec("discovery", version, GetPodNamespaceEnvVar(), GetPodNamespaceStats())

					rb := ResourceBuilder{
						Namespace:   namespace,
						Name:        "discovery",
						Labels:      labels,
						Annotations: statsAnnotations,
						Containers:  []ContainerSpec{container},
					}
					deploy := rb.GetDeploymentAppsv1()
					updateDeployment(deploy)
					deploy.Spec.Template.Spec.ServiceAccountName = "discovery"
					discoveryDeployment = deploy
				})

				It("has a creates a deployment", func() {
					helmFlags := "--namespace " + namespace + " --set namespace.create=true"
					prepareMakefile(helmFlags)
					testManifest.ExpectDeploymentAppsV1(discoveryDeployment)
				})

				It("disables probes", func() {
					helmFlags := "--namespace " + namespace + " --set namespace.create=true --set discovery.deployment.probes=false"
					prepareMakefile(helmFlags)
					discoveryDeployment.Spec.Template.Spec.Containers[0].ReadinessProbe = nil
					discoveryDeployment.Spec.Template.Spec.Containers[0].LivenessProbe = nil
					testManifest.ExpectDeploymentAppsV1(discoveryDeployment)
				})

				It("has limits", func() {
					helmFlags := "--namespace " + namespace + " --set namespace.create=true --set discovery.deployment.resources.limits.memory=2  --set discovery.deployment.resources.limits.cpu=3 --set discovery.deployment.resources.requests.memory=4  --set discovery.deployment.resources.requests.cpu=5"
					prepareMakefile(helmFlags)

					// Add the limits we are testing:
					discoveryDeployment.Spec.Template.Spec.Containers[0].Resources = v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("2"),
							v1.ResourceCPU:    resource.MustParse("3"),
						},
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("4"),
							v1.ResourceCPU:    resource.MustParse("5"),
						},
					}
					testManifest.ExpectDeploymentAppsV1(discoveryDeployment)
				})

				It("can overwrite the container image information", func() {
					container := GetContainerSpec("gcr.io/solo-public", "discovery", version, GetPodNamespaceEnvVar(), GetPodNamespaceStats())
					container.PullPolicy = "Always"
					rb := ResourceBuilder{
						Namespace:   namespace,
						Name:        "discovery",
						Labels:      labels,
						Annotations: statsAnnotations,
						Containers:  []ContainerSpec{container},
					}
					deploy := rb.GetDeploymentAppsv1()
					updateDeployment(deploy)

					discoveryDeployment = deploy
					deploy.Spec.Template.Spec.ServiceAccountName = "discovery"
					helmFlags := "--namespace " + namespace + " --set namespace.create=true --set discovery.deployment.image.pullPolicy=Always --set discovery.deployment.image.registry=gcr.io/solo-public"
					prepareMakefile(helmFlags)

				})
			})

		})

		Describe("configmaps", func() {
			var (
				gatewayProxyConfigMapName = "gateway-proxy-v2-envoy-config"
			)

			labels := map[string]string{
				"gloo":             "gateway-proxy",
				"app":              "gloo",
				"gateway-proxy-id": "gateway-proxy-v2",
			}

			BeforeEach(func() {
				setGlobalLabels(labels)
			})

			Describe("gateway proxy - tracing config", func() {
				It("has a proxy without tracing", func() {
					helmFlags := "--namespace " + namespace + " --set namespace.create=true  --set gatewayProxies.gatewayProxyV2.service.extraAnnotations.test=test"
					prepareMakefile(helmFlags)
					proxySpec := make(map[string]string)
					proxySpec["envoy.yaml"] = confWithoutTracing
					cmRb := ResourceBuilder{
						Namespace: namespace,
						Name:      gatewayProxyConfigMapName,
						Labels:    labels,
						Data:      proxySpec,
					}
					proxy := cmRb.GetConfigMap()
					testManifest.ExpectConfigMapWithYamlData(proxy)
				})

				It("has a proxy with tracing provider", func() {
					prepareMakefileFromValuesFile("install/test/val_tracing_provider.yaml")
					proxySpec := make(map[string]string)
					proxySpec["envoy.yaml"] = confWithTracingProvider
					cmRb := ResourceBuilder{
						Namespace: namespace,
						Name:      gatewayProxyConfigMapName,
						Labels:    labels,
						Data:      proxySpec,
					}
					proxy := cmRb.GetConfigMap()
					testManifest.ExpectConfigMapWithYamlData(proxy)
				})

				It("has a proxy with tracing provider and cluster", func() {
					prepareMakefileFromValuesFile("install/test/val_tracing_provider_cluster.yaml")
					proxySpec := make(map[string]string)
					proxySpec["envoy.yaml"] = confWithTracingProviderCluster
					cmRb := ResourceBuilder{
						Namespace: namespace,
						Name:      gatewayProxyConfigMapName,
						Labels:    labels,
						Data:      proxySpec,
					}
					proxy := cmRb.GetConfigMap()
					testManifest.ExpectConfigMapWithYamlData(proxy)
				})
			})

			Describe("gateway proxy -- readConfig config", func() {
				It("has a listener for reading a subset of the admin api", func() {
					helmFlags := "--namespace " + namespace + " --set gatewayProxies.gatewayProxyV2.readConfig=true"
					prepareMakefile(helmFlags)
					proxySpec := make(map[string]string)
					proxySpec["envoy.yaml"] = confWithReadConfig
					cmRb := ResourceBuilder{
						Namespace: namespace,
						Name:      gatewayProxyConfigMapName,
						Labels:    labels,
						Data:      proxySpec,
					}
					proxy := cmRb.GetConfigMap()
					testManifest.ExpectConfigMapWithYamlData(proxy)
				})
			})

		})

		Describe("merge ingress and gateway", func() {

			// helper for passing a values file
			prepareMakefileFromValuesFile := func(valuesFile string) {
				helmFlags := "--namespace " + namespace +
					" -f " + valuesFile
				prepareMakefile(helmFlags)
			}

			It("merges the config correctly, allow override of ingress without altering gloo", func() {
				deploymentLabels := map[string]string{
					"app": "gloo", "gloo": "gloo",
				}
				setGlobalLabels(deploymentLabels)
				selectors := map[string]string{
					"gloo": "gloo",
				}
				setGlobalLabels(selectors)
				podLabels := map[string]string{
					"gloo": "gloo",
				}
				setGlobalLabels(podLabels)
				var glooDeploymentPostMerge = &appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Deployment",
						APIVersion: "apps/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "gloo",
						Namespace: "gloo-system",
						Labels:    deploymentLabels,
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: pointer.Int32Ptr(1),
						Selector: &metav1.LabelSelector{MatchLabels: selectors},
						Template: v1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels:      podLabels,
								Annotations: statsAnnotations,
							},
							Spec: v1.PodSpec{
								Volumes: []v1.Volume{{
									Name: "labels-volume",
									VolumeSource: v1.VolumeSource{
										DownwardAPI: &v1.DownwardAPIVolumeSource{
											Items: []v1.DownwardAPIVolumeFile{{
												Path: "labels",
												FieldRef: &v1.ObjectFieldSelector{
													FieldPath: "metadata.labels",
												},
											}},
										},
									},
								}},
								ServiceAccountName: "gloo",
								Containers: []v1.Container{
									{
										VolumeMounts: []v1.VolumeMount{{
											Name:      "labels-volume",
											MountPath: "/etc/gloo",
											ReadOnly:  true,
										}},
										Name: "gloo",
										// Note: this was NOT overwritten
										Image: "quay.io/solo-io/gloo:dev",
										Ports: glooPorts,
										Env: []v1.EnvVar{
											{
												Name: "POD_NAMESPACE",
												ValueFrom: &v1.EnvVarSource{
													FieldRef: &v1.ObjectFieldSelector{APIVersion: "", FieldPath: "metadata.namespace"},
												},
											},
											{
												Name:  "START_STATS_SERVER",
												Value: "true",
											},
										},
										Resources: v1.ResourceRequirements{
											Limits: nil,
											Requests: v1.ResourceList{
												v1.ResourceMemory: resource.MustParse("256Mi"),
												v1.ResourceCPU:    resource.MustParse("500m"),
											},
										},
										ImagePullPolicy: "Always",
										SecurityContext: &v1.SecurityContext{
											Capabilities:             &v1.Capabilities{Add: nil, Drop: []v1.Capability{"ALL"}},
											RunAsUser:                pointer.Int64Ptr(10101),
											RunAsNonRoot:             pointer.BoolPtr(true),
											ReadOnlyRootFilesystem:   pointer.BoolPtr(true),
											AllowPrivilegeEscalation: pointer.BoolPtr(false),
										},
										ReadinessProbe: &v1.Probe{
											Handler: v1.Handler{
												TCPSocket: &v1.TCPSocketAction{
													Port: intstr.FromInt(9977),
												},
											},
											InitialDelaySeconds: 1,
											PeriodSeconds:       2,
											FailureThreshold:    10,
										},
									},
								},
							},
						},
					},
				}
				ingressDeploymentLabels := map[string]string{
					"app": "gloo", "gloo": "ingress",
				}
				setGlobalLabels(ingressDeploymentLabels)
				ingressSelector := map[string]string{
					"gloo": "ingress",
				}
				setGlobalLabels(ingressSelector)
				ingressPodLabels := map[string]string{
					"gloo": "ingress",
				}
				setGlobalLabels(ingressPodLabels)
				var ingressDeploymentPostMerge = &appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Deployment",
						APIVersion: "apps/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ingress",
						Namespace: "gloo-system",
						Labels:    ingressDeploymentLabels,
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: pointer.Int32Ptr(1),
						Selector: &metav1.LabelSelector{MatchLabels: ingressSelector},
						Template: v1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: ingressPodLabels,
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Name: "ingress",
										// Note: this WAS overwritten
										Image: "docker.io/ilackarms/ingress:test-ilackarms",
										Env: []v1.EnvVar{
											{
												Name: "POD_NAMESPACE",
												ValueFrom: &v1.EnvVarSource{
													FieldRef: &v1.ObjectFieldSelector{APIVersion: "", FieldPath: "metadata.namespace"},
												},
											},
											{
												Name:  "ENABLE_KNATIVE_INGRESS",
												Value: "true",
											},
											{
												Name:  "KNATIVE_VERSION",
												Value: "0.8.0",
											},
											{
												Name:  "DISABLE_KUBE_INGRESS",
												Value: "true",
											},
										},
										Resources: v1.ResourceRequirements{
											Limits: nil,
										},
										ImagePullPolicy: "Always",
									},
								},
							},
						},
					},
				}
				prepareMakefileFromValuesFile("install/test/merge_ingress_values.yaml")
				testManifest.ExpectDeploymentAppsV1(glooDeploymentPostMerge)
				testManifest.ExpectDeploymentAppsV1(ingressDeploymentPostMerge)
			})

		})

	})
})

func makeUnstructured(yam string) *unstructured.Unstructured {
	jsn, err := yaml.YAMLToJSON([]byte(yam))
	Expect(err).NotTo(HaveOccurred())
	runtimeObj, err := runtime.Decode(unstructured.UnstructuredJSONScheme, jsn)
	Expect(err).NotTo(HaveOccurred())
	return runtimeObj.(*unstructured.Unstructured)
}
