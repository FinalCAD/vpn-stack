package main

import (
	"github.com/FinalCAD/vpn-stack/aws-openvpn-updater/internal/app"
	"github.com/FinalCAD/vpn-stack/aws-openvpn-updater/internal/configs"
	"github.com/rs/zerolog/log"
	"os"
)

type Exit struct{ Code int }

func exitHandler() {
	if e := recover(); e != nil {
		if exit, ok := e.(Exit); ok {
			os.Exit(exit.Code)
		}
		panic(e)
	}
}

func main() {
	defer exitHandler()
	config := configs.InitApp()
	if config.Environment == "" {
		log.Fatal().Msg("Missing environment command line parameter")
	}
	app, err := app.Create(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Error during setup")
	}
	app.Start()
}
