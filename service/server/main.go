package main

import (
	"context"
	"errors"
	"net/http"

	"github/TheSilentNights/VeloScriptsManager/service/configs"
	"github/TheSilentNights/VeloScriptsManager/service/services"
	"github/TheSilentNights/VeloScriptsManager/service/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	err := configs.InitConfig("./temp/config.json")

	if err != nil {
		println(err.Error())
		return
	}

	r := gin.Default()

	db, err := storage.OpenOrCreate("./temp/test_repo.db")
	if err != nil {
		println("failed to open database: " + err.Error())
		return
	}

	scriptRepo := storage.CreateScriptRepo(db)

	service := services.NewService(scriptRepo)

	router := NewRouter(service)

	router.RegisterRoutes(r)

	server := &http.Server{
		Addr:    ":8080",
		Handler: r.Handler(),
	}

	go func() {
		err := server.ListenAndServe()

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			println(err.Error())
		}
	}()

	//wait for the shutdown Channel to close
	<-service.GetShutdownSignalChan()

	err = server.Shutdown(context.Background())
	if err != nil {
		println(err.Error())
	}

	errDbClose := db.Close()
	if errDbClose != nil {
		//todo: fix the db error
		return
	}
}
