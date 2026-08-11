package main

import (
	"ride-hailing/admin-server/global"
	"go.uber.org/zap"
)

func initializeStandaloneLogger() error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}

	global.GVA_LOG = logger
	zap.ReplaceGlobals(logger)
	return nil
}
