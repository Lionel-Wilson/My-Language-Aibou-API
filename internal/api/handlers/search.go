package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *Application) DefineWord(c *gin.Context) {

	c.JSON(http.StatusOK, "Here's your definition")
}

func (app *Application) DefinePhrase(c *gin.Context) {

	c.JSON(http.StatusOK, "Here's a breakdown of the phrase")
}
