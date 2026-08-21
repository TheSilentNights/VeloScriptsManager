package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.Default()

	RegisterRoutes(r)

	err := r.Run()
	if err != nil {
		println(err.Error())
	}
}
