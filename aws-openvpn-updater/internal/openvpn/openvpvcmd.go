package openvpn

import (
	"fmt"
	"os"
	"os/exec"
)

const (
	newUserCmd    string = "cd %s && ./easyrsa --batch --days=3650 build-client-full \"%s\" nopass > /dev/null 2>&1"
	revokeUserCmd string = "cd %s && ./easyrsa --batch revoke \"%s\" > /dev/null 2>&1 && ./easyrsa --batch --days=3650 gen-crl > /dev/null 2>&1"
	updateCrtCmd  string = "rm -f %[1]s > /dev/null 2>&1 && cp %s/crl.pem %[1]s> /dev/null 2>&1 && chown nobody:nogroup %[1]s > /dev/null 2>&1"
)

func cmdNewUser(user string, path string) error {
	cmd := exec.Command("bash", "-c", fmt.Sprintf(newUserCmd, path, user))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func cmdRevokeUser(user string, crlPath string, easyRsaPath string, easyRsaKeyDirectoryPath string) error {
	cmd := exec.Command("bash", "-c", fmt.Sprintf(revokeUserCmd, easyRsaPath, user))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("bash", "-c", fmt.Sprintf(updateCrtCmd, crlPath, easyRsaKeyDirectoryPath))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
