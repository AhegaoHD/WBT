package bootstrap

import (
	"context"
	"fmt"
	"github.com/AhegaoHD/WBT/config"
	httpController "github.com/AhegaoHD/WBT/internal/controller/http"
	"github.com/AhegaoHD/WBT/internal/repository"
	"github.com/AhegaoHD/WBT/internal/service"
	"github.com/AhegaoHD/WBT/pkg/httpserver"
	"github.com/AhegaoHD/WBT/pkg/postgres"
	"github.com/gorilla/mux"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func Run(cfg *config.Config) {
	pg, err := postgres.New(postgres.GetConnString(&cfg.Db), postgres.MaxPoolSize(cfg.Db.MaxPoolSize))
	if err != nil {
		log.Fatal("APP - START - POSTGRES INI PROBLEM: %v", err)
	}
	defer pg.Close()

	err = pg.Pool.Ping(context.Background())
	if err != nil {
		log.Fatal("APP - START - POSTGRES INI PROBLEM: %v", err)
		return
	}

	userRepository := repository.NewUserRepository(pg)
	customerRepository := repository.NewCustomerRepository(pg)
	loaderRepository := repository.NewLoaderRepository(pg)
	taskRepository := repository.NewTaskRepository(pg)

	userService := service.NewUserService(pg, userRepository, customerRepository, loaderRepository, taskRepository)
	taskService := service.NewTaskService(pg, taskRepository, loaderRepository, userRepository, customerRepository)

	r := mux.NewRouter()

	authController := httpController.NewAuthController(userService)
	authController.RegisterRoutes(r)

	userController := httpController.NewUsersController(userService, taskService)
	userController.RegisterRoutes(r)

	httpServer := httpserver.New(r,
		httpserver.Port(cfg.HttpServer.Addr),
		httpserver.ReadTimeout(cfg.HttpServer.ReadTimeout),
		httpserver.WriteTimeout(cfg.HttpServer.WriteTimeout),
		httpserver.ShutdownTimeout(cfg.HttpServer.ShutdownTimeout),
	)
	log.Println("Starting HTTP server on port 8080")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	log.Println("RUNNING APP:%v VERSION:%v", cfg.App.Name, cfg.App.Version)

	select {
	case s := <-interrupt:
		log.Println("APP - RUN - signal: " + s.String())
	case err = <-httpServer.Notify():
		log.Fatal(fmt.Errorf("APP - RUN - HTTPSERVER.NOTIFY: %v", err))
	}

	err = httpServer.Shutdown()
	if err != nil {
		log.Fatal(fmt.Errorf("APP - RUN - HTPPSERVER.SHUTDOWN: %v", err))
	}
}
