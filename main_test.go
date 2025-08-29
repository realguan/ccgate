package main

import (
	"testing"
)

// TestPlatformValidate tests the Platform.Validate method
func TestPlatformValidate(t *testing.T) {
	// Test valid platform
	platform := Platform{
		Name:                "test",
		AnthropicBaseURL:    "https://api.test.com",
		AnthropicAuthToken:  "test-token",
		AnthropicModel:      "test-model",
		AnthropicSmallModel: "test-small-model",
	}

	if err := platform.Validate(); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test platform with missing name
	platform.Name = ""
	if err := platform.Validate(); err == nil {
		t.Error("Expected error for missing name, got nil")
	}

	// Test platform with missing URL
	platform.Name = "test"
	platform.AnthropicBaseURL = ""
	if err := platform.Validate(); err == nil {
		t.Error("Expected error for missing URL, got nil")
	}

	// Test platform with missing auth token
	platform.AnthropicBaseURL = "https://api.test.com"
	platform.AnthropicAuthToken = ""
	if err := platform.Validate(); err == nil {
		t.Error("Expected error for missing auth token, got nil")
	}

	// Test platform with missing model
	platform.AnthropicAuthToken = "test-token"
	platform.AnthropicModel = ""
	if err := platform.Validate(); err == nil {
		t.Error("Expected error for missing model, got nil")
	}
}

// TestConfigValidate tests the Config.Validate method
func TestConfigValidate(t *testing.T) {
	// Test valid config
	config := Config{
		Platforms: []Platform{
			{
				Name:                "test1",
				AnthropicBaseURL:    "https://api.test1.com",
				AnthropicAuthToken:  "test-token1",
				AnthropicModel:      "test-model1",
				AnthropicSmallModel: "test-small-model1",
			},
			{
				Name:                "test2",
				AnthropicBaseURL:    "https://api.test2.com",
				AnthropicAuthToken:  "test-token2",
				AnthropicModel:      "test-model2",
				AnthropicSmallModel: "test-small-model2",
			},
		},
	}

	if err := config.Validate(); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test config with no platforms
	config.Platforms = []Platform{}
	if err := config.Validate(); err == nil {
		t.Error("Expected error for empty platforms, got nil")
	}

	// Test config with invalid platform
	config.Platforms = []Platform{
		{
			Name: "",
		},
	}
	if err := config.Validate(); err == nil {
		t.Error("Expected error for invalid platform, got nil")
	}
}

// TestFindPlatformByName tests the findPlatformByName function
func TestFindPlatformByName(t *testing.T) {
	platforms := []Platform{
		{
			Name:                "platform1",
			AnthropicBaseURL:    "https://api.platform1.com",
			AnthropicAuthToken:  "token1",
			AnthropicModel:      "model1",
			AnthropicSmallModel: "small-model1",
		},
		{
			Name:                "platform2",
			AnthropicBaseURL:    "https://api.platform2.com",
			AnthropicAuthToken:  "token2",
			AnthropicModel:      "model2",
			AnthropicSmallModel: "small-model2",
		},
	}

	// Test finding existing platform
	platform, err := findPlatformByName(platforms, "platform1")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if platform.Name != "platform1" {
		t.Errorf("Expected platform1, got %s", platform.Name)
	}

	// Test finding non-existing platform
	_, err = findPlatformByName(platforms, "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existing platform, got nil")
	}
}
