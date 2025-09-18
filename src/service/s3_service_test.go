package service

import (
	"context"
	"testing"

	"github.com/denys89/wadugs-worker-cleansing/src/dto"
)

func TestNullS3Service_ListContractorFiles(t *testing.T) {
	service := NewNullS3Service()
	ctx := context.Background()

	files, err := service.ListContractorFiles(ctx, 1)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected empty slice, got %d files", len(files))
	}
}

func TestNullS3Service_ListProjectFiles(t *testing.T) {
	service := NewNullS3Service()
	ctx := context.Background()

	files, err := service.ListProjectFiles(ctx, 1)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected empty slice, got %d files", len(files))
	}
}

func TestNullS3Service_ListSiteFiles(t *testing.T) {
	service := NewNullS3Service()
	ctx := context.Background()

	files, err := service.ListSiteFiles(ctx, 1)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected empty slice, got %d files", len(files))
	}
}

func TestNullS3Service_DeleteObjects(t *testing.T) {
	service := NewNullS3Service()
	ctx := context.Background()

	tests := []struct {
		name    string
		objects []dto.S3Object
		want    int
	}{
		{
			name:    "Empty objects list",
			objects: []dto.S3Object{},
			want:    0,
		},
		{
			name: "Single object",
			objects: []dto.S3Object{
				{Bucket: "test-bucket", Key: "test-key"},
			},
			want: 1,
		},
		{
			name: "Multiple objects",
			objects: []dto.S3Object{
				{Bucket: "test-bucket", Key: "test-key-1"},
				{Bucket: "test-bucket", Key: "test-key-2"},
				{Bucket: "another-bucket", Key: "test-key-3"},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deleted, err := service.DeleteObjects(ctx, tt.objects)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if deleted != tt.want {
				t.Errorf("Expected %d deleted objects, got %d", tt.want, deleted)
			}
		})
	}
}

func TestS3Object_Validation(t *testing.T) {
	tests := []struct {
		name   string
		object dto.S3Object
		valid  bool
	}{
		{
			name:   "Valid object",
			object: dto.S3Object{Bucket: "test-bucket", Key: "test-key"},
			valid:  true,
		},
		{
			name:   "Empty bucket",
			object: dto.S3Object{Bucket: "", Key: "test-key"},
			valid:  false,
		},
		{
			name:   "Empty key",
			object: dto.S3Object{Bucket: "test-bucket", Key: ""},
			valid:  false,
		},
		{
			name:   "Both empty",
			object: dto.S3Object{Bucket: "", Key: ""},
			valid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.object.Bucket != "" && tt.object.Key != ""
			if isValid != tt.valid {
				t.Errorf("Expected validity %v, got %v", tt.valid, isValid)
			}
		})
	}
}

// Test helper functions that would be used in the real S3ServiceImpl
func TestS3ServiceHelpers(t *testing.T) {
	// Test batch size calculations
	objects := make([]dto.S3Object, 2500)
	for i := range objects {
		objects[i] = dto.S3Object{
			Bucket: "test-bucket",
			Key:    "test-key-" + string(rune(i)),
		}
	}

	// Test batch calculation
	batchSize := 1000
	expectedBatches := (len(objects) + batchSize - 1) / batchSize // Ceiling division
	actualBatches := 0

	for i := 0; i < len(objects); i += batchSize {
		actualBatches++
	}

	if actualBatches != expectedBatches {
		t.Errorf("Expected %d batches, got %d", expectedBatches, actualBatches)
	}
}

// Benchmark tests
func BenchmarkNullS3Service_ListContractorFiles(b *testing.B) {
	service := NewNullS3Service()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.ListContractorFiles(ctx, int64(i%1000+1))
	}
}

func BenchmarkNullS3Service_DeleteObjects(b *testing.B) {
	service := NewNullS3Service()
	ctx := context.Background()

	// Create test objects
	objects := make([]dto.S3Object, 100)
	for i := range objects {
		objects[i] = dto.S3Object{
			Bucket: "test-bucket",
			Key:    "test-key-" + string(rune(i)),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.DeleteObjects(ctx, objects)
	}
}

// Test concurrent operations
func TestNullS3Service_ConcurrentOperations(t *testing.T) {
	service := NewNullS3Service()
	ctx := context.Background()

	// Test concurrent list operations
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			_, err := service.ListContractorFiles(ctx, int64(id))
			if err != nil {
				t.Errorf("Concurrent operation failed: %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// Test edge cases
func TestS3Service_EdgeCases(t *testing.T) {
	service := NewNullS3Service()
	ctx := context.Background()

	// Test with very large ID
	_, err := service.ListContractorFiles(ctx, 9223372036854775807) // max int64
	if err != nil {
		t.Errorf("Should handle large IDs gracefully: %v", err)
	}

	// Test with zero ID
	_, err = service.ListContractorFiles(ctx, 0)
	if err != nil {
		t.Errorf("Should handle zero ID gracefully: %v", err)
	}

	// Test with negative ID
	_, err = service.ListContractorFiles(ctx, -1)
	if err != nil {
		t.Errorf("Should handle negative ID gracefully: %v", err)
	}
}