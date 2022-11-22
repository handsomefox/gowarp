package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/handsomefox/gowarp/pkg/models/mongo"
	"github.com/handsomefox/gowarp/pkg/server/storage"
	"github.com/handsomefox/gowarp/pkg/warp"
	"github.com/handsomefox/gowarp/pkg/warp/pastebin"
)

// Server is the main gowarp http.Handler.
type Server struct {
	handler   http.Handler
	templates *storage.TemplateStorage
	service   *warp.Service
	storage   *storage.Storage
}

// NewServer returns a *Server with all the required setup done.
func NewServer(useProxy bool, connStr string) (*Server, error) {
	// Connect to database
	db, err := mongo.NewAccountModel(context.TODO(), connStr)
	if err != nil {
		panic(err) // nowhere to store keys :/
	}

	log.Println("Connected to the database")

	// Create storage for templates
	ts, err := storage.NewTemplateStorage()
	if err != nil {
		return nil, fmt.Errorf("error creating the server: %w", err)
	}

	// Create server
	service := warp.NewService(nil, useProxy)
	config, err := pastebin.GetConfig(context.Background())
	if err != nil {
		panic(err) // we probably have outdated keys anyway, no point in continuing.
	}
	service.UpdateConfig(config)

	store := &storage.Storage{AM: db}

	server := &Server{
		templates: ts,
		service:   service,
		storage:   store,
	}

	server.SetupRoutes()

	// Start a goroutine to automatically update the config.
	go func() {
		for {
			time.Sleep(1 * time.Hour) // update config every hour.
			config, err := pastebin.GetConfig(context.Background())
			if err != nil {
				continue
			}
			server.service.UpdateConfig(config)
		}
	}()

	// Start a goroutine to generate keys in the background.
	go server.storage.Fill(server.service)

	return server, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}
