package nethttp

import (
	"testing"

	"goark.dev/arkarta/servlet/container"
	"goark.dev/arkarta/servlet/tck"
)

func TestContainerHTTP(t *testing.T) {
	tck.RunHTTPContainer(t, func() tck.HTTPContainer {
		return NewContainer()
	})
}

func TestContainerMetadata(t *testing.T) {
	target := NewContainer()
	metadata := target.Metadata()

	if metadata.Name() != "arkhos" {
		t.Fatalf("name = %q, want arkhos", metadata.Name())
	}
	if metadata.Version() != Version {
		t.Fatalf("version = %q, want %q", metadata.Version(), Version)
	}
	if !metadata.Supports(container.ProfileCore) {
		t.Fatal("container must support core profile")
	}

	profiles := metadata.Profiles()
	profiles[0] = "mutated"
	if !metadata.Supports(container.ProfileCore) {
		t.Fatal("metadata profiles must be immutable snapshots")
	}
}
