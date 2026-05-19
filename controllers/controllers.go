package controllers

import "qolboard-api/services/email"

type GetParams struct {
	With []string `form:"with[]"`
}

type IndexParams struct {
	Page  int      `form:"page" binding:"gte=1"`
	Limit int      `form:"limit" binding:"gte=1,lte=100"`
	With  []string `form:"with[]"`
}

type RESTHandler struct {
	emailClient email.EmailClient
}

func NewRESTHAndler(emailClient email.EmailClient) *RESTHandler {
	return &RESTHandler{
		emailClient: emailClient,
	}
}
