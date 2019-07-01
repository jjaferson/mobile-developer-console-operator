package mobiledeveloperconsole

import (
	"fmt"
	"github.com/aerogear/mobile-developer-console-operator/pkg/util"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/intstr"

	mdcv1alpha1 "github.com/aerogear/mobile-developer-console-operator/pkg/apis/mdc/v1alpha1"
	openshiftappsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newMDCServiceAccount(cr *mdcv1alpha1.MobileDeveloperConsole) (*corev1.ServiceAccount, error) {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Annotations: map[string]string{
				"serviceaccounts.openshift.io/oauth-redirectreference.mdc": fmt.Sprintf("{\"kind\":\"OAuthRedirectReference\",\"apiVersion\":\"v1\",\"reference\":{\"kind\":\"Route\",\"name\":\"%s-mdc-proxy\"}}", cr.Name),
			},
		},
	}, nil
}

func newOauthProxyService(cr *mdcv1alpha1.MobileDeveloperConsole) (*corev1.Service, error) {
	return &corev1.Service{
		ObjectMeta: util.ObjectMeta(&cr.ObjectMeta, "mdc-proxy"),
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":     cr.Name,
				"service": "mdc",
			},
			Ports: []corev1.ServicePort{
				{
					Name:     "web",
					Protocol: corev1.ProtocolTCP,
					Port:     80,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 4180,
					},
				},
			},
		},
	}, nil
}

func newMDCService(cr *mdcv1alpha1.MobileDeveloperConsole) (*corev1.Service, error) {
	serviceObjectMeta := util.ObjectMeta(&cr.ObjectMeta, "mdc")
	serviceObjectMeta.Labels["mobile"] = "enabled"
	serviceObjectMeta.Labels["internal"] = "mdc"

	return &corev1.Service{
		ObjectMeta: serviceObjectMeta,
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":     cr.Name,
				"service": "mdc",
			},
			Ports: []corev1.ServicePort{
				{
					Name:     "web",
					Protocol: corev1.ProtocolTCP,
					Port:     80,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 4000,
					},
				},
			},
		},
	}, nil
}

func newOauthProxyRoute(cr *mdcv1alpha1.MobileDeveloperConsole) (*routev1.Route, error) {
	return &routev1.Route{
		ObjectMeta: util.ObjectMeta(&cr.ObjectMeta, "mdc-proxy"),
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: fmt.Sprintf("%s-%s", cr.Name, "mdc-proxy"),
			},
			TLS: &routev1.TLSConfig{
				Termination:                   routev1.TLSTerminationEdge,
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyNone,
			},
		},
	}, nil
}

func newOauthProxyImageStream(cr *mdcv1alpha1.MobileDeveloperConsole) (*imagev1.ImageStream, error) {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cr.Namespace,
			Name:      cfg.OauthProxyImageStreamName,
			Labels:    util.Labels(&cr.ObjectMeta, cfg.OauthProxyImageStreamName),
		},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				{
					Name: cfg.OauthProxyImageStreamTag,
					From: &corev1.ObjectReference{
						Kind: "DockerImage",
						Name: cfg.OauthProxyImageStreamInitialImage,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						Scheduled: false,
					},
				},
			},
		},
	}, nil
}

func newMDCImageStream(cr *mdcv1alpha1.MobileDeveloperConsole) (*imagev1.ImageStream, error) {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cr.Namespace,
			Name:      cfg.MDCImageStreamName,
			Labels:    util.Labels(&cr.ObjectMeta, cfg.MDCImageStreamName),
		},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				{
					Name: cfg.MDCImageStreamTag,
					From: &corev1.ObjectReference{
						Kind: "DockerImage",
						Name: cfg.MDCImageStreamInitialImage,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						Scheduled: false,
					},
				},
			},
		},
	}, nil
}

func newMDCDeploymentConfig(cr *mdcv1alpha1.MobileDeveloperConsole) (*openshiftappsv1.DeploymentConfig, error) {
	labels := map[string]string{
		"app":     cr.Name,
		"service": "mdc",
	}

	cookieSecret, err := util.GeneratePassword()
	if err != nil {
		return nil, errors.Wrap(err, "error generating cookie secret")
	}

	return &openshiftappsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: openshiftappsv1.DeploymentConfigSpec{
			Replicas: 1,
			Selector: labels,
			Triggers: openshiftappsv1.DeploymentTriggerPolicies{
				openshiftappsv1.DeploymentTriggerPolicy{
					Type: openshiftappsv1.DeploymentTriggerOnConfigChange,
				},
				openshiftappsv1.DeploymentTriggerPolicy{
					Type: openshiftappsv1.DeploymentTriggerOnImageChange,
					ImageChangeParams: &openshiftappsv1.DeploymentTriggerImageChangeParams{
						Automatic:      true,
						ContainerNames: []string{cfg.MDCContainerName},
						From: corev1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: cfg.MDCImageStreamName + ":" + cfg.MDCImageStreamTag,
						},
					},
				},
				openshiftappsv1.DeploymentTriggerPolicy{
					Type: openshiftappsv1.DeploymentTriggerOnImageChange,
					ImageChangeParams: &openshiftappsv1.DeploymentTriggerImageChangeParams{
						Automatic:      true,
						ContainerNames: []string{cfg.OauthProxyContainerName},
						From: corev1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: cfg.OauthProxyImageStreamName + ":" + cfg.OauthProxyImageStreamTag,
						},
					},
				},
			},
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: cr.Name,
					Containers: []corev1.Container{
						{
							Name:            cfg.MDCContainerName,
							Image:           cfg.MDCImageStreamName + ":" + cfg.MDCImageStreamTag,
							ImagePullPolicy: corev1.PullAlways,
							Env: []corev1.EnvVar{
								{
									Name:  "NAMESPACE",
									Value: cr.Namespace,
								},
								{
									Name:  "NODE_ENV",
									Value: "production",
								},
								{
									Name:  "OPENSHIFT_HOST",
									Value: cfg.OpenShiftHost,
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          cfg.MDCContainerName,
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 4000,
								},
							},
						},
						{
							Name:            cfg.OauthProxyContainerName,
							Image:           cfg.OauthProxyImageStreamName + ":" + cfg.OauthProxyImageStreamTag,
							ImagePullPolicy: corev1.PullAlways,
							Ports: []corev1.ContainerPort{
								{
									Name:          "public",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 4180,
								},
							},
							Args: []string{
								"--provider=openshift",
								fmt.Sprintf("--openshift-service-account=%s", cr.Name),
								"--upstream=http://localhost:4000",
								"--http-address=0.0.0.0:4180",
								"--skip-auth-regex=/rest/sender,/rest/registry/device,/rest/prometheus/metrics,/rest/auth/config",
								"--https-address=",
								fmt.Sprintf("--cookie-secret=%s", cookieSecret),
							},
						},
					},
				},
			},
		},
	}, nil
}
