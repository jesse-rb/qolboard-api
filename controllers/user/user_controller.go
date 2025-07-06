package user_controller

import (
	database_config "qolboard-api/config/database"
	controller "qolboard-api/controllers"
	model "qolboard-api/models"
	user_model "qolboard-api/models/user"
	error_service "qolboard-api/services/error"
	generator_service "qolboard-api/services/generator"
	relations_service "qolboard-api/services/relations"
	response_service "qolboard-api/services/response"

	"github.com/gin-gonic/gin"
)

type getParams struct {
	controller.GetParams
}

func Get(c *gin.Context) {
	var params getParams = getParams{
		GetParams: controller.GetParams{
			With: make([]string, 0),
		},
	}

	if err := c.ShouldBindQuery(&params); err != nil {
		error_service.ValidationError(c, err)
		return
	}

	tx, err := database_config.DB(c)
	defer tx.Commit()
	if err != nil {
		error_service.InternalError(c, err.Error())
		tx.Rollback()
		return
	}

	user, err := user_model.Get(tx)
	if err != nil {
		error_service.InternalError(c, err.Error())
		tx.Rollback()
		return
	}

	err = relations_service.Load(tx, model.UserRelations, user, params.With)
	if err != nil {
		tx.Rollback()
		error_service.InternalError(c, err.Error())
		return
	}

	resp := generator_service.BuildResponse(*user)

	response_service.SetJSON(c, gin.H{
		"data": resp,
	})
}
