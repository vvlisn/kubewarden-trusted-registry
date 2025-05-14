package main

import (
	"fmt"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	kubewarden "github.com/kubewarden/policy-sdk-go"
	kubewarden_protocol "github.com/kubewarden/policy-sdk-go/protocol"
	"github.com/tidwall/gjson"
)

const httpBadRequestStatusCode = 400

func validate(payload []byte) ([]byte, error) {
	if !gjson.ValidBytes(payload) {
		return kubewarden.RejectRequest(
			kubewarden.Message("invalid json payload"),
			kubewarden.Code(httpBadRequestStatusCode))
	}

	validationRequest := gjson.ParseBytes(payload)
	settings, err := NewSettingsFromValidationReq(&kubewarden_protocol.ValidationRequest{
		Settings: []byte(validationRequest.Get("settings").Raw),
	})
	if err != nil {
		return kubewarden.RejectRequest(
			kubewarden.Message(err.Error()),
			kubewarden.Code(httpBadRequestStatusCode))
	}

	// 获取容器列表
	containers := getContainers(validationRequest.Get("request.object.spec.containers"))
	if validationErr := validateContainers(containers, settings.TrustedRegistries); validationErr != nil {
		return kubewarden.RejectRequest(
			kubewarden.Message(validationErr.Error()),
			kubewarden.NoCode)
	}

	// 获取初始化容器列表
	initContainers := getContainers(validationRequest.Get("request.object.spec.initContainers"))
	if validationErr := validateContainers(initContainers, settings.TrustedRegistries); validationErr != nil {
		return kubewarden.RejectRequest(
			kubewarden.Message(validationErr.Error()),
			kubewarden.NoCode)
	}

	return kubewarden.AcceptRequest()
}

func getContainers(result gjson.Result) []string {
	var images []string
	result.ForEach(func(_, value gjson.Result) bool {
		if img := value.Get("image").String(); img != "" {
			images = append(images, img)
		}
		return true
	})
	return images
}

func validateContainers(containers []string, trustedRegistries mapset.Set[string]) error {
	for _, image := range containers {
		logger.Debug(fmt.Sprintf("Checking container image: %s", image))
		if !isImageTrusted(image, trustedRegistries) {
			logger.Error(fmt.Sprintf("Container image %s is not from a trusted registry", image))
			return fmt.Errorf("image '%s' is not from a trusted registry", image)
		}
		logger.Debug(fmt.Sprintf("Container image %s is from a trusted registry", image))
	}
	return nil
}

func isImageTrusted(image string, trustedRegistries mapset.Set[string]) bool {
	for _, registry := range trustedRegistries.ToSlice() {
		if strings.HasPrefix(image, registry) {
			return true
		}
	}
	return false
}
