package settings

import (
	"fmt"
	"os"

	"github.com/FinalCAD/vpn-stack/aws-openvpn-updater/internal/configs"
	"github.com/pelletier/go-toml/v2"
)

const (
	defaultRequestInterval     int    = 300
	defaultEasyRsaPath         string = "/etc/openvpn/server/easy-rsa"
	defaultEasyRsaKeyDirectory string = "pki"
	defaultOpenVpnServerPath   string = "/etc/openvpn/server"
	defaultRegion              string = "eu-central-1"
	defaultSenderMail          string = ""
)

type Settings struct {
	Aws     *Aws `toml:"aws"`
	Config  *configs.Config
	OpenVpn *OpenVpn `toml:"openvpn"`
	Params  *Params  `toml:"settings"`
}

func (s Settings) String() string {
	return fmt.Sprintf("[ Aws: %v, Config: %v, OpenVpn: %v, Params: %v ]", s.Aws, s.Config, s.OpenVpn, s.Params)
}

type Params struct {
	Dryrun          bool   `toml:"dry-run"`
	RequestInterval int    `toml:"request-interval"`
	S3Upload        bool   `toml:"s3-upload"`
	SenderMail      string `toml:"sender"`
	SendMail        bool   `toml:"send-mail"`
}

func (p Params) String() string {
	return fmt.Sprintf("[ RequestInterval: %v, S3Upload: %v, SendMail: %v, SenderMail: %v, Dryrun: %v ]",
		p.RequestInterval, p.S3Upload, p.SendMail, p.SenderMail, p.Dryrun)
}

type OpenVpn struct {
	EasyRsaPath         string `toml:"easy-rsa-path"`
	EasyRsaKeyDirectory string `toml:"key-directory"`
	OpenVpnServerPath   string `toml:"server-path"`
}

func (o OpenVpn) String() string {
	return fmt.Sprintf("[ EasyRsaPath: %v, EasyRsaKeyDirectory: %v, OpenVpnServerPath: %v ]", o.EasyRsaPath, o.EasyRsaKeyDirectory, o.OpenVpnServerPath)
}

type Aws struct {
	Profile      string `toml:"profile"`
	BucketName   string `toml:"s3-bucket-name"`
	Region       string `toml:"region"`
	RoleToAssume string `toml:"assume-role"`
	VpnGroup     string `toml:"vpn-group"`
}

func (a Aws) String() string {
	return fmt.Sprintf("[ Profile: %v, Region: %v, BucketName: %v, VpnGroup: %v , RoleToAssume: %v ]", a.Profile, a.Region, a.BucketName, a.VpnGroup, a.RoleToAssume)
}

func CreateSettings(config *configs.Config) (*Settings, error) {
	params := &Params{RequestInterval: defaultRequestInterval, S3Upload: true, SendMail: true,
		SenderMail: defaultSenderMail, Dryrun: false}
	openvpn := &OpenVpn{EasyRsaPath: defaultEasyRsaPath,
		EasyRsaKeyDirectory: defaultEasyRsaKeyDirectory,
		OpenVpnServerPath:   defaultOpenVpnServerPath}
	aws := &Aws{Profile: "", Region: defaultRegion, RoleToAssume: ""}
	settings := &Settings{Config: config, Params: params, OpenVpn: openvpn, Aws: aws}
	cfg, err := os.ReadFile(config.ConfigFile)
	if err != nil {
		return nil, err
	}
	err = toml.Unmarshal(cfg, &settings)
	if err != nil {
		return nil, err
	}
	return settings, nil
}
