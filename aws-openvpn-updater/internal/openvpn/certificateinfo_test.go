package openvpn

import (
	"reflect"
	"testing"
)

func TestCreateCertificateInfo(t *testing.T) {
	line := "V       330729111815Z                   612D9DE2717A9D139908E09C243F9ADA        unknown /CN=test"

	got := CreateCertificateInfo(line)
	want := &CertificateInfo{
		State: "V",
		Date:  "330729111815Z",
		Hash:  "612D9DE2717A9D139908E09C243F9ADA",
		Name:  "test",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func TestCreateCertificateInfoServer(t *testing.T) {
	line := "V       330729111815Z                   612D9DE2717A9D139908E09C243F9ADA        unknown /CN=server"

	got := CreateCertificateInfo(line)

	if got != nil {
		t.Errorf("got %q, wanted nil", got)
	}
}

func TestCreateCertificateInfoRevoked(t *testing.T) {
	line := "R       290429132915Z   200525133920Z   0C                                      unknown /CN=test"

	got := CreateCertificateInfo(line)

	if got != nil {
		t.Errorf("got %q, wanted nil", got)
	}
}
