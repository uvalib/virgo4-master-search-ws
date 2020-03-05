package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	dbx "github.com/go-ozzo/ozzo-dbx"
	_ "github.com/lib/pq"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// ServiceContext contains common data used by all handlers
type ServiceContext struct {
	Version      string
	DB           *dbx.DB
	SuggestorURL string
	I18NBundle   *i18n.Bundle
}

// InitializeService will initialize the service context based on the config parameters.
// Any pools found in the DB will be added to the context and polled for status.
// Any errors are FATAL.
func InitializeService(version string, cfg *ServiceConfig) *ServiceContext {
	log.Printf("Initializing Service")
	svc := ServiceContext{Version: version, SuggestorURL: cfg.SuggestorURL}

	log.Printf("Connect to Postgres")
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=disable",
		cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBHost, cfg.DBPort)
	db, err := dbx.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	db.LogFunc = log.Printf
	svc.DB = db

	log.Printf("Init localization")
	svc.I18NBundle = i18n.NewBundle(language.English)
	svc.I18NBundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	svc.I18NBundle.MustLoadMessageFile("./i18n/active.en.toml")
	svc.I18NBundle.MustLoadMessageFile("./i18n/active.es.toml")

	return &svc
}

// IgnoreFavicon is a dummy to handle browser favicon requests without warnings
func (svc *ServiceContext) IgnoreFavicon(c *gin.Context) {
}

// GetVersion reports the version of the serivce
func (svc *ServiceContext) GetVersion(c *gin.Context) {
	build := "unknown"
	// cos our CWD is the bin directory
	files, _ := filepath.Glob("../buildtag.*")
	if len(files) == 1 {
		build = strings.Replace(files[0], "../buildtag.", "", 1)
	}

	vMap := make(map[string]string)
	vMap["version"] = svc.Version
	vMap["build"] = build
	c.JSON(http.StatusOK, vMap)
}

// HealthCheck reports the health of the serivce
func (svc *ServiceContext) HealthCheck(c *gin.Context) {
	type hcResp struct {
		Healthy bool   `json:"healthy"`
		Message string `json:"message,omitempty"`
	}
	hcMap := make(map[string]hcResp)

	tq := svc.DB.NewQuery("select count(*) as total from sources")
	var total int
	err := tq.Row(&total)
	if err != nil {
		hcMap["postgres"] = hcResp{Healthy: false, Message: err.Error()}
	} else {
		hcMap["postgres"] = hcResp{Healthy: true}
	}

	if svc.SuggestorURL != "" {
		timeout := time.Duration(5 * time.Second)
		client := http.Client{
			Timeout: timeout,
		}
		apiURL := fmt.Sprintf("%s/version", svc.SuggestorURL)
		_, err := client.Get(apiURL)
		if err != nil {
			log.Printf("ERROR: Suggestor %s ping failed: %s", svc.SuggestorURL, err.Error())
			hcMap["suggestor"] = hcResp{Healthy: false, Message: err.Error()}
		} else {
			hcMap["suggestor"] = hcResp{Healthy: true}
		}
	}

	c.JSON(http.StatusOK, hcMap)
}

// getBearerToken is a helper to extract the user auth token from the Auth header
func getBearerToken(authorization string) (string, error) {
	components := strings.Split(strings.Join(strings.Fields(authorization), " "), " ")

	// must have two components, the first of which is "Bearer", and the second a non-empty token
	if len(components) != 2 || components[0] != "Bearer" || components[1] == "" {
		return "", fmt.Errorf("Invalid Authorization header: [%s]", authorization)
	}

	return components[1], nil
}

// AuthMiddleware is a middleware handler that verifies presence of a
// user Bearer token in the Authorization header.
func (svc *ServiceContext) AuthMiddleware(c *gin.Context) {
	token, err := getBearerToken(c.Request.Header.Get("Authorization"))

	if err != nil {
		log.Printf("Authentication failed: [%s]", err.Error())
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// do something with token

	log.Printf("got bearer token: [%s]", token)
}
