package main

import (
	"fmt"
	"go-base/pkg/restful"
)

type UpdatePasswordData struct {
	Password           string `json:"password" validate:"required,min=8,max=32"`
	NewPassword        string `json:"new_password" validate:"required,min=8,max=32"`
	NewPasswordConfirm string `json:"new_password_confirm" validate:"eqfield=NewPassword"`
}

func TestPostHandler(c *restful.Context) (any, error) {
	name := c.Request.PathParam("name")
	page := c.Request.Query("page")
	fmt.Println(name, page)

	if name == "" {
		c.Logger().Warn("Name came empty")
		name = "World"
	}

	return true, nil
}
