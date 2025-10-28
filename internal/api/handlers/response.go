package handlers

import (
	"net/http"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"result"`
}

func NewErrorResponse(c *ginext.Context, statusCode int, message string) {
	zlog.Logger.Error().Msg(message)
	c.AbortWithStatusJSON(statusCode, ErrorResponse{Error: message})
}

func NewSuccessResponse(c *ginext.Context, statusCode int, message string) {
	c.JSON(statusCode, SuccessResponse{Message: message})
}

func GetHome(c *ginext.Context) {
	c.HTML(http.StatusOK, "user.html", ginext.H{"result": "ok"})
}

func GetAdmin(c *ginext.Context) {
	c.HTML(http.StatusOK, "admin.html", ginext.H{"result": "ok"})

}

func GetAdminLogin(c *ginext.Context) {
	c.HTML(http.StatusOK, "admin_login.html", ginext.H{"result": "ok"})
}
