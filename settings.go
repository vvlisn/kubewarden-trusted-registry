package main

import (
	"encoding/json"
	"errors"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	kubewarden "github.com/kubewarden/policy-sdk-go"
	kubewarden_protocol "github.com/kubewarden/policy-sdk-go/protocol"
)

type Settings struct {
	TrustedRegistries mapset.Set[string] `json:"trusted_registries"`
}

func (s *Settings) UnmarshalJSON(data []byte) error {
	rawSettings := struct {
		TrustedRegistries []string `json:"trusted_registries"`
	}{}

	err := json.Unmarshal(data, &rawSettings)
	if err != nil {
		return err
	}

	s.TrustedRegistries = mapset.NewThreadUnsafeSet[string](rawSettings.TrustedRegistries...)

	return nil
}

func NewSettingsFromValidationReq(validationReq *kubewarden_protocol.ValidationRequest) (Settings, error) {
	settings := Settings{}
	err := json.Unmarshal(validationReq.Settings, &settings)
	if err != nil {
		return Settings{}, err
	}

	return settings, nil
}

func (s *Settings) Valid() (bool, error) {
	if s.TrustedRegistries.Cardinality() == 0 {
		return false, errors.New("no trusted registries provided")
	}

	return true, nil
}

func validateSettings(payload []byte) ([]byte, error) {
	settings := Settings{}
	err := json.Unmarshal(payload, &settings)
	if err != nil {
		return kubewarden.RejectSettings(
			kubewarden.Message(fmt.Sprintf("Provided settings are not valid: %v", err)))
	}

	valid, err := settings.Valid()
	if valid {
		return kubewarden.AcceptSettings()
	}

	return kubewarden.RejectSettings(
		kubewarden.Message(fmt.Sprintf("Provided settings are not valid: %v", err)))
}
