package grimoire

import (
	"errors"
	"reflect"
	"testing"
)

func TestSortByChangeFrequency(t *testing.T) {
	tests := []struct {
		name         string
		files        []string
		commitCounts map[string]int
		counterErr   error
		want         []string
		wantErr      bool
	}{
		{
			name:       "already sorted",
			files:      []string{"file1", "file2", "file3"},
			counterErr: nil,
			commitCounts: map[string]int{
				"file1": 1,
				"file2": 2,
				"file3": 3,
			},
			want: []string{"file1", "file2", "file3"},
		},
		{
			name:       "reverse sorted",
			files:      []string{"file1", "file2", "file3"},
			counterErr: nil,
			commitCounts: map[string]int{
				"file1": 3,
				"file2": 2,
				"file3": 1,
			},
			want: []string{"file3", "file2", "file1"},
		},
		{
			name:       "unsorted",
			files:      []string{"file1", "file2", "file3"},
			counterErr: nil,
			commitCounts: map[string]int{
				"file1": 2,
				"file2": 1,
				"file3": 3,
			},
			want: []string{"file2", "file1", "file3"},
		},
		{
			name:         "empty",
			files:        []string{},
			counterErr:   nil,
			commitCounts: map[string]int{},
			want:         []string{},
		},
		{
			name:         "error from counter",
			files:        []string{"file1", "file2"},
			commitCounts: nil,
			counterErr:   errors.New("test error"),
			want:         nil,
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCounter := func(repoDir string) (map[string]int, error) {
				return tt.commitCounts, tt.counterErr
			}
			got, err := SortByChangeFrequency("test", tt.files, mockCounter)

			if (err != nil) != tt.wantErr {
				t.Errorf("SortByChangeFrequency() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SortByChangeFrequency() got = %v, want %v", got, tt.want)
			}
		})
	}
}
