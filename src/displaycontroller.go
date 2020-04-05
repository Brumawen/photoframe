package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// DisplayController handles the Web Methods for reading log records.
type DisplayController struct {
	Srv *Server
}

// AddController adds the controller routes to the router
func (c *DisplayController) AddController(router *mux.Router, s *Server) {
	c.Srv = s
	router.Methods("GET").Path("/display/refresh").Name("RefreshDisplay").
		Handler(Logger(c, http.HandlerFunc(c.handleRefreshDisplay)))
	router.Methods("GET").Path("/display/rebuild").Name("RebuildDisplay").
		Handler(Logger(c, http.HandlerFunc(c.handleRebuildDisplay)))
}

func (c *DisplayController) handleRefreshDisplay(w http.ResponseWriter, r *http.Request) {
	go func() {
		c.Srv.Display.StopUSB()
		time.Sleep(5 * time.Second)
		c.Srv.Display.StartUSB()
	}()
	w.Write([]byte("Refresh started."))
}

func (c *DisplayController) handleRebuildDisplay(w http.ResponseWriter, r *http.Request) {
	c.Srv.Display.Run()
	w.Write([]byte("Rebuild complete."))
}

// LogInfo is used to log information messages for this controller.
func (c *DisplayController) LogInfo(v ...interface{}) {
	a := fmt.Sprint(v...)
	logger.Info("DisplayController: [Inf] ", a)
}
