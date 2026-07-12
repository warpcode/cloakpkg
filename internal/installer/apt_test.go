package installer

import (
	"testing"
)

func BenchmarkExpandRepoVariables(b *testing.B) {
	for i := 0; i < b.N; i++ {
		expandRepoVariables("test string with ${DISTRO} and ${VERSION_CODENAME} and ${ARCH}")
	}
}
