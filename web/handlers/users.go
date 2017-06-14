package handlers

import (
	"log"
	"net/http"

	"github.com/dennisbappert/fileharbor/common"
	"github.com/dennisbappert/fileharbor/web/context"
	"github.com/dennisbappert/fileharbor/web/helper"
	"github.com/labstack/echo"
)

// TODO: password
// TODO: validate mail
func UsersRegister(c echo.Context) error {
	ctx := c.(*context.Context)

	type registration struct {
		GivenName string `json:"givenname"`
		Email     string `json:"email"`
		Password  string `json:"password"`
	}

	reg := new(registration)
	if err := c.Bind(reg); err != nil {
		log.Println("unable to bind request body to type restration")
		return c.NoContent(http.StatusBadRequest)
	}

	id, err := ctx.UserService.Register(reg.Email, reg.GivenName, reg.Password)

	if err != nil {
		if applicationError, ok := err.(*common.ApplicationError); ok {
			response := helper.NewErrorResponse(applicationError.Code, applicationError.Error())
			return c.JSON(http.StatusOK, response)
		}

		log.Println("unexpected error occurred - sending 500", err)
		response := helper.NewUnexpectedErrorResponse()
		return c.JSON(http.StatusInternalServerError, response)
	}

	return c.JSON(http.StatusOK, struct {
		Id string `json:"id"`
	}{id})
}