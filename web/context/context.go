package context

import (
	"log"

	"github.com/dennisbappert/fileharbor/common"
	"github.com/dennisbappert/fileharbor/services"
	"github.com/labstack/echo"
)

type Context struct {
	echo.Context
	services.Services

	Configuration *common.Configuration
}

func New(c *echo.Context, configuration *common.Configuration, services *services.Services) (*Context, error) {
	ctx := &Context{Context: *c, Services: *services, Configuration: configuration}
	log.Println("services instance", services) // TODO: this instance should be created dynamically - see caller!
	return ctx, nil
}
