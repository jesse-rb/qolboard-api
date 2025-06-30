package controller

type GetParams struct {
	With []string `form:"with[]"`
}

type IndexParams struct {
	Page  int      `form:"page" binding:"gte=1"`
	Limit int      `form:"limit" binding:"gte=1,lte=100"`
	With  []string `form:"with[]"`
}
