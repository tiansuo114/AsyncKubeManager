package main

import (
	"asyncKubeManager/cmd/console/app"
	"go.uber.org/zap"
)

func main() {
	cmd := app.NewAPIServerCommand()
	if err := cmd.Execute(); err != nil {
		zap.S().Fatal(err)
	}
}
