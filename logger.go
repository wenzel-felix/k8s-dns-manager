package main

import "go.uber.org/zap"

var logger *zap.SugaredLogger

// initLogger initializes the logger
func initLogger() {
	var err error
	zaplogger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	logger = zaplogger.Sugar()
}
