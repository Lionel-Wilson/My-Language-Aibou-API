package handlers

import (
	"log"
)

type Application struct {
	ErrorLog   *log.Logger
	InfoLog    *log.Logger
	OpenApiKey *string
}
