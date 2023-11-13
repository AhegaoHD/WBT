package bootstrap

import (
	"context"
	"fmt"
	"github.com/AhegaoHD/WBT/config"
	"github.com/AhegaoHD/WBT/internal/controller/http/authController"
	"github.com/AhegaoHD/WBT/internal/controller/http/middleware"
	httpController "github.com/AhegaoHD/WBT/internal/controller/http/userController"
	"github.com/AhegaoHD/WBT/internal/repository/customerRepository"
	"github.com/AhegaoHD/WBT/internal/repository/loaderRepository"
	"github.com/AhegaoHD/WBT/internal/repository/taskRepository"
	"github.com/AhegaoHD/WBT/internal/repository/userRepository"
	"github.com/AhegaoHD/WBT/internal/service/jwtService"
	"github.com/AhegaoHD/WBT/internal/service/taskService"
	"github.com/AhegaoHD/WBT/internal/service/userService"
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

	userRepositoryInstance := userRepository.NewUserRepository(pg)
	customerRepositoryInstance := customerRepository.NewCustomerRepository(pg)
	loaderRepositoryInstance := loaderRepository.NewLoaderRepository(pg)
	taskRepositoryInstance := taskRepository.NewTaskRepository(pg)

	userServiceInstance := userService.NewUserService(pg, userRepositoryInstance, customerRepositoryInstance, loaderRepositoryInstance, taskRepositoryInstance)
	taskServiceInstance := taskService.NewTaskService(pg, taskRepositoryInstance, loaderRepositoryInstance, customerRepositoryInstance)
	jwtServiceInstance := jwtService.NewJWTService(cfg.SecretJWT)

	r := mux.NewRouter()
	middlewareInstance := middleware.NewJWTMiddleware(jwtServiceInstance)

	authControllerInstance := authController.NewAuthController(userServiceInstance, jwtServiceInstance)
	authControllerInstance.RegisterRoutes(r)

	userControllerInstance := httpController.NewUsersController(userServiceInstance, taskServiceInstance, middlewareInstance)
	userControllerInstance.RegisterRoutes(r)

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
