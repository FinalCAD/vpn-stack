package openvpn

import (
	"fmt"
	"github.com/FinalCAD/vpn-stack/aws-openvpn-updater/internal/utils"
)

type ConfigUser struct {
	ClientCommon   string
	ClientCa       string
	ClientCert     string
	ClientKey      string
	ClientTlsCrypt string
}

func (c ConfigUser) String() string {
	return fmt.Sprintf("[ ClientCommon: %v, ClientCa: %v, ClientCert: %v, ClientKey: %v, ClientTlsCrypt: %v ]",
		c.ClientCommon[0:10], c.ClientCa[0:10], c.ClientCert[0:10], c.ClientKey[0:10], c.ClientTlsCrypt[0:10])
}

func CreateConfigUser(user string, clientCommonPath string, caPath string, easyRsaKeyDirectoryPath string,
	clientTlsCryptPath string) (*ConfigUser, error) {
	var err error
	configUser := &ConfigUser{}
	configUser.ClientCommon, err = utils.ReadContentFromFile(clientCommonPath)
	if err != nil {
		return nil, err
	}

	configUser.ClientCa, err = utils.ReadContentFromFile(caPath)
	if err != nil {
		return nil, err
	}

	clientCert := fmt.Sprintf("%s/issued/%s.crt", easyRsaKeyDirectoryPath, user)
	configUser.ClientCert, err = utils.ReadCertificateFromFile(clientCert)
	if err != nil {
		return nil, err
	}

	clientKey := fmt.Sprintf("%s/private/%s.key", easyRsaKeyDirectoryPath, user)
	configUser.ClientKey, err = utils.ReadContentFromFile(clientKey)
	if err != nil {
		return nil, err
	}

	configUser.ClientTlsCrypt, err = utils.ReadCertificateFromFile(clientTlsCryptPath)
	if err != nil {
		return nil, err
	}
	return configUser, nil
}
