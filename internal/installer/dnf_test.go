package installer

import (
	"testing"
)

func BenchmarkExpandRepoVariablesDnf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		expandRepoVariablesDnf("test string with ${DISTRO} and ${VERSION_ID}")
	}
}
