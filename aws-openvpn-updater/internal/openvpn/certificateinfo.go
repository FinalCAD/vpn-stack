package openvpn

import (
	"fmt"
	"strings"
)

type CertificateInfo struct {
	State string
	Date  string
	Hash  string
	Name  string
}

func (c CertificateInfo) String() string {
	return fmt.Sprintf("[ Name: %s ]", c.Name)
}

func CreateCertificateInfo(line string) *CertificateInfo {
	fields := strings.Fields(line)
	if len(fields) != 5 {
		return nil
	}

	name := strings.TrimPrefix(fields[4], "/CN=")

	if name == "server" || name == "client" || fields[0] != "V" {
		return nil
	}

	return &CertificateInfo{
		State: fields[0],
		Date:  fields[1],
		Hash:  fields[2],
		Name:  name,
	}
}
