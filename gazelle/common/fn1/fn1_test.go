package fn1

import (
	"fmt"
	"iter"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func ExampleMap() {
	numbers := slices.Values([]int{1, 2, 3, 4})
	squared := Map(numbers, func(n int) int { return n * n })

	for next := range squared {
		fmt.Println(next)
	}

	// Output:
	// 1
	// 4
	// 9
	// 16
}

func TestAssociateBy(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	tests := []struct {
		name        string
		input       iter.Seq[Person]
		keySelector func(Person) string
		want        map[string]Person
		wantErr     bool
	}{
		{
			name:  "Simple Association",
			input: slices.Values([]Person{{"Alice", 30}, {"Bob", 25}}),
			keySelector: func(p Person) string {
				return p.Name
			},
			want: map[string]Person{
				"Alice": {Name: "Alice", Age: 30},
				"Bob":   {Name: "Bob", Age: 25},
			},
			wantErr: false,
		},
		{
			name:  "Duplicate Keys",
			input: slices.Values([]Person{{"Alice", 30}, {"Bob", 25}, {"Alice", 28}}),
			keySelector: func(p Person) string {
				return p.Name
			},
			want:    nil,
			wantErr: true, // Expecting a panic due to duplicate keys
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); (r != nil) != tt.wantErr {
					if tt.wantErr {
						t.Errorf("AssociateBy() did not panic as expected")
					} else {
						t.Errorf("AssociateBy() panicked unexpectedly: %v", r)
					}
				}
			}()

			got := AssociateBy(tt.input, tt.keySelector)
			if !tt.wantErr && !cmp.Equal(got, tt.want) {
				t.Errorf("AssociateBy() = %v, want %v", got, tt.want)
			}
		})
	}
}
