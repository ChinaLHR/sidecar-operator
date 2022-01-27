package v1alpha1

import (
	"context"
	"encoding/json"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

//+kubebuilder:webhook:path=/mutate-sidecar-pod,mutating=true,failurePolicy=fail,groups="",resources=pods,verbs=create;update,versions=v1,name=mpod.kb.io,admissionReviewVersions=v1,sideEffects=None

type podHandler struct {
	Client  client.Client
	Decoder *admission.Decoder
}

var (
	SidecarEnvKey            = "IS_INJECTED"
	SidecarAnnotationHashKey = "chinalhr.github.io/sidecar-hash"
)

func (r *SidecarSet) Register(mgr ctrl.Manager) error {
	mgr.GetWebhookServer().Register("/mutate-sidecar-pod", &webhook.Admission{Handler: &podHandler{Client: mgr.GetClient()}})
	return nil
}

func (handler *podHandler) Handle(ctx context.Context, request admission.Request) admission.Response {
	log := log.FromContext(ctx)
	pod := &v1.Pod{}
	err := handler.Decoder.Decode(request, pod)
	copy := pod.DeepCopy()

	log.V(0).Info("begin to process mutating pod webhook")
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	log.V(0).Info("mutating pod webhook handle pod", "pod", pod.Name)
	err = handler.mutatingPodFunc(ctx, copy)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	marshaledPod, err := json.Marshal(copy)
	return admission.PatchResponseFromRaw(request.Object.Raw, marshaledPod)
}

func (handler *podHandler) mutatingPodFunc(ctx context.Context, pod *v1.Pod) error {
	//1. get sidecarSet resource list
	sidecarSets := &SidecarSetList{}
	if err := handler.Client.List(ctx, sidecarSets); err != nil {
		return err
	}

	var sidecarContainers []v1.Container
	sidecarSetHash := make(map[string]string)

	matchNothing := true
	for _, sidecarSet := range sidecarSets.Items {
		needInject, err := PodSidecarSetMatch(pod, sidecarSet)
		if err != nil {
			return err
		}

		if !needInject {
			continue
		}

		matchNothing = false
		sidecarSetHash[sidecarSet.Name] = sidecarSet.Annotations[SidecarAnnotationHashKey]

		for i := range sidecarSet.Spec.Containers {
			sidecarContainer := &sidecarSet.Spec.Containers[i]
			sidecarContainer.Env = append(sidecarContainer.Env, v1.EnvVar{Name: SidecarEnvKey, Value: "true"})
			sidecarContainers = append(sidecarContainers, sidecarContainer.Container)
		}
	}

	if matchNothing {
		return nil
	}

	pod.Spec.Containers = append(pod.Spec.Containers, sidecarContainers...)
	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string)
	}
	if len(sidecarSetHash) != 0 {
		encodedStr, err := json.Marshal(sidecarSetHash)
		if err != nil {
			return err
		}
		pod.Annotations[SidecarAnnotationHashKey] = string(encodedStr)
	}

	return nil
}

func PodSidecarSetMatch(pod *v1.Pod, sidecarSet SidecarSet) (bool, error) {
	selector, err := metav1.LabelSelectorAsSelector(sidecarSet.Spec.Selector)
	if err != nil {
		return false, err
	}

	if !selector.Empty() && selector.Matches(labels.Set(pod.Labels)) {
		return true, nil
	}
	return false, nil
}

// InjectDecoder injects the decoder.
func (handler *podHandler) InjectDecoder(d *admission.Decoder) error {
	handler.Decoder = d
	return nil
}
