package handlers

import (
	"github.com/go-kit/kit/log"
)

var Config config

type config struct {
	Logger log.Logger
}