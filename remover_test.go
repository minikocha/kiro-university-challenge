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
			name:       "0件は空スライスを返す",
			input:      []ObjectVersion{},
			size:       1000,
			wantChunks: 0,
			wantSizes:  []int{},
		},
		{
			name:       "999件は1チャンク(999)に分割される",
			input:      makeObjects(999),
			size:       1000,
			wantChunks: 1,
			wantSizes:  []int{999},
		},
		{
			name:       "1000件はちょうど1チャンク(1000)に分割される",
			input:      makeObjects(1000),
			size:       1000,
			wantChunks: 1,
			wantSizes:  []int{1000},
		},
		{
			name:       "1001件は2チャンク(1000, 1)に分割される",
			input:      makeObjects(1001),
			size:       1000,
			wantChunks: 2,
			wantSizes:  []int{1000, 1},
		},
		{
			name:       "2000件はちょうど2チャンク(1000, 1000)に分割される",
			input:      makeObjects(2000),
			size:       1000,
			wantChunks: 2,
			wantSizes:  []int{1000, 1000},
		},
		{
			name:       "2001件は3チャンク(1000, 1000, 1)に分割される",
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
				t.Errorf("チャンク数: got %d, want %d", len(got), tt.wantChunks)
			}

			for i, wantSize := range tt.wantSizes {
				if i >= len(got) {
					t.Errorf("チャンク[%d]が存在しません", i)
					continue
				}
				if len(got[i]) != wantSize {
					t.Errorf("チャンク[%d]のサイズ: got %d, want %d", i, len(got[i]), wantSize)
				}
			}
		})
	}
}

// makeObjects はテスト用にn件のObjectVersionスライスを生成する。
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
