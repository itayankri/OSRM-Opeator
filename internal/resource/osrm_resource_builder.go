package resource

import (
	osrmv1alpha1 "github.com/itayankri/OSRM-Operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OSRMService string

type ResourceBuilder interface {
	Build() (client.Object, error)
	Update(client.Object) error
}

type OSRMResourceBuilder struct {
	Instance *osrmv1alpha1.OSRMCluster
	Scheme   *runtime.Scheme
}

type ProfileScopedBuilder struct {
	profile *osrmv1alpha1.ProfileSpec
}

type ClusterScopedBuilder struct {
	profiles []string
}

func (builder *OSRMResourceBuilder) ResourceBuilders() []ResourceBuilder {
	builders := []ResourceBuilder{}
	profilesEndpoints := []string{}
	profiles := []string{}

	for _, profile := range builder.Instance.Spec.Profiles {
		profilesEndpoints = append(profilesEndpoints, profile.EndpointName)
		profiles = append(profiles, profile.Name)
		builders = append(builders, []ResourceBuilder{
			builder.Deployment(&profile),
			builder.Service(&profile),
			builder.Job(&profile),
			builder.HorizontalPodAutoscaler(&profile),
			builder.PersistentVolumeClaim(&profile),
		}...)
	}

	if len(builders) > 0 {
		builders = append(builders, []ResourceBuilder{
			builder.ConfigMap(profiles, profilesEndpoints),
			builder.GatewayServiceBuilder(profiles),
			builder.GatewayDeployment(profiles),
		}...)
	}

	return builders
}
