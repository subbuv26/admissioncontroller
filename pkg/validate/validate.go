package validate

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	admission "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecFactory  = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecFactory.UniversalDeserializer()
)

type Validator interface {
	Validate(admission.AdmissionReview) *admission.AdmissionResponse
}

type validator struct {
	cli *kubernetes.Clientset
}

func New() (Validator, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &validator{
		cli: clientset,
	}, nil
}

func (v *validator) Validate(ar admission.AdmissionReview) *admission.AdmissionResponse {
	slog.Info("validating deployments")
	deploymentResource := metav1.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	if ar.Request.Resource != deploymentResource {
		slog.Error(fmt.Sprintf("expect resource to be %s", deploymentResource))
		return nil
	}
	raw := ar.Request.Object.Raw
	deployment := appsv1.Deployment{}
	if _, _, err := deserializer.Decode(raw, nil, &deployment); err != nil {
		slog.Error("failed to decode deployment", "error", err)
		return &admission.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}
	unavailableSecrets := getUnavailableSecrets(&deployment)
	if len(unavailableSecrets) != 0 {
		message := "secrets unavailable: " + strings.Join(unavailableSecrets, ", ")
		return &admission.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: message,
			},
			UID: ar.Request.UID,
		}
	}

	return &admission.AdmissionResponse{Allowed: true}
}

func getUnavailableSecrets(deployment *appsv1.Deployment) []string {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	var unavailableSecrets []string
	secretsClient := clientset.CoreV1().Secrets(deployment.Namespace)

	for _, vol := range deployment.Spec.Template.Spec.Volumes {
		if vol.Secret != nil {
			_, err := secretsClient.Get(context.TODO(), vol.Secret.SecretName, metav1.GetOptions{})
			if err != nil {
				unavailableSecrets = append(unavailableSecrets, vol.Secret.SecretName)
			}
		}
	}
	return unavailableSecrets
}
