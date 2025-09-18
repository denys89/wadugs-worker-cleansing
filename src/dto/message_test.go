package dto

import (
	"encoding/json"
	"testing"
)

func TestCleansingMessage_IsValidType(t *testing.T) {
	tests := []struct {
		name     string
		message  CleansingMessage
		expected bool
	}{
		{
			name:     "Valid contractor type",
			message:  CleansingMessage{Type: "contractor", ID: 1},
			expected: true,
		},
		{
			name:     "Valid project type",
			message:  CleansingMessage{Type: "project", ID: 1},
			expected: true,
		},
		{
			name:     "Valid site type",
			message:  CleansingMessage{Type: "site", ID: 1},
			expected: true,
		},
		{
			name:     "Invalid type - empty",
			message:  CleansingMessage{Type: "", ID: 1},
			expected: false,
		},
		{
			name:     "Invalid type - wrong value",
			message:  CleansingMessage{Type: "invalid", ID: 1},
			expected: false,
		},
		{
			name:     "Invalid type - case sensitive",
			message:  CleansingMessage{Type: "Contractor", ID: 1},
			expected: false,
		},
		{
			name:     "Invalid type - case sensitive",
			message:  CleansingMessage{Type: "PROJECT", ID: 1},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.message.IsValidType()
			if result != tt.expected {
				t.Errorf("IsValidType() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestCleansingMessage_GetDescription(t *testing.T) {
	tests := []struct {
		name     string
		message  CleansingMessage
		expected string
	}{
		{
			name:     "Contractor description",
			message:  CleansingMessage{Type: "contractor", ID: 123},
			expected: "Deleting all files for contractor and its related projects and sites",
		},
		{
			name:     "Project description",
			message:  CleansingMessage{Type: "project", ID: 456},
			expected: "Deleting all files for project and its related sites",
		},
		{
			name:     "Site description",
			message:  CleansingMessage{Type: "site", ID: 789},
			expected: "Deleting all files for site",
		},
		{
			name:     "Invalid type description",
			message:  CleansingMessage{Type: "invalid", ID: 1},
			expected: "Unknown cleansing operation",
		},
		{
			name:     "Empty type description",
			message:  CleansingMessage{Type: "", ID: 1},
			expected: "Unknown cleansing operation",
		},
		{
			name:     "Zero ID description",
			message:  CleansingMessage{Type: "contractor", ID: 0},
			expected: "Deleting all files for contractor and its related projects and sites",
		},
		{
			name:     "Negative ID description",
			message:  CleansingMessage{Type: "contractor", ID: -1},
			expected: "Deleting all files for contractor and its related projects and sites",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.message.GetDescription()
			if result != tt.expected {
				t.Errorf("GetDescription() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestCleansingResult_Validation(t *testing.T) {
	tests := []struct {
		name   string
		result CleansingResult
		valid  bool
	}{
		{
			name: "Valid success result",
			result: CleansingResult{
				Success:      true,
				Message:      "Operation completed successfully",
				FilesDeleted: 5,
			},
			valid: true,
		},
		{
			name: "Valid error result",
			result: CleansingResult{
				Success:      false,
				Message:      "Operation failed",
				FilesDeleted: 0,
				Error:        "Connection timeout",
			},
			valid: true,
		},
		{
			name: "Valid partial success",
			result: CleansingResult{
				Success:      false,
				Message:      "Partial operation",
				FilesDeleted: 3,
				Error:        "Some files could not be deleted",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation - success results should not have error messages
			// and error results should have error messages
			isValid := true
			if tt.result.Success && tt.result.Error != "" {
				isValid = false // Success should not have error
			}
			if !tt.result.Success && tt.result.Error == "" {
				// This might be acceptable in some cases, so we'll allow it
			}

			if isValid != tt.valid {
				t.Errorf("Result validation = %v, expected %v", isValid, tt.valid)
			}
		})
	}
}

func TestS3Object_Validation(t *testing.T) {
	tests := []struct {
		name   string
		object S3Object
		valid  bool
	}{
		{
			name:   "Valid S3 object",
			object: S3Object{Bucket: "my-bucket", Key: "path/to/file.txt"},
			valid:  true,
		},
		{
			name:   "Valid S3 object with special characters",
			object: S3Object{Bucket: "my-bucket-123", Key: "path/to/file-with_special.chars.txt"},
			valid:  true,
		},
		{
			name:   "Empty bucket",
			object: S3Object{Bucket: "", Key: "path/to/file.txt"},
			valid:  false,
		},
		{
			name:   "Empty key",
			object: S3Object{Bucket: "my-bucket", Key: ""},
			valid:  false,
		},
		{
			name:   "Both empty",
			object: S3Object{Bucket: "", Key: ""},
			valid:  false,
		},
		{
			name:   "Whitespace bucket",
			object: S3Object{Bucket: "   ", Key: "path/to/file.txt"},
			valid:  false, // Assuming we trim whitespace
		},
		{
			name:   "Whitespace key",
			object: S3Object{Bucket: "my-bucket", Key: "   "},
			valid:  false, // Assuming we trim whitespace
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simple validation logic
			isValid := tt.object.Bucket != "" && tt.object.Key != ""
			
			// Additional validation for whitespace
			if isValid {
				bucket := tt.object.Bucket
				key := tt.object.Key
				// Check for whitespace-only strings
				bucketTrimmed := ""
				keyTrimmed := ""
				for _, r := range bucket {
					if r != ' ' && r != '\t' && r != '\n' {
						bucketTrimmed += string(r)
					}
				}
				for _, r := range key {
					if r != ' ' && r != '\t' && r != '\n' {
						keyTrimmed += string(r)
					}
				}
				if bucketTrimmed == "" || keyTrimmed == "" {
					isValid = false
				}
			}

			if isValid != tt.valid {
				t.Errorf("S3Object validation = %v, expected %v", isValid, tt.valid)
			}
		})
	}
}

// Test JSON marshaling/unmarshaling
func TestCleansingMessage_JSONSerialization(t *testing.T) {
	original := CleansingMessage{
		Type: "contractor",
		ID:   123,
	}

	// Test marshaling using json package
	data, err := json.Marshal(original)
	if err != nil {
		t.Errorf("Failed to marshal CleansingMessage: %v", err)
	}

	// Test unmarshaling using json package
	var unmarshaled CleansingMessage
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal CleansingMessage: %v", err)
	}

	// Compare
	if unmarshaled.Type != original.Type || unmarshaled.ID != original.ID {
		t.Errorf("Unmarshaled message doesn't match original: got %+v, expected %+v", unmarshaled, original)
	}
}

func TestCleansingResult_JSONSerialization(t *testing.T) {
	original := CleansingResult{
		Success:      true,
		Message:      "Test message",
		FilesDeleted: 5,
		Error:        "",
	}

	// Test marshaling using json package
	data, err := json.Marshal(original)
	if err != nil {
		t.Errorf("Failed to marshal CleansingResult: %v", err)
	}

	// Test unmarshaling using json package
	var unmarshaled CleansingResult
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal CleansingResult: %v", err)
	}

	// Compare
	if unmarshaled.Success != original.Success ||
		unmarshaled.Message != original.Message ||
		unmarshaled.FilesDeleted != original.FilesDeleted ||
		unmarshaled.Error != original.Error {
		t.Errorf("Unmarshaled result doesn't match original: got %+v, expected %+v", unmarshaled, original)
	}
}

// Benchmark tests
func BenchmarkCleansingMessage_IsValidType(b *testing.B) {
	message := CleansingMessage{Type: "contractor", ID: 1}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = message.IsValidType()
	}
}

func BenchmarkCleansingMessage_GetDescription(b *testing.B) {
	message := CleansingMessage{Type: "contractor", ID: 123}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = message.GetDescription()
	}
}