package main

import (
	"fmt"
	"net/http"

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
}

func (c *DisplayController) handleRefreshDisplay(w http.ResponseWriter, r *http.Request) {
	c.Srv.Display.RefreshUSB()
	w.Write([]byte("Refresh started."))
}

// LogInfo is used to log information messages for this controller.
func (c *DisplayController) LogInfo(v ...interface{}) {
	a := fmt.Sprint(v)
	logger.Info("DisplayController: [Inf] ", a[1:len(a)-1])
}