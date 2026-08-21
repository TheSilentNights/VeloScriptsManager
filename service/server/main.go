package main

import (
	"github/TheSilentNights/VeloScriptsManager/service/configs"

	"github.com/gin-gonic/gin"
)

func main() {
	err := configs.InitConfig("../temp/config.json")

	if err != nil {
		println(err.Error())
		return
	}

	r := gin.Default()

	RegisterRoutes(r)

	err = r.Run()
	if err != nil {
		println(err.Error())
	}
}
