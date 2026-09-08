package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github/TheSilentNights/VeloScriptsManager/service/executor"
	"github/TheSilentNights/VeloScriptsManager/service/utils"
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
	release := flag.Bool("release", false, "run gin in release mode")
	port := flag.Int("port", 19278, "http listen port")
	flag.Parse()

	if *release {
		gin.SetMode(gin.ReleaseMode)
	} else {
		utils.SetDev()
	}

	dataDir := "./temp"
	if *release {
		dataDir = "./data"
	}

	if err := configs.InitConfig(dataDir + "/config.json"); err != nil {
		log.Println(err.Error())
		return
	}

	r := gin.Default()

	db, err := storage.OpenOrCreate(dataDir + "/data.db")
	if err != nil {
		log.Println("failed to open database: " + err.Error())
		return
	}

	scriptRepo := storage.CreateScriptRepo(db)
	environmentRepo := storage.CreateEnvironmentRepo(db)
	executionManager := executor.NewExecutionManager()

	serverController := services.NewServerController(
		executionManager,
	)
	environmentService := services.NewEnvironmentService(environmentRepo)
	scriptService := services.NewScriptService(scriptRepo)
	executionService := services.NewExecutionService(executionManager)

	callerService := services.NewCallerService(executionService)

	router := NewRouter(
		scriptService,
		environmentService,
		serverController,
		executionService,
		callerService,
	)

	router.RegisterRoutes(r)

	server := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", *port),
		Handler: r.Handler(),
	}

	go func() {
		err := server.ListenAndServe()

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	//wait for the shutdown Channel to close
	<-serverController.GetShutdownSignalChan()

	ctx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		panic(err)
	}

	if errDbClose := db.Close(); errDbClose != nil {
		panic(errDbClose)
	}
}
