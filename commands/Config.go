package commands

import (
	"freelance-discord/models"
	"log"
	"os"
)

type Application struct {
	Db *models.DatabaseModel
}

var logger = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
