package main

import (
	"encoding/json"
	"testing"

	mapset "github.com/deckarep/golang-set/v2"
)

func TestParsingSettingsWithNoValueProvided(t *testing.T) {
	rawSettings := []byte(`{}`)
	settings := &Settings{}
	if unmarshalErr := json.Unmarshal(rawSettings, settings); unmarshalErr != nil {
		t.Errorf("Unexpected error %+v", unmarshalErr)
	}

	if settings.TrustedRegistries.Cardinality() != 0 {
		t.Errorf("Expected TrustedRegistries to be empty")
	}

	valid, validationErr := settings.Valid()
	if valid {
		t.Errorf("Settings are reported as valid when no trusted registries are provided")
	}
	if validationErr == nil {
		t.Errorf("Expected an error when no trusted registries are provided")
	}
}

func TestParsingSettingsWithValidTrustedRegistries(t *testing.T) {
	rawSettings := []byte(`{"trusted_registries": ["quay.io", "docker.io/library"]}`)
	settings := &Settings{}
	if unmarshalErr := json.Unmarshal(rawSettings, settings); unmarshalErr != nil {
		t.Errorf("Unexpected error %+v", unmarshalErr)
	}

	if settings.TrustedRegistries.Cardinality() != 2 {
		t.Errorf("Expected TrustedRegistries to have 2 elements, got %d", settings.TrustedRegistries.Cardinality())
	}

	valid, validationErr := settings.Valid()
	if !valid {
		t.Errorf("Settings are reported as not valid")
	}
	if validationErr != nil {
		t.Errorf("Unexpected error %+v", validationErr)
	}
}

func TestParsingSettingsWithInvalidJSON(t *testing.T) {
	rawSettings := []byte(`{"trusted_registries": "not an array"}`)
	settings := &Settings{}
	if unmarshalErr := json.Unmarshal(rawSettings, settings); unmarshalErr == nil {
		t.Errorf("Expected an error for invalid JSON")
	}
}

func TestValidMethod(t *testing.T) {
	tests := []struct {
		settings Settings
		expected bool
	}{
		{Settings{TrustedRegistries: mapset.NewThreadUnsafeSet[string]()}, false},
		{Settings{TrustedRegistries: mapset.NewThreadUnsafeSet[string]("quay.io")}, true},
	}

	for _, test := range tests {
		valid, validationErr := test.settings.Valid()
		if valid != test.expected {
			t.Errorf("Expected Valid() to be %v, got %v", test.expected, valid)
		}
		if validationErr != nil && test.expected {
			t.Errorf("Unexpected error %+v", validationErr)
		}
	}
}
