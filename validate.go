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
	// 解析 ValidationRequest
	validationRequest, err := parseValidationRequest(payload)
	if err != nil {
		return kubewarden.RejectRequest(
			kubewarden.Message(err.Error()),
			kubewarden.Code(httpBadRequestStatusCode))
	}

	// 从 ValidationRequest 获取配置
	settings, err := NewSettingsFromValidationReq(validationRequest)
	if err != nil {
		return kubewarden.RejectRequest(
			kubewarden.Message(err.Error()),
			kubewarden.Code(httpBadRequestStatusCode))
	}

	// 解析 Pod 对象
	pod, err := parsePod(validationRequest.Request.Object)
	if err != nil {
		return kubewarden.RejectRequest(
			kubewarden.Message(err.Error()),
			kubewarden.Code(httpBadRequestStatusCode))
	}
	// 将 []*corev1.Container 转换为 []corev1.Container
	var containers []corev1.Container
	if pod.Spec.Containers != nil {
		for _, container := range pod.Spec.Containers {
			containers = append(containers, *container)
		}
	}
	// 验证 Pod 中的容器镜像
	if err := validateContainers(containers, settings.TrustedRegistries); err != nil {
		return kubewarden.RejectRequest(
			kubewarden.Message(err.Error()),
			kubewarden.NoCode)
	}
	// 将 []*corev1.Container 转换为 []corev1.Container
	var initContainers []corev1.Container
	if pod.Spec.InitContainers != nil {
		for _, initContainer := range pod.Spec.InitContainers {
			initContainers = append(initContainers, *initContainer)
		}
	}
	// 验证 Pod 中的初始化容器镜像
	if err := validateContainers(initContainers, settings.TrustedRegistries); err != nil {
		return kubewarden.RejectRequest(
			kubewarden.Message(err.Error()),
			kubewarden.NoCode)
	}

	// 所有验证通过，接受请求
	return kubewarden.AcceptRequest()
}

// 解析验证请求
func parseValidationRequest(payload []byte) (*kubewarden_protocol.ValidationRequest, error) {
	validationRequest := &kubewarden_protocol.ValidationRequest{}
	if err := json.Unmarshal(payload, validationRequest); err != nil {
		logger.Error("Failed to parse validation request: " + err.Error())
		return nil, fmt.Errorf("invalid validation request: %w", err)
	}
	return validationRequest, nil
}

// 解析 Pod 对象
func parsePod(podJSON json.RawMessage) (*corev1.Pod, error) {
	pod := &corev1.Pod{}
	if err := json.Unmarshal([]byte(podJSON), pod); err != nil {
		logger.Error("Failed to parse Pod object: " + err.Error())
		return nil, fmt.Errorf("invalid Pod object: %w", err)
	}
	return pod, nil
}

// 验证容器镜像是否来自受信任的仓库
func validateContainers(containers []corev1.Container, trustedRegistries mapset.Set[string]) error {
	for _, container := range containers {
		logger.Debug(fmt.Sprintf("Checking container image: %s", container.Image))
		if !isImageTrusted(container.Image, trustedRegistries) {
			logger.Error(fmt.Sprintf("Container image %s is not from a trusted registry", container.Image))
			return fmt.Errorf("image '%s' is not from a trusted registry", container.Image)
		}
		logger.Debug(fmt.Sprintf("Container image %s is from a trusted registry", container.Image))
	}
	return nil
}

// 判断镜像是否来自受信任的仓库
func isImageTrusted(image string, trustedRegistries mapset.Set[string]) bool {
	for _, registry := range trustedRegistries.ToSlice() {
		if strings.HasPrefix(image, registry) {
			return true
		}
	}
	return false
}
