package remover

import (
	"fmt"
	"testing"
)

func TestChunkObjectVersions(t *testing.T) {
	tests := []struct {
		name       string
		input      []ObjectVersion
		size       int
		wantChunks int
		wantSizes  []int
	}{
		{
			name:       "0 objects returns empty slice",
			input:      []ObjectVersion{},
			size:       1000,
			wantChunks: 0,
			wantSizes:  []int{},
		},
		{
			name:       "999 objects returns 1 chunk of 999",
			input:      makeObjects(999),
			size:       1000,
			wantChunks: 1,
			wantSizes:  []int{999},
		},
		{
			name:       "1000 objects returns 1 chunk of 1000",
			input:      makeObjects(1000),
			size:       1000,
			wantChunks: 1,
			wantSizes:  []int{1000},
		},
		{
			name:       "1001 objects returns 2 chunks of 1000 and 1",
			input:      makeObjects(1001),
			size:       1000,
			wantChunks: 2,
			wantSizes:  []int{1000, 1},
		},
		{
			name:       "2000 objects returns 2 chunks of 1000 and 1000",
			input:      makeObjects(2000),
			size:       1000,
			wantChunks: 2,
			wantSizes:  []int{1000, 1000},
		},
		{
			name:       "2001 objects returns 3 chunks of 1000, 1000 and 1",
			input:      makeObjects(2001),
			size:       1000,
			wantChunks: 3,
			wantSizes:  []int{1000, 1000, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := chunkObjectVersions(tt.input, tt.size)

			if len(got) != tt.wantChunks {
				t.Errorf("chunk count: got %d, want %d", len(got), tt.wantChunks)
			}

			for i, wantSize := range tt.wantSizes {
				if i >= len(got) {
					t.Errorf("chunk[%d] does not exist", i)
					continue
				}
				if len(got[i]) != wantSize {
					t.Errorf("chunk[%d] size: got %d, want %d", i, len(got[i]), wantSize)
				}
			}
		})
	}
}

// makeObjects generates a slice of n ObjectVersions for testing.
func makeObjects(n int) []ObjectVersion {
	objects := make([]ObjectVersion, n)
	for i := range objects {
		objects[i] = ObjectVersion{
			Key:       fmt.Sprintf("key-%d", i),
			VersionId: fmt.Sprintf("version-%d", i),
		}
	}
	return objects
}
