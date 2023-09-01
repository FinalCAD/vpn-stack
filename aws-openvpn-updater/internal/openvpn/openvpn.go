package openvpn

import (
	"bufio"
	_ "embed"
	"fmt"
	"os"
	"text/template"
	"time"

	"github.com/FinalCAD/vpn-stack/aws-openvpn-updater/internal/settings"
	"github.com/FinalCAD/vpn-stack/aws-openvpn-updater/internal/utils"
	"github.com/rs/zerolog/log"
)

//go:embed user.tmpl
var templateConfig string

type OpenVpnConfig struct {
	SavedLastModifiedTime   time.Time
	CertificateInfos        []CertificateInfo
	IndexPah                string
	ClientCommonPath        string
	CaPath                  string
	EasyRsaPath             string
	EasyRsaKeyDirectoryPath string
	ClientTlsCryptPath      string
	CrlPath                 string
}

func (o OpenVpnConfig) String() string {
	return fmt.Sprintf("[ IndexPah: %v, ClientCommonPath: %v, CaPath: %v,"+
		"EasyRsaPath: %v, EasyRsaKeyDirectoryPath: %v, ClientTlsCryptPath: %v, CrlPath: %v ]",
		o.IndexPah, o.ClientCommonPath, o.CaPath, o.EasyRsaPath,
		o.EasyRsaKeyDirectoryPath, o.ClientTlsCryptPath, o.CrlPath)
}

func CreateOpenVpnConfig(config *settings.OpenVpn) *OpenVpnConfig {
	easyRsaKeyDirectoryPath := fmt.Sprintf("%s/%s", config.EasyRsaPath, config.EasyRsaKeyDirectory)
	indexPah := fmt.Sprintf("%s/index.txt", easyRsaKeyDirectoryPath)
	caPath := fmt.Sprintf("%s/ca.crt", easyRsaKeyDirectoryPath)
	clientCommonPath := fmt.Sprintf("%s/client-common.txt", config.OpenVpnServerPath)
	clientTlsCryptPath := fmt.Sprintf("%s/tc.key", config.OpenVpnServerPath)
	crlPath := fmt.Sprintf("%s/crl.pem", config.OpenVpnServerPath)
	return &OpenVpnConfig{
		IndexPah:                indexPah,
		ClientCommonPath:        clientCommonPath,
		CaPath:                  caPath,
		EasyRsaPath:             config.EasyRsaPath,
		EasyRsaKeyDirectoryPath: easyRsaKeyDirectoryPath,
		ClientTlsCryptPath:      clientTlsCryptPath,
		CrlPath:                 crlPath,
	}
}

func (o *OpenVpnConfig) GetUser() error {
	var certArray []CertificateInfo

	fileInfo, err := os.Stat(o.IndexPah)
	if err != nil {
		return err
	}

	lastModifiedTime := fileInfo.ModTime()
	if lastModifiedTime == o.SavedLastModifiedTime {
		log.Info().Msg("No change detected in openvpn index file")
		return nil
	}

	o.SavedLastModifiedTime = lastModifiedTime
	file, err := os.Open(o.IndexPah)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		certInfo := CreateCertificateInfo(line)
		if certInfo != nil {
			certArray = append(certArray, *certInfo)
		}
	}
	o.CertificateInfos = certArray
	log.Debug().Msgf("List of valid certificate: %s", o.CertificateInfos)
	return nil
}

func (o *OpenVpnConfig) CreateUser(user string) (string, error) {
	log.Debug().Msgf("Creating config for user: %s", user)
	var err error

	err = cmdNewUser(user, o.EasyRsaPath)
	if err != nil {
		return "", err
	}
	log.Debug().Msgf("Easyrsa command succesfull for user: %s", user)

	outputFileName := fmt.Sprintf("%s/client_configs/%s.ovpn", o.EasyRsaPath, user)

	outputFile, err := utils.CreateFile(outputFileName)
	if err != nil {
		return "", err
	}
	defer outputFile.Close()
	log.Debug().Msgf("Client config file: %s", outputFileName)

	configUser, err := CreateConfigUser(user, o.ClientCommonPath, o.CaPath, o.EasyRsaKeyDirectoryPath, o.ClientTlsCryptPath)
	if err != nil {
		return "", err
	}
	log.Debug().Msgf("Client config infos: %s", configUser)

	tmpl := template.Must(template.New("configTemplate").Parse(templateConfig))
	err = tmpl.Execute(outputFile, configUser)
	if err != nil {
		return "", err
	}
	log.Debug().Msgf("Client config generated succesfully for: %s", user)
	return outputFileName, nil
}

func (o *OpenVpnConfig) DeleteUser(user string) error {
	log.Debug().Msgf("Revoking config for user: %s", user)
	var err error

	err = cmdRevokeUser(user, o.CrlPath, o.EasyRsaPath, o.EasyRsaKeyDirectoryPath)
	if err != nil {
		return err
	}

	configFileName := fmt.Sprintf("%s/client_configs/%s.ovpn", o.EasyRsaPath, user)
	if _, err = os.Stat(configFileName); err == nil {
		err := os.Remove(configFileName)
		if err != nil {
			return err
		}
	}

	log.Debug().Msgf("Client config succesfully revoked for: %s", user)
	return nil
}
