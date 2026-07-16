package installer

import (
	"cloakpkg/internal/config"
	"testing"
)

func BenchmarkExpandRepoVariables(b *testing.B) {
	for i := 0; i < b.N; i++ {
		expandRepoVariables("test string with ${DISTRO} and ${VERSION_CODENAME} and ${ARCH}")
	}
}

func BenchmarkAptCheckNPlus1(b *testing.B) {
	apt := &Apt{}
	pkgs := []config.Package{
		{Name: "bash"},
		{Name: "curl"},
		{Name: "grep"},
		{Name: "awk"},
		{Name: "sed"},
		{Name: "tar"},
		{Name: "gzip"},
		{Name: "xz-utils"},
		{Name: "bzip2"},
		{Name: "zip"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, pkg := range pkgs {
			apt.Installed(pkg)
		}
	}
}

func BenchmarkAptCheckBatch(b *testing.B) {
	apt := &Apt{}
	pkgs := []config.Package{
		{Name: "bash"},
		{Name: "curl"},
		{Name: "grep"},
		{Name: "awk"},
		{Name: "sed"},
		{Name: "tar"},
		{Name: "gzip"},
		{Name: "xz-utils"},
		{Name: "bzip2"},
		{Name: "zip"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		apt.bulkInstalled(pkgs)
	}
}
