package controllers_test

import (
	"context"
	"fmt"
	"sync"
	"time"

	osrmv1alpha1 "github.com/itayankri/OSRM-Operator/api/v1alpha1"
	"github.com/itayankri/OSRM-Operator/internal/metadata"
	osrmResource "github.com/itayankri/OSRM-Operator/internal/resource"
	"github.com/itayankri/OSRM-Operator/internal/status"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ClusterDeletionTimeout = 5 * time.Second
	MapBuildingTimeout     = 2 * 60 * time.Second
)

var instance *osrmv1alpha1.OSRMCluster
var defaultNamespace = "default"

var _ = Describe("OSRMClusterController", func() {
	Context("Resource requirements configurations", func() {
		AfterEach(func() {
			Expect(k8sClient.Delete(ctx, instance)).To(Succeed())
		})

		It("uses resource requirements from profile spec when provided", func() {
			instance = generateOSRMCluster("resource-requirements-config")
			expectedResources := corev1.ResourceRequirements{
				Limits: map[corev1.ResourceName]resource.Quantity{
					corev1.ResourceMemory: resource.MustParse("4Gi"),
				},
				Requests: map[corev1.ResourceName]resource.Quantity{
					corev1.ResourceMemory: resource.MustParse("4Gi"),
				},
			}
			instance.Spec.Profiles[0].Resources = &expectedResources
			Expect(k8sClient.Create(ctx, instance)).To(Succeed())
			waitForDeployment(ctx, instance, k8sClient)
			deployment := deployment(ctx, instance.Name, instance.Spec.Profiles[0].Name, osrmResource.DeploymentSuffix)
			actualResources := deployment.Spec.Template.Spec.Containers[0].Resources
			Expect(actualResources).To(Equal(expectedResources))
		})
	})

	Context("Custom Resource updates", func() {
		BeforeEach(func() {
			instance = generateOSRMCluster("custom-resource-updates")
			Expect(k8sClient.Create(ctx, instance)).To(Succeed())
			waitForDeployment(ctx, instance, k8sClient)
		})

		AfterEach(func() {
			Expect(k8sClient.Delete(ctx, instance)).To(Succeed())
		})

		It("Should update deployment CPU and memory requests and limits", func() {
			var resourceRequirements corev1.ResourceRequirements
			expectedRequirements := &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("200m"),
					corev1.ResourceMemory: resource.MustParse("200Mi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("200m"),
					corev1.ResourceMemory: resource.MustParse("200Mi"),
				},
			}

			Expect(updateWithRetry(instance, func(v *osrmv1alpha1.OSRMCluster) {
				v.Spec.Profiles[0].Resources = expectedRequirements
			})).To(Succeed())

			Eventually(func() corev1.ResourceList {
				deployment := deployment(ctx, instance.Name, instance.Spec.Profiles[0].Name, osrmResource.DeploymentSuffix)
				resourceRequirements = deployment.Spec.Template.Spec.Containers[0].Resources
				return resourceRequirements.Requests
			}, 3).Should(HaveKeyWithValue(corev1.ResourceCPU, expectedRequirements.Requests[corev1.ResourceCPU]))
			Expect(resourceRequirements.Limits).To(HaveKeyWithValue(corev1.ResourceCPU, expectedRequirements.Limits[corev1.ResourceCPU]))
			Expect(resourceRequirements.Requests).To(HaveKeyWithValue(corev1.ResourceMemory, expectedRequirements.Requests[corev1.ResourceMemory]))
			Expect(resourceRequirements.Limits).To(HaveKeyWithValue(corev1.ResourceMemory, expectedRequirements.Limits[corev1.ResourceMemory]))
		})
	})

	Context("ConfigMap updates", func() {
		testNumber := 0
		BeforeEach(func() {
			instance = generateOSRMCluster(fmt.Sprintf("configmap-updates-%d", testNumber))
			testNumber += 1
			Expect(k8sClient.Create(ctx, instance)).To(Succeed())
			waitForDeployment(ctx, instance, k8sClient)
		})

		AfterEach(func() {
			Expect(k8sClient.Delete(ctx, instance)).To(Succeed())
		})

		It("Should rollout gateway deployment after adding a new profile", func() {
			osrmProfile := "foot"
			internalEndpoint := "walking"
			minReplicas := int32(1)
			maxReplicas := int32(2)
			newProfile := &osrmv1alpha1.ProfileSpec{
				Name:             "new-profile",
				EndpointName:     "custom-endpoint",
				InternalEndpoint: &internalEndpoint,
				OSRMProfile:      &osrmProfile,
				MinReplicas:      &minReplicas,
				MaxReplicas:      &maxReplicas,
			}

			gateway := deployment(ctx, instance.Name, "", osrmResource.DeploymentSuffix)
			gatewayConfigVersionAnnotation := gateway.Spec.Template.ObjectMeta.Annotations[osrmResource.GatewayConfigVersion]

			Expect(updateWithRetry(instance, func(v *osrmv1alpha1.OSRMCluster) {
				v.Spec.Profiles = append(v.Spec.Profiles, newProfile)
			})).To(Succeed())

			Eventually(func() string {
				return deployment(ctx, instance.Name, "", osrmResource.DeploymentSuffix).Spec.Template.ObjectMeta.Annotations[osrmResource.GatewayConfigVersion]
			}, 180*time.Second).ShouldNot(Equal(gatewayConfigVersionAnnotation))
		})

		It("Should rollout gateway deployment after modifying ExposingServices", func() {
			gateway := deployment(ctx, instance.Name, "", osrmResource.DeploymentSuffix)
			gatewayConfigVersionAnnotation := gateway.Spec.Template.ObjectMeta.Annotations[osrmResource.GatewayConfigVersion]

			Expect(updateWithRetry(instance, func(v *osrmv1alpha1.OSRMCluster) {
				v.Spec.Service.ExposingServices = append(v.Spec.Service.ExposingServices, "table")
			})).To(Succeed())

			Eventually(func() string {
				return deployment(ctx, instance.Name, "", osrmResource.DeploymentSuffix).Spec.Template.ObjectMeta.Annotations[osrmResource.GatewayConfigVersion]
			}, 180*time.Second).ShouldNot(Equal(gatewayConfigVersionAnnotation))
		})

		It("Should rollout gateway deployment after editing a profile's EndpointName", func() {
			gateway := deployment(ctx, instance.Name, "", osrmResource.DeploymentSuffix)
			gatewayConfigVersionAnnotation := gateway.Spec.Template.ObjectMeta.Annotations[osrmResource.GatewayConfigVersion]

			Expect(updateWithRetry(instance, func(v *osrmv1alpha1.OSRMCluster) {
				v.Spec.Profiles[0].EndpointName = "ankri"
			})).To(Succeed())

			Eventually(func() string {
				return deployment(ctx, instance.Name, "", osrmResource.DeploymentSuffix).Spec.Template.ObjectMeta.Annotations[osrmResource.GatewayConfigVersion]
			}, 180*time.Second).ShouldNot(Equal(gatewayConfigVersionAnnotation))
		})
	})

	Context("Recreate child resources after deletion", func() {
		BeforeEach(func() {
			instance = generateOSRMCluster("recreate-children")
			Expect(k8sClient.Create(ctx, instance)).To(Succeed())
			waitForDeployment(ctx, instance, k8sClient)
		})

		AfterEach(func() {
			Expect(k8sClient.Delete(ctx, instance)).To(Succeed())
		})

		It("recreates child resources after deletion", func() {
			oldService := service(ctx, instance.Name, instance.Spec.Profiles[0].Name, osrmResource.ServiceSuffix)
			oldDeployment := deployment(ctx, instance.Name, instance.Spec.Profiles[0].Name, osrmResource.DeploymentSuffix)
			oldHpa := hpa(ctx, instance.Name, instance.Spec.Profiles[0].Name, osrmResource.HorizontalPodAutoscalerSuffix)

			Expect(k8sClient.Delete(ctx, oldService)).NotTo(HaveOccurred())
			Expect(k8sClient.Delete(ctx, oldHpa)).NotTo(HaveOccurred())
			Expect(k8sClient.Delete(ctx, oldDeployment)).NotTo(HaveOccurred())

			Eventually(func() bool {
				deployment := deployment(ctx, instance.Name, instance.Spec.Profiles[0].Name, osrmResource.DeploymentSuffix)
				return string(deployment.UID) != string(oldDeployment.UID)
			}, 5).Should(BeTrue())

			Eventually(func() bool {
				svc := service(ctx, instance.Name, instance.Spec.Profiles[0].Name, osrmResource.ServiceSuffix)
				return string(svc.UID) != string(oldService.UID)
			}, 5).Should(BeTrue())

			Eventually(func() bool {
				hpa := hpa(ctx, instance.Name, instance.Spec.Profiles[0].Name, osrmResource.HorizontalPodAutoscalerSuffix)
				return string(hpa.UID) != string(oldHpa.UID)
			}, 5).Should(BeTrue())
		})
	})

	Context("OSRMCluster CR ReconcileSuccess condition", func() {
		BeforeEach(func() {
			instance = generateOSRMCluster("reconcile-success-condition")
		})

		AfterEach(func() {
			Expect(k8sClient.Delete(ctx, instance)).To(Succeed())
		})

		It("Should keep ReconcileSuccess condition updated", func() {
			By("setting to False when spec is not valid", func() {
				// It is impossible to create a deployment with -1 replicas. Thus we expect reconcilication to fail.
				instance.Spec.Profiles[0].MinReplicas = pointer.Int32Ptr(-1)
				Expect(k8sClient.Create(ctx, instance)).To(Succeed())
				waitForOSRMClusterCreation(ctx, instance, k8sClient)

				Eventually(func() metav1.ConditionStatus {
					osrmCluster := &osrmv1alpha1.OSRMCluster{}
					Expect(k8sClient.Get(ctx, types.NamespacedName{
						Name:      instance.Name,
						Namespace: instance.Namespace,
					}, osrmCluster)).To(Succeed())

					for _, condition := range osrmCluster.Status.Conditions {
						if condition.Type == status.ConditionReconciliationSuccess {
							return condition.Status
						}
					}
					return metav1.ConditionUnknown
				}, 180*time.Second).Should(Equal(metav1.ConditionFalse))
			})

			By("setting to True when spec is valid", func() {
				// It is impossible to create a deployment with -1 replicas. Thus we expect reconcilication to fail.
				Expect(updateWithRetry(instance, func(v *osrmv1alpha1.OSRMCluster) {
					v.Spec.Profiles[0].MinReplicas = pointer.Int32Ptr(2)
				})).To(Succeed())

				Eventually(func() metav1.ConditionStatus {
					osrmCluster := &osrmv1alpha1.OSRMCluster{}
					Expect(k8sClient.Get(ctx, types.NamespacedName{
						Name:      instance.Name,
						Namespace: instance.Namespace,
					}, osrmCluster)).To(Succeed())

					for _, condition := range osrmCluster.Status.Conditions {
						if condition.Type == status.ConditionReconciliationSuccess {
							return condition.Status
						}
					}
					return metav1.ConditionUnknown
				}, 60*time.Second).Should(Equal(metav1.ConditionTrue))
			})
		})
	})

	Context("Pause reconciliation", func() {
		BeforeEach(func() {
			instance = generateOSRMCluster("pause-reconcile")
			Expect(k8sClient.Create(ctx, instance)).To(Succeed())
			waitForDeployment(ctx, instance, k8sClient)
		})

		AfterEach(func() {
			Expect(k8sClient.Delete(ctx, instance)).To(Succeed())
		})

		It("Should skip OSRMCluster if pause reconciliation annotation is set to true", func() {
			minReplicas := int32(2)
			originalMinReplicas := *instance.Spec.Profiles[0].MinReplicas
			Expect(updateWithRetry(instance, func(v *osrmv1alpha1.OSRMCluster) {
				v.SetAnnotations(map[string]string{"osrm.itayankri/operator.paused": "true"})
				v.Spec.Profiles[0].MinReplicas = &minReplicas
			})).To(Succeed())

			Eventually(func() int32 {
				return *hpa(ctx, instance.Name, instance.Spec.Profiles[0].Name, osrmResource.HorizontalPodAutoscalerSuffix).Spec.MinReplicas
			}, MapBuildingTimeout).Should(Equal(originalMinReplicas))

			Expect(updateWithRetry(instance, func(v *osrmv1alpha1.OSRMCluster) {
				v.SetAnnotations(map[string]string{"osrm.itayankri/operator.paused": "false"})
			})).To(Succeed())

			Eventually(func() int32 {
				return *hpa(ctx, instance.Name, instance.Spec.Profiles[0].Name, osrmResource.HorizontalPodAutoscalerSuffix).Spec.MinReplicas
			}, 10*time.Second).Should(Equal(minReplicas))
		})
	})

	Context("Garbage Collection", func() {
		testNumber := 0
		BeforeEach(func() {
			instance = generateOSRMCluster(fmt.Sprintf("garbage-collection-%d", testNumber))
			testNumber += 1
			Expect(k8sClient.Create(ctx, instance)).To(Succeed())
			waitForDeployment(ctx, instance, k8sClient)
		})

		AfterEach(func() {
			Expect(k8sClient.Delete(ctx, instance)).To(Succeed())
		})

		It("Should add an internal label with the custom resource's generation to all child resources", func() {
			resources := getGatewayResources(ctx, instance)
			resources = append(resources, getProfileResources(ctx, instance.Spec.Profiles[0])...)

			for _, resource := range resources {
				By(fmt.Sprintf("checking label on %s after cluster creation", resource.GetName()), func() {
					label, labelExists := resource.GetLabels()[metadata.GenerationLabelKey]
					Expect(labelExists).To(BeTrue())
					Expect(label).To(Equal("1"))
				})
			}

			Expect(updateWithRetry(instance, func(v *osrmv1alpha1.OSRMCluster) {
				v.Spec.Service.ExposingServices = append(v.Spec.Service.ExposingServices, "table")
			})).To(Succeed())

			Eventually(func() bool {
				resources = getGatewayResources(ctx, instance)
				resources = append(resources, getProfileResources(ctx, instance.Spec.Profiles[0])...)
				for _, resource := range resources {
					label, labelExists := resource.GetLabels()[metadata.GenerationLabelKey]
					if !labelExists || label != "2" {
						return false
					}
				}
				return true
			}, 180*time.Second).Should(BeTrue())
		})

		It("Should delete all child resources that has a different generation than the custom resource", func() {
			firstGenerationProfile := instance.Spec.Profiles[0]
			firstGenerationService := service(ctx, instance.Name, firstGenerationProfile.Name, osrmResource.ServiceSuffix)
			firstGenerationDeployment := deployment(ctx, instance.Name, firstGenerationProfile.Name, osrmResource.DeploymentSuffix)
			firstGenerationHPA := hpa(ctx, instance.Name, firstGenerationProfile.Name, osrmResource.HorizontalPodAutoscalerSuffix)
			firstGenerationPDB := pdb(ctx, instance.Name, firstGenerationProfile.Name, osrmResource.PodDisruptionBudgetSuffix)
			firstGenerationPVC := pvc(ctx, instance.Name, firstGenerationProfile.Name, osrmResource.PersistentVolumeClaimSuffix)
			firstGenerationJob := job(ctx, instance.Name, firstGenerationProfile.Name, osrmResource.JobSuffix)

			Expect(updateWithRetry(instance, func(v *osrmv1alpha1.OSRMCluster) {
				v.Spec.Profiles[0] = &osrmv1alpha1.ProfileSpec{
					Name:         "foot",
					EndpointName: firstGenerationProfile.EndpointName,
					MinReplicas:  firstGenerationProfile.MinReplicas,
					MaxReplicas:  firstGenerationProfile.MaxReplicas,
					Resources:    firstGenerationProfile.Resources,
				}
			})).To(Succeed())

			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      firstGenerationService.Name,
					Namespace: firstGenerationService.Namespace,
				}, &corev1.Service{})
				return errors.IsNotFound(err)
			}, 10*time.Second).Should(BeTrue())

			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      firstGenerationDeployment.Name,
					Namespace: firstGenerationDeployment.Namespace,
				}, &appsv1.Deployment{})
				return errors.IsNotFound(err)
			}, 10*time.Second).Should(BeTrue())

			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      firstGenerationHPA.Name,
					Namespace: firstGenerationHPA.Namespace,
				}, &autoscalingv1.HorizontalPodAutoscaler{})
				return errors.IsNotFound(err)
			}, 10*time.Second).Should(BeTrue())

			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      firstGenerationPDB.Name,
					Namespace: firstGenerationPDB.Namespace,
				}, &policyv1.PodDisruptionBudget{})
				return errors.IsNotFound(err)
			}, 10*time.Second).Should(BeTrue())

			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      firstGenerationJob.Name,
					Namespace: firstGenerationJob.Namespace,
				}, &batchv1.Job{})
				return errors.IsNotFound(err)
			}, 10*time.Second).Should(BeTrue())

			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      firstGenerationPVC.Name,
					Namespace: firstGenerationPVC.Namespace,
				}, &corev1.PersistentVolumeClaim{})
				return errors.IsNotFound(err)
			}, 180*time.Second).Should(BeTrue())
		})

		It("Should delete only resources that belong to the current reconciled CR", func() {
			secondInstance := generateOSRMCluster(fmt.Sprintf("garbage-collection-%d-b", testNumber))
			Expect(k8sClient.Create(ctx, secondInstance)).To(Succeed())
			waitForDeployment(ctx, secondInstance, k8sClient)

			secondInstanceService := service(ctx, secondInstance.Name, secondInstance.Spec.Profiles[0].Name, osrmResource.ServiceSuffix)
			secondInstanceDeployment := deployment(ctx, secondInstance.Name, secondInstance.Spec.Profiles[0].Name, osrmResource.DeploymentSuffix)
			secondInstanceHPA := hpa(ctx, secondInstance.Name, secondInstance.Spec.Profiles[0].Name, osrmResource.HorizontalPodAutoscalerSuffix)
			secondInstancePDB := pdb(ctx, secondInstance.Name, secondInstance.Spec.Profiles[0].Name, osrmResource.PodDisruptionBudgetSuffix)
			secondInstancePVC := pvc(ctx, secondInstance.Name, secondInstance.Spec.Profiles[0].Name, osrmResource.PersistentVolumeClaimSuffix)
			secondInstanceJob := job(ctx, secondInstance.Name, secondInstance.Spec.Profiles[0].Name, osrmResource.JobSuffix)

			Expect(updateWithRetry(instance, func(v *osrmv1alpha1.OSRMCluster) {
				v.Spec.Profiles[0].Name = "foot"
			})).To(Succeed())

			Consistently(func() string {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      secondInstanceService.Name,
					Namespace: secondInstanceService.Namespace,
				}, &corev1.Service{})
				if errors.IsNotFound(err) {
					return fmt.Sprintf("service %s deleted", secondInstanceService.Name)
				}

				err = k8sClient.Get(ctx, types.NamespacedName{
					Name:      secondInstanceDeployment.Name,
					Namespace: secondInstanceDeployment.Namespace,
				}, &appsv1.Deployment{})
				if errors.IsNotFound(err) {
					return fmt.Sprintf("deployment %s deleted", secondInstanceDeployment.Name)
				}

				err = k8sClient.Get(ctx, types.NamespacedName{
					Name:      secondInstanceHPA.Name,
					Namespace: secondInstanceHPA.Namespace,
				}, &autoscalingv1.HorizontalPodAutoscaler{})
				if errors.IsNotFound(err) {
					return fmt.Sprintf("hpa %s deleted", secondInstanceHPA.Name)
				}

				err = k8sClient.Get(ctx, types.NamespacedName{
					Name:      secondInstancePDB.Name,
					Namespace: secondInstancePDB.Namespace,
				}, &policyv1.PodDisruptionBudget{})
				if errors.IsNotFound(err) {
					return fmt.Sprintf("pdb %s deleted", secondInstancePDB.Name)
				}

				err = k8sClient.Get(ctx, types.NamespacedName{
					Name:      secondInstanceJob.Name,
					Namespace: secondInstanceJob.Namespace,
				}, &batchv1.Job{})
				if errors.IsNotFound(err) {
					return fmt.Sprintf("job %s deleted", secondInstanceJob.Name)
				}

				err = k8sClient.Get(ctx, types.NamespacedName{
					Name:      secondInstancePVC.Name,
					Namespace: secondInstancePVC.Namespace,
				}, &corev1.PersistentVolumeClaim{})
				if errors.IsNotFound(err) {
					return fmt.Sprintf("pvc %s deleted", secondInstancePVC.Name)
				}

				return "nothing deleted"
			}, 15*time.Second).Should(Equal("nothing deleted"))

			Expect(k8sClient.Delete(ctx, secondInstance)).To(Succeed())
		})
	})
})

func getGatewayResources(ctx context.Context, instance *osrmv1alpha1.OSRMCluster) []client.Object {
	wg := sync.WaitGroup{}
	wg.Add(3)
	resources := []client.Object{}

	go func() {
		defer func() {
			wg.Done()
			GinkgoRecover()
		}()
		gatewayService := service(ctx, instance.Name, "", osrmResource.ServiceSuffix)
		resources = append(resources, gatewayService)
	}()

	go func() {
		defer func() {
			wg.Done()
			GinkgoRecover()
		}()
		gatewayDeployment := deployment(ctx, instance.Name, "", osrmResource.DeploymentSuffix)
		resources = append(resources, gatewayDeployment)
	}()

	go func() {
		defer func() {
			wg.Done()
			GinkgoRecover()
		}()
		gatewayConfigMap := configMap(ctx, instance.Name, "", osrmResource.ConfigMapSuffix)
		resources = append(resources, gatewayConfigMap)
	}()

	wg.Wait()

	return resources
}

func getProfileResources(ctx context.Context, profile *osrmv1alpha1.ProfileSpec) []client.Object {
	wg := sync.WaitGroup{}
	wg.Add(6)
	resources := []client.Object{}

	go func() {
		defer func() {
			wg.Done()
			GinkgoRecover()
		}()
		profileService := service(ctx, instance.Name, instance.Spec.Profiles[0].Name, osrmResource.ServiceSuffix)
		resources = append(resources, profileService)
	}()

	go func() {
		defer func() {
			wg.Done()
			GinkgoRecover()
		}()
		profileDeployment := deployment(ctx, instance.Name, instance.Spec.Profiles[0].Name, osrmResource.DeploymentSuffix)
		resources = append(resources, profileDeployment)
	}()

	go func() {
		defer func() {
			wg.Done()
			GinkgoRecover()
		}()
		profileHpa := hpa(ctx, instance.Name, instance.Spec.Profiles[0].Name, osrmResource.HorizontalPodAutoscalerSuffix)
		resources = append(resources, profileHpa)
	}()

	go func() {
		defer func() {
			wg.Done()
			GinkgoRecover()
		}()
		profileJob := job(ctx, instance.Name, instance.Spec.Profiles[0].Name, osrmResource.JobSuffix)
		resources = append(resources, profileJob)
	}()

	go func() {
		defer func() {
			wg.Done()
			GinkgoRecover()
		}()
		profilePvc := pvc(ctx, instance.Name, instance.Spec.Profiles[0].Name, osrmResource.PersistentVolumeClaimSuffix)
		resources = append(resources, profilePvc)
	}()

	go func() {
		defer func() {
			wg.Done()
			GinkgoRecover()
		}()
		profilePdb := pdb(ctx, instance.Name, instance.Spec.Profiles[0].Name, osrmResource.PodDisruptionBudgetSuffix)
		resources = append(resources, profilePdb)
	}()

	wg.Wait()

	return resources
}

func generateOSRMCluster(name string) *osrmv1alpha1.OSRMCluster {
	storage := resource.MustParse("10Mi")
	image := "osrm/osrm-backend"
	accessMode := corev1.ReadWriteOnce
	minReplicas := int32(1)
	maxReplicas := int32(3)
	osrmCluster := &osrmv1alpha1.OSRMCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: defaultNamespace,
		},
		Spec: osrmv1alpha1.OSRMClusterSpec{
			PBFURL: "https://download.geofabrik.de/australia-oceania/marshall-islands-latest.osm.pbf",
			Image:  &image,
			Persistence: osrmv1alpha1.PersistenceSpec{
				StorageClassName: "nfs-csi",
				Storage:          &storage,
				AccessMode:       &accessMode,
			},
			Profiles: []*osrmv1alpha1.ProfileSpec{
				{
					Name:         "car",
					EndpointName: "driving",
					MinReplicas:  &minReplicas,
					MaxReplicas:  &maxReplicas,
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("100Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("100Mi"),
						},
					},
				},
			},
			Service: osrmv1alpha1.ServiceSpec{
				ExposingServices: []string{"route"},
			},
		},
	}
	return osrmCluster
}

func waitForOSRMClusterCreation(ctx context.Context, instance *osrmv1alpha1.OSRMCluster, client client.Client) {
	EventuallyWithOffset(1, func() string {
		instanceCreated := osrmv1alpha1.OSRMCluster{}
		if err := k8sClient.Get(
			ctx,
			types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace},
			&instanceCreated,
		); err != nil {
			return fmt.Sprintf("%v+", err)
		}

		if len(instanceCreated.Status.Conditions) == 0 {
			return "not ready"
		}

		return "ready"

	}, MapBuildingTimeout, 1*time.Second).Should(Equal("ready"))
}

func waitForDeployment(ctx context.Context, instance *osrmv1alpha1.OSRMCluster, client client.Client) {
	EventuallyWithOffset(1, func() string {
		instanceCreated := osrmv1alpha1.OSRMCluster{}
		if err := k8sClient.Get(
			ctx,
			types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace},
			&instanceCreated,
		); err != nil {
			return fmt.Sprintf("%v+", err)
		}

		for _, condition := range instanceCreated.Status.Conditions {
			if condition.Type == status.ConditionAvailable && condition.Status == metav1.ConditionTrue {
				return "ready"
			}
		}

		return "not ready"

	}, MapBuildingTimeout, 1*time.Second).Should(Equal("ready"))
}

func hpa(ctx context.Context, clusterName string, profileName string, suffix string) *autoscalingv1.HorizontalPodAutoscaler {
	name := fmt.Sprintf("%s-%s", clusterName, profileName)
	if len(suffix) > 0 {
		name = fmt.Sprintf("%s-%s", name, suffix)
	}
	hpa := &autoscalingv1.HorizontalPodAutoscaler{}
	EventuallyWithOffset(1, func() error {
		if err := k8sClient.Get(
			ctx,
			types.NamespacedName{Name: name, Namespace: defaultNamespace},
			hpa,
		); err != nil {
			return err
		}
		return nil
	}, MapBuildingTimeout).Should(Succeed())
	return hpa
}

func service(ctx context.Context, clusterName string, profileName string, suffix string) *corev1.Service {
	name := clusterName
	if len(profileName) > 0 {
		name = fmt.Sprintf("%s-%s", clusterName, profileName)
	}

	if len(suffix) > 0 {
		name = fmt.Sprintf("%s-%s", name, suffix)
	}
	svc := &corev1.Service{}
	EventuallyWithOffset(1, func() error {
		if err := k8sClient.Get(
			ctx,
			types.NamespacedName{Name: name, Namespace: defaultNamespace},
			svc,
		); err != nil {
			return err
		}
		return nil
	}, MapBuildingTimeout).Should(Succeed())
	return svc
}

func deployment(ctx context.Context, clusterName string, profileName string, suffix string) *appsv1.Deployment {
	name := clusterName
	if len(profileName) > 0 {
		name = fmt.Sprintf("%s-%s", clusterName, profileName)
	}

	if len(suffix) > 0 {
		name = fmt.Sprintf("%s-%s", name, suffix)
	}
	deployment := &appsv1.Deployment{}
	EventuallyWithOffset(1, func() error {
		if err := k8sClient.Get(
			ctx,
			types.NamespacedName{Name: name, Namespace: defaultNamespace},
			deployment,
		); err != nil {
			return err
		}
		return nil
	}, MapBuildingTimeout).Should(Succeed())
	return deployment
}

func pvc(ctx context.Context, clusterName string, profileName string, suffix string) *corev1.PersistentVolumeClaim {
	name := clusterName
	if len(profileName) > 0 {
		name = fmt.Sprintf("%s-%s", clusterName, profileName)
	}

	if len(suffix) > 0 {
		name = fmt.Sprintf("%s-%s", name, suffix)
	}
	pvc := &corev1.PersistentVolumeClaim{}
	EventuallyWithOffset(1, func() error {
		if err := k8sClient.Get(
			ctx,
			types.NamespacedName{Name: name, Namespace: defaultNamespace},
			pvc,
		); err != nil {
			return err
		}
		return nil
	}, MapBuildingTimeout).Should(Succeed())
	return pvc
}

func job(ctx context.Context, clusterName string, profileName string, suffix string) *batchv1.Job {
	name := clusterName
	if len(profileName) > 0 {
		name = fmt.Sprintf("%s-%s", clusterName, profileName)
	}

	if len(suffix) > 0 {
		name = fmt.Sprintf("%s-%s", name, suffix)
	}
	job := &batchv1.Job{}
	EventuallyWithOffset(1, func() error {
		if err := k8sClient.Get(
			ctx,
			types.NamespacedName{Name: name, Namespace: defaultNamespace},
			job,
		); err != nil {
			return err
		}
		return nil
	}, MapBuildingTimeout).Should(Succeed())
	return job
}

func pdb(ctx context.Context, clusterName string, profileName string, suffix string) *policyv1.PodDisruptionBudget {
	name := clusterName
	if len(profileName) > 0 {
		name = fmt.Sprintf("%s-%s", clusterName, profileName)
	}

	if len(suffix) > 0 {
		name = fmt.Sprintf("%s-%s", name, suffix)
	}
	pdb := &policyv1.PodDisruptionBudget{}
	EventuallyWithOffset(1, func() error {
		if err := k8sClient.Get(
			ctx,
			types.NamespacedName{Name: name, Namespace: defaultNamespace},
			pdb,
		); err != nil {
			return err
		}
		return nil
	}, MapBuildingTimeout).Should(Succeed())
	return pdb
}

func configMap(ctx context.Context, clusterName string, profileName string, suffix string) *corev1.ConfigMap {
	name := clusterName
	if len(profileName) > 0 {
		name = fmt.Sprintf("%s-%s", clusterName, profileName)
	}

	if len(profileName) > 0 {
		name = fmt.Sprintf("%s-%s", clusterName, profileName)
	}

	if len(suffix) > 0 {
		name = fmt.Sprintf("%s-%s", name, suffix)
	}
	configMap := &corev1.ConfigMap{}
	EventuallyWithOffset(1, func() error {
		if err := k8sClient.Get(
			ctx,
			types.NamespacedName{Name: name, Namespace: defaultNamespace},
			configMap,
		); err != nil {
			return err
		}
		return nil
	}, MapBuildingTimeout).Should(Succeed())
	return configMap
}
