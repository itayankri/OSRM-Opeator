package resource

import (
	"fmt"

	"github.com/itayankri/OSRM-Operator/internal/metadata"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type GatewayServiceBuilder struct {
	ClusterScopedBuilder
	*OSRMResourceBuilder
}

func (builder *OSRMResourceBuilder) GatewayServiceBuilder(profiles []string) *GatewayServiceBuilder {
	return &GatewayServiceBuilder{
		ClusterScopedBuilder{profiles},
		builder,
	}
}

func (builder *GatewayServiceBuilder) Build() (client.Object, error) {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      builder.Instance.Name,
			Namespace: builder.Instance.Namespace,
		},
	}, nil
}

func (builder *GatewayServiceBuilder) Update(object client.Object) error {
	service := object.(*corev1.Service)

	service.Labels = metadata.GetLabels(builder.Instance.Name, builder.Instance.Labels)

	service.Spec.Type = corev1.ServiceTypeClusterIP
	service.Spec.Ports = []corev1.ServicePort{
		{
			Name:     fmt.Sprintf("%s-port", builder.Instance.Name),
			Protocol: corev1.ProtocolTCP,
			Port:     80,
			TargetPort: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: 80,
			},
		},
	}
	service.Spec.Selector = map[string]string{
		"app": fmt.Sprintf("%s-%s", builder.Instance.Name, gatewayPostfix),
	}

	if err := controllerutil.SetControllerReference(builder.Instance, service, builder.Scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %v", err)
	}

	return nil
}
