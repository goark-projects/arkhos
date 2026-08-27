package arkhos

import (
	"testing"

	servletcontainer "goark.dev/arkarta/servlet/container"
)

func TestNewReturnsDefaultContainer(t *testing.T) {
	target := New()
	metadata := target.Metadata()

	if metadata.Name() != "arkhos" {
		t.Fatalf("name = %q, want arkhos", metadata.Name())
	}
	if metadata.Version() != Version {
		t.Fatalf("version = %q, want %q", metadata.Version(), Version)
	}
	if ArkartaVersion != "v0.0.1" {
		t.Fatalf("ArkartaVersion = %q, want v0.0.1", ArkartaVersion)
	}
	if !metadata.Supports(servletcontainer.ProfileCore) {
		t.Fatal("default container must support core profile")
	}
}
