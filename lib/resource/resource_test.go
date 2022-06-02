package resource

import (
	"testing"
)

func BenchmarkNanoID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewNanoID(12)
	}
}
