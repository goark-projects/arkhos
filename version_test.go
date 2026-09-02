package arkhos

import (
	"testing"

	"goark.dev/arkhos/hertz"

	servletcontainer "goark.dev/arkarta/servlet/container"
)

func TestNewReturnsDefaultContainer(t *testing.T) {
	target := New()
	metadata := target.Metadata()

	if metadata.Name() != hertz.Name {
		t.Fatalf("name = %q, want %q", metadata.Name(), hertz.Name)
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
	if transport := metadata.Limits()["transport"]; transport != "hertz" {
		t.Fatalf("default transport = %q, want hertz", transport)
	}
}
