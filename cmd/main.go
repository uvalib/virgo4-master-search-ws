package main

import (
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/uvalib/virgo4-parser/v4parser"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

// Version of the service
const version = "1.0.0"

/**
 * MAIN
 */
func main() {
	log.Printf("===> V4 Master Search service staring up <===")

	log.Printf("Testing out teh validtaion behavior...")
	simple := "title: {bannanas}"
	validator := v4parser.Validator{}
	valid, errors := validator.Validate(simple)
	if valid == false {
		log.Printf("ERROR: [%s] is not valid, but is should be: %s", simple, errors)
	} else {
		log.Printf("SUCCESS: [%s] is valid", simple)
	}

	// Get config params; service port, directories, DB
	cfg := ServiceConfig{}
	cfg.Load()
	svc := ServiceContext{Version: version}
	err := svc.Init(&cfg)
	if err != nil {
		log.Fatalf("Unable to initialize service: %s", err.Error())
	}

	log.Printf("Setup routes...")
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	router := gin.Default()
	router.Use(cors.Default())
	p := ginprometheus.NewPrometheus("gin")
	p.Use(router)

	router.GET("/", svc.GetVersion)
	router.GET("/favicon.ico", svc.IgnoreFavicon)
	router.GET("/version", svc.GetVersion)
	router.GET("/healthcheck", svc.HealthCheck)
	api := router.Group("/api")
	{
		api.GET("/pools", svc.GetPools)
		api.POST("/pools/register", svc.RegisterPool)
		api.DELETE("/pools/register", svc.DeRegisterPool)
		api.POST("/search", svc.Search)
	}

	portStr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Start service v%s on port %s", version, portStr)
	log.Fatal(router.Run(portStr))
}
