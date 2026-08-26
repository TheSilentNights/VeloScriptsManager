package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github/TheSilentNights/VeloScriptsManager/service/configs"
	"github/TheSilentNights/VeloScriptsManager/service/services"
	"github/TheSilentNights/VeloScriptsManager/service/storage"

	"github.com/gin-gonic/gin"
)

// serverShutdownTimeout bounds the graceful HTTP shutdown so lingering
// connections cannot block process exit forever.
const serverShutdownTimeout = 10 * time.Second

func main() {
	if err := configs.InitConfig("./temp/config.json"); err != nil {
		log.Println(err.Error())
		return
	}

	r := gin.Default()

	db, err := storage.OpenOrCreate("./temp/test_repo.db")
	if err != nil {
		log.Println("failed to open database: " + err.Error())
		return
	}

	scriptRepo := storage.CreateScriptRepo(db)
	environmentRepo := storage.CreateEnvironmentRepo(db)

	service := services.NewService(scriptRepo, environmentRepo)

	router := NewRouter(service)

	router.RegisterRoutes(r)

	server := &http.Server{
		Addr:    ":8080",
		Handler: r.Handler(),
	}

	go func() {
		err := server.ListenAndServe()

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Println(err.Error())
		}
	}()

	//wait for the shutdown Channel to close
	<-service.GetShutdownSignalChan()

	// Close attached execution WebSockets first; Shutdown waits for active
	// connections and would otherwise hang on long-lived streams.
	router.CloseWebSockets()

	ctx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Println(err.Error())
	}

	if errDbClose := db.Close(); errDbClose != nil {
		log.Println("close db: " + errDbClose.Error())
	}
}
