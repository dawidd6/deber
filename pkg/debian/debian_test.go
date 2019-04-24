package debian_test

import (
	"github.com/dawidd6/deber/pkg/debian"
	"testing"
)

func test(t *testing.T, got, expected *debian.Debian) {
	if *got != *expected {
		t.Log("got:", got)
		t.Log("expected:", expected)
		t.Fatal("got != expected")
	}
}

func TestNew1(t *testing.T) {
	line := "golang-github-alcortesm-tgz (0.0~git20161220.9c5fe88-1) unstable; urgency=medium"

	got := debian.New(line)
	expected := &debian.Debian{
		SourceName:      "golang-github-alcortesm-tgz",
		PackageVersion:  "0.0~git20161220.9c5fe88-1",
		UpstreamVersion: "0.0~git20161220.9c5fe88",
		TargetDist:      "unstable",
		IsNative:        false,
	}

	test(t, got, expected)
}

func TestNew2(t *testing.T) {
	line := "ansible (2.7.5+dfsg-2) experimental; urgency=medium"

	got := debian.New(line)
	expected := &debian.Debian{
		SourceName:      "ansible",
		PackageVersion:  "2.7.5+dfsg-2",
		UpstreamVersion: "2.7.5+dfsg",
		TargetDist:      "experimental",
		IsNative:        false,
	}

	test(t, got, expected)
}

func TestNew3(t *testing.T) {
	line := "ansible (2.2.1.0-2+deb9u1) stretch-security; urgency=high"

	got := debian.New(line)
	expected := &debian.Debian{
		SourceName:      "ansible",
		PackageVersion:  "2.2.1.0-2+deb9u1",
		UpstreamVersion: "2.2.1.0",
		TargetDist:      "stretch",
		IsNative:        false,
	}

	test(t, got, expected)
}

func TestNew4(t *testing.T) {
	line := "procps (2:3.3.10-4ubuntu2.4) xenial-security; urgency=medium"

	got := debian.New(line)
	expected := &debian.Debian{
		SourceName:      "procps",
		PackageVersion:  "2:3.3.10-4ubuntu2.4",
		UpstreamVersion: "3.3.10",
		TargetDist:      "xenial",
		IsNative:        false,
	}

	test(t, got, expected)
}

func TestNew5(t *testing.T) {
	line := "git-buildpackage (0.9.14) unstable; urgency=medium"

	got := debian.New(line)
	expected := &debian.Debian{
		SourceName:      "git-buildpackage",
		PackageVersion:  "0.9.14",
		UpstreamVersion: "0.9.14",
		TargetDist:      "unstable",
		IsNative:        true,
	}

	test(t, got, expected)
}

func TestNew6(t *testing.T) {
	line := "ansible (2.7.5+dfsg-1~bpo9+1) stretch-backports; urgency=medium"

	got := debian.New(line)
	expected := &debian.Debian{
		SourceName:      "ansible",
		PackageVersion:  "2.7.5+dfsg-1~bpo9+1",
		UpstreamVersion: "2.7.5+dfsg",
		TargetDist:      "stretch-backports",
		IsNative:        false,
	}

	test(t, got, expected)
}

func TestNew7(t *testing.T) {
	line := "procps (2:3.3.10-4ubuntu2.4) xenial-backports; urgency=medium"

	got := debian.New(line)
	expected := &debian.Debian{
		SourceName:      "procps",
		PackageVersion:  "2:3.3.10-4ubuntu2.4",
		UpstreamVersion: "3.3.10",
		TargetDist:      "xenial",
		IsNative:        false,
	}

	test(t, got, expected)
}
