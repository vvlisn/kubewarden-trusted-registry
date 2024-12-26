package main

import (
	"encoding/json"
	"testing"

	mapset "github.com/deckarep/golang-set/v2"
	corev1 "github.com/kubewarden/k8s-objects/api/core/v1"
	metav1 "github.com/kubewarden/k8s-objects/apimachinery/pkg/apis/meta/v1"
	kubewarden_protocol "github.com/kubewarden/policy-sdk-go/protocol"
	kubewarden_testing "github.com/kubewarden/policy-sdk-go/testing"
)

func TestIsImageTrusted(t *testing.T) {
	cases := []struct {
		podImages         []string
		trustedRegistries mapset.Set[string]
		expectedIsValid   bool
	}{
		{
			// ➀
			// Pod has no containers -> should be accepted
			podImages:         []string{},
			trustedRegistries: mapset.NewThreadUnsafeSet[string]("quay.io", "docker.io/library"),
			expectedIsValid:   true,
		},
		{
			// ➁
			// Pod has containers, all images are from trusted registries -> should be accepted
			podImages: []string{
				"quay.io/some/image",
				"docker.io/library/another/image",
			},
			trustedRegistries: mapset.NewThreadUnsafeSet[string]("quay.io", "docker.io/library"),
			expectedIsValid:   true,
		},
		{
			// ➂
			// Pod has containers, one image is not from a trusted registry -> should be rejected
			podImages: []string{
				"quay.io/some/image",
				"gcr.io/some/image",
			},
			trustedRegistries: mapset.NewThreadUnsafeSet[string]("quay.io", "docker.io/library"),
			expectedIsValid:   false,
		},
		{
			// ➃
			// Pod has containers, all images are from trusted registries with tags -> should be accepted
			podImages: []string{
				"quay.io/some/image:tag",
				"docker.io/library/another/image:tag",
			},
			trustedRegistries: mapset.NewThreadUnsafeSet[string]("quay.io", "docker.io/library"),
			expectedIsValid:   true,
		},
		{
			// ➄
			// Pod has containers, one image is not from a trusted registry with tag -> should be rejected
			podImages: []string{
				"quay.io/some/image:tag",
				"gcr.io/some/image:tag",
			},
			trustedRegistries: mapset.NewThreadUnsafeSet[string]("quay.io", "docker.io/library"),
			expectedIsValid:   false,
		},
		{
			// ➅
			// Pod has containers, all images are from trusted registries with SHA256 -> should be accepted
			podImages: []string{
				"quay.io/some/image@sha256:1234567890abcdef",
				"docker.io/library/another/image@sha256:1234567890abcdef",
			},
			trustedRegistries: mapset.NewThreadUnsafeSet[string]("quay.io", "docker.io/library"),
			expectedIsValid:   true,
		},
		{
			// ➆
			// Pod has containers, one image is not from a trusted registry with SHA256 -> should be rejected
			podImages: []string{
				"quay.io/some/image@sha256:1234567890abcdef",
				"gcr.io/some/image@sha256:1234567890abcdef",
			},
			trustedRegistries: mapset.NewThreadUnsafeSet[string]("quay.io", "docker.io/library"),
			expectedIsValid:   false,
		},
	}

	for _, testCase := range cases {
		settings := Settings{
			TrustedRegistries: testCase.trustedRegistries,
		}

		pod := corev1.Pod{
			Metadata: &metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "default",
			},
			Spec: &corev1.PodSpec{
				Containers: []*corev1.Container{},
			},
		}

		for _, image := range testCase.podImages {
			container := corev1.Container{
				Image: image,
			}
			pod.Spec.Containers = append(pod.Spec.Containers, &container)
		}

		payload, err := kubewarden_testing.BuildValidationRequest(&pod, &settings)
		if err != nil {
			t.Errorf("Unexpected error: %+v", err)
		}

		responsePayload, err := validate(payload)
		if err != nil {
			t.Errorf("Unexpected error: %+v", err)
		}

		var response kubewarden_protocol.ValidationResponse
		if unmarshalErr := json.Unmarshal(responsePayload, &response); unmarshalErr != nil {
			t.Errorf("Unexpected error: %+v", unmarshalErr)
		}

		if testCase.expectedIsValid && !response.Accepted {
			t.Errorf("Unexpected rejection: msg %s - code %d with pod images: %v, trusted registries: %v",
				*response.Message, *response.Code, testCase.podImages, testCase.trustedRegistries)
		}

		if !testCase.expectedIsValid && response.Accepted {
			t.Errorf("Unexpected acceptance with pod images: %v, trusted registries: %v",
				testCase.podImages, testCase.trustedRegistries)
		}
	}
}
