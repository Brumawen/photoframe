package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/kardianos/service"
	"github.com/onatm/clockwerk"
)

// Server holds the web server
type Server struct {
	PortNo         int                  // Port number the server will listen on
	VerboseLogging bool                 // Verbose logging on/off
	Config         *Config              // Configuration settings
	Display        Display              // Display module
	exit           chan struct{}        // Exit flag
	shutdown       chan struct{}        // Shutdown complete flag
	http           *http.Server         // HTTP server
	router         *mux.Router          // HTTP router
	cw             *clockwerk.Clockwerk // Clockwerk scheduler
}

// Start is called when the service is starting
func (s *Server) Start(v service.Service) error {
	// Create a channel that will be used to block until the Stop signal is received
	s.exit = make(chan struct{})
	go s.run()
	return nil
}

// Stop is called when the service is stopping
func (s *Server) Stop(v service.Service) error {
	s.logInfo("Service stopping")
	// Close the channel, this will automatically release the block
	s.shutdown = make(chan struct{})
	close(s.exit)
	// Wait for the shutdown to complete
	_ = <-s.shutdown
	return nil
}

// run will start up and run the service and wait for a Stop signal
func (s *Server) run() {
	if s.PortNo < 0 {
		s.PortNo = 20511
	}

	s.Display.Srv = s

	// Get the configuration
	if s.Config == nil {
		s.Config = &Config{}
	}
	s.Config.ReadFromFile("config.json")
	s.Config.SetDefaults()

	// Create a router
	s.router = mux.NewRouter().StrictSlash(true)
	s.router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("./html/assets"))))

	// Add the controllers
	s.addController(new(LogController))
	s.addController(new(ConfigController))

	// Create an HTTP server
	s.http = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.PortNo),
		Handler: s.router,
	}

	// Start the web server
	go func() {
		s.logInfo("Server listening on port", s.PortNo)
		if err := s.http.ListenAndServe(); err != nil {
			msg := err.Error()
			if !strings.Contains(msg, "http: Server closed") {
				s.logError("Error starting Web Server.", err.Error())
			}
		}
	}()

	// Wait for an exit signal
	_ = <-s.exit

	// Shutdown the HTTP server
	s.http.Shutdown(nil)

	s.logDebug("Shutdown complete")
	close(s.shutdown)
}

func (s *Server) startSchedule() {
	if s.cw != nil {
		s.cw.Stop()
		s.cw = nil
	}
	s.cw = clockwerk.New()
	s.cw.Every(30 * time.Minute).Do(&s.Display)
}

// AddController adds the specified web service controller to the Router
func (s *Server) addController(c Controller) {
	c.AddController(s.router, s)
}

// logDebug logs a debug message to the logger
func (s *Server) logDebug(v ...interface{}) {
	if s.VerboseLogging {
		a := fmt.Sprint(v)
		logger.Info("Server: ", a[1:len(a)-1])
	}
}

// logInfo logs an information message to the logger
func (s *Server) logInfo(v ...interface{}) {
	a := fmt.Sprint(v)
	logger.Info("Server: ", a[1:len(a)-1])
}

// logError logs an error message to the logger
func (s *Server) logError(v ...interface{}) {
	a := fmt.Sprint(v)
	logger.Error("Server: ", a[1:len(a)-1])
}
