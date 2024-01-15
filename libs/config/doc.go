// Package config provides functions for interacting with configurations
// and a server that listen for requests to change configurations on runtime
package config

import "errors"

var (
	ErrConfigNotFound = errors.New("no configuration match with key")
)
