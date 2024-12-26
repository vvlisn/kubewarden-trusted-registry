package main

import (
	"encoding/json"
	"fmt"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	corev1 "github.com/kubewarden/k8s-objects/api/core/v1"
	kubewarden "github.com/kubewarden/policy-sdk-go"
	kubewarden_protocol "github.com/kubewarden/policy-sdk-go/protocol"
)

const httpBadRequestStatusCode = 400

func validate(payload []byte) ([]byte, error) {
	// Create a ValidationRequest instance from the incoming payload
	validationRequest := kubewarden_protocol.ValidationRequest{}
	err := json.Unmarshal(payload, &validationRequest)
	if err != nil {
		logger.Error(err.Error())
		return kubewarden.RejectRequest(
			kubewarden.Message(err.Error()),
			kubewarden.Code(httpBadRequestStatusCode))
	}

	// Create a Settings instance from the ValidationRequest object
	settings, err := NewSettingsFromValidationReq(&validationRequest)
	if err != nil {
		logger.Error(err.Error())
		return kubewarden.RejectRequest(
			kubewarden.Message(err.Error()),
			kubewarden.Code(httpBadRequestStatusCode))
	}

	// Access the **raw** JSON that describes the object
	podJSON := validationRequest.Request.Object

	// Debugging: Print raw JSON before unmarshalling
	logger.Debug(fmt.Sprintf("Raw pod JSON: %s", podJSON))

	// Try to create a Pod instance using the RAW JSON we got from the
	// ValidationRequest.
	pod := &corev1.Pod{}
	if err = json.Unmarshal([]byte(podJSON), pod); err != nil {
		logger.Error(fmt.Sprintf("Cannot decode Pod object: %s", err.Error()))
		return kubewarden.RejectRequest(
			kubewarden.Message(
				fmt.Sprintf("Cannot decode Pod object: %s", err.Error())),
			kubewarden.Code(httpBadRequestStatusCode))
	}

	// Debugging: Print the pod spec and trusted registries
	logger.Debug(fmt.Sprintf("Pod spec: %+v", pod.Spec))
	logger.Debug(fmt.Sprintf("Trusted registries: %v", settings.TrustedRegistries.ToSlice()))

	// Validate each container image in the Pod
	if pod.Spec != nil && pod.Spec.Containers != nil {
		for _, container := range pod.Spec.Containers {
			logger.Debug(fmt.Sprintf("Checking container image: %s", container.Image))
			if !isImageTrusted(container.Image, settings.TrustedRegistries) {
				logger.Error(fmt.Sprintf("Container image %s is not from a trusted registry", container.Image))
				return kubewarden.RejectRequest(
					kubewarden.Message(
						fmt.Sprintf("The image '%s' is not from a trusted registry", container.Image)),
					kubewarden.NoCode)
			} else {
				logger.Debug(fmt.Sprintf("Container image %s is from a trusted registry", container.Image))
			}
		}
	} else {
		logger.Info("No containers found in the Pod")
	}

	// Optionally, validate init containers if needed
	if pod.Spec != nil && pod.Spec.InitContainers != nil {
		for _, initContainer := range pod.Spec.InitContainers {
			logger.Debug(fmt.Sprintf("Checking init container image: %s", initContainer.Image))
			if !isImageTrusted(initContainer.Image, settings.TrustedRegistries) {
				logger.Error(fmt.Sprintf("Init container image %s is not from a trusted registry", initContainer.Image))
				return kubewarden.RejectRequest(
					kubewarden.Message(
						fmt.Sprintf("The init container image '%s' is not from a trusted registry", initContainer.Image)),
					kubewarden.NoCode)
			} else {
				logger.Debug(fmt.Sprintf("Init container image %s is from a trusted registry", initContainer.Image))
			}
		}
	} else {
		logger.Info("No init containers found in the Pod")
	}

	return kubewarden.AcceptRequest()
}

func isImageTrusted(image string, trustedRegistries mapset.Set[string]) bool {
	for _, registry := range trustedRegistries.ToSlice() {
		if strings.HasPrefix(image, registry) {
			return true
		}
	}
	return false
}
