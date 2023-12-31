package app

import (
	"errors"
	"fmt"
	"time"

	"github.com/FinalCAD/vpn-stack/aws-openvpn-updater/internal/awssdk"
	"github.com/FinalCAD/vpn-stack/aws-openvpn-updater/internal/configs"
	"github.com/FinalCAD/vpn-stack/aws-openvpn-updater/internal/openvpn"
	"github.com/FinalCAD/vpn-stack/aws-openvpn-updater/internal/settings"
	"github.com/FinalCAD/vpn-stack/aws-openvpn-updater/internal/utils"
	"github.com/rs/zerolog/log"
)

type App struct {
	Settings      *settings.Settings
	OpenVpnConfig *openvpn.OpenVpnConfig
	AwsSdkConfig  *awssdk.AwsSdkConfig
	IamUsers      []awssdk.User
}

func (a App) String() string {
	return fmt.Sprintf("[ Settings: %v, OpenVpn: %v, AWS: %v ]", a.Settings, a.OpenVpnConfig, a.AwsSdkConfig)
}

func Create(cfg *configs.Config) (*App, error) {
	settings, err := settings.CreateSettings(cfg)
	if err != nil {
		return nil, err
	}
	log.Debug().Msgf("Settings: %s", settings)

	if settings.Params.SendMail && settings.Params.SenderMail == "" {
		errtxt := "error in configuration file, if send-mail is enable you must provide a sender"
		log.Error().Err(err).Msg(errtxt)
		return nil, errors.New(errtxt)
	}

	openvpncfg := openvpn.CreateOpenVpnConfig(settings.OpenVpn)
	awssdkcfg, err := awssdk.CreateIAMConfig(settings.Aws)
	if err != nil {
		return nil, err
	}
	app := &App{Settings: settings, OpenVpnConfig: openvpncfg, AwsSdkConfig: awssdkcfg}
	return app, nil
}

func (app *App) Start() {
	var err error
	exitChan := utils.GetFireSignalsChannel()
	go func() {
		for {
			log.Debug().Msg("-- Start update user loop --")
			err = app.lookupUsers()
			if err == nil {
				app.createUsers()
				app.deleteUsers()
			}
			log.Debug().Msg("-- End update user loop --")
			time.Sleep(time.Second * time.Duration(app.Settings.Params.RequestInterval))
		}
	}()
	<-exitChan
	log.Info().Msg("Program ended from signal")
}

func (app *App) lookupUsers() error {
	err := app.OpenVpnConfig.GetUser()
	if err != nil {
		log.Error().Err(err).Msg("Error getting openvpn users")
		return err
	}

	app.IamUsers, err = app.AwsSdkConfig.GetIAMUser()
	if err != nil {
		log.Error().Err(err).Msg("Error getting iam users")
		return err
	}
	return nil
}

func (app *App) createUsers() {
	var found bool
	for _, user := range app.IamUsers {
		found = false
		for _, account := range app.OpenVpnConfig.CertificateInfos {
			if user.Name == account.Name {
				found = true
				break
			}
		}
		if !found {
			app.createUser(user)
		}
	}
}

func (app *App) createUser(user awssdk.User) {
	log.Info().Msgf("Adding new user: %s", user.Name)
	var presignUrl string
	var filePath string
	var err error
	if !app.Settings.Params.Dryrun {
		filePath, err = app.OpenVpnConfig.CreateUser(user.Name, app.Settings.Params.UseFqdn)
		if err != nil {
			log.Error().Err(err).Msgf("Error creating openvpn client config: %s", user.Name)
			return
		}
	} else {
		log.Info().Msgf("Dry run creating config for user: %s", user.Name)
	}
	if app.Settings.Params.S3Upload && !app.Settings.Params.Dryrun {
		presignUrl, err = app.AwsSdkConfig.SaveConfS3(app.Settings.Config.Environment, user.Name, filePath)
		if err != nil {
			log.Error().Err(err).Msgf("Error s3 upload: %s", user.Name)
			return
		}
	}
	if app.Settings.Params.SendMail && !app.Settings.Params.Dryrun {
		err = app.AwsSdkConfig.SendMail(app.Settings.Config.Environment, user, presignUrl, app.Settings.Params.SenderMail)
		if err != nil {
			log.Error().Err(err).Msgf("Error sending email: %s", user.Name)
			return
		}
	}
	log.Info().Msgf("Added new user successfully: %s", user.Name)
}

func (app *App) deleteUsers() {
	var found bool
	for _, account := range app.OpenVpnConfig.CertificateInfos {
		found = false
		for _, user := range app.IamUsers {
			if user.Name == account.Name {
				found = true
				break
			}
		}
		if !found {
			app.deleteUser(account.Name)
		}
	}
}

func (app *App) deleteUser(user string) {
	log.Info().Msgf("Deleting existing user: %s", user)
	var err error
	if !app.Settings.Params.Dryrun {
		err = app.OpenVpnConfig.DeleteUser(user)
		if err != nil {
			log.Error().Err(err).Msgf("Error revoking openvpn client config: %s", user)
			return
		}
	} else {
		log.Info().Msgf("Dry run deleting config for user: %s", user)
	}
	if app.Settings.Params.S3Upload && !app.Settings.Params.Dryrun {
		err = app.AwsSdkConfig.RemoveConfS3(app.Settings.Config.Environment, user)
		log.Error().Err(err).Msgf("Error removing S3 file client config: %s", user)
		return
	}
	log.Info().Msgf("Deleted user successfully: %s", user)
}
