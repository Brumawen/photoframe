package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// ConfigController handles the Web Methods for configuring the module.
type ConfigController struct {
	Srv *Server
}

// ConfigPageData holds the data used to write to the configuration page.
type ConfigPageData struct {
	Resolution     int
	Provider       int
	ImgCount       int
	EnableWeather  string
	EnableCalendar string
}

// AddController adds the controller routes to the router
func (c *ConfigController) AddController(router *mux.Router, s *Server) {
	c.Srv = s
	router.Path("/config.html").Handler(Logger(c, http.HandlerFunc(c.handleConfigWebPage)))
	router.Methods("GET").Path("/config/get").Name("GetConfig").
		Handler(Logger(c, http.HandlerFunc(c.handleGetConfig)))
	router.Methods("POST").Path("/config/set").Name("SetConfig").
		Handler(Logger(c, http.HandlerFunc(c.handleSetConfig)))
}

func (c *ConfigController) handleConfigWebPage(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("./html/config.html"))

	v := ConfigPageData{
		Resolution: c.Srv.Config.Resolution,
		Provider:   c.Srv.Config.Provider,
		ImgCount:   c.Srv.Config.ImgCount,
	}
	if c.Srv.Config.Weather {
		v.EnableWeather = "checked"
	}
	if c.Srv.Config.Calendar {
		v.EnableCalendar = "checked"
	}

	err := t.Execute(w, v)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (c *ConfigController) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	if err := c.Srv.Config.WriteTo(w); err != nil {
		http.Error(w, "Error serializing configuration. "+err.Error(), 500)
	}
}

func (c *ConfigController) handleSetConfig(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	res := r.Form.Get("resolution")
	pro := r.Form.Get("provider")
	img := r.Form.Get("imgcount")

	weather := r.Form.Get("weather")
	calendar := r.Form.Get("calendar")

	if res == "" {
		http.Error(w, "The Resolution must be specified", 500)
		return
	}
	resv, err := strconv.Atoi(res)
	if err != nil || resv != 0 {
		http.Error(w, "Invalid Resolution value", 500)
		return
	}
	if pro == "" {
		http.Error(w, "The Image Provider must be selected", 500)
		return
	}
	prov, err := strconv.Atoi(pro)
	if err != nil || prov < 0 || prov > 3 {
		http.Error(w, "Invalid Image Provider value", 500)
		return
	}
	if img == "" {
		http.Error(w, "The Image Count must be provided", 500)
		return
	}
	imgv, err := strconv.Atoi(img)
	if err != nil || imgv <= 0 {
		http.Error(w, "Image Count must be greater than zero", 500)
		return
	}

	c.LogInfo("Setting new configuration values.")

	c.Srv.Config.Resolution = resv
	c.Srv.Config.Provider = prov
	c.Srv.Config.ImgCount = imgv
	c.Srv.Config.Weather = (weather == "on")
	c.Srv.Config.Calendar = (calendar == "on")
	c.Srv.Config.SetDefaults()

	c.Srv.Config.WriteToFile("config.json")
}

// LogInfo is used to log information messages for this controller.
func (c *ConfigController) LogInfo(v ...interface{}) {
	a := fmt.Sprint(v...)
	logger.Info("ConfigController: [Inf] ", a)
}

// LogError is used to log error messages for this controller.
func (c *ConfigController) LogError(v ...interface{}) {
	a := fmt.Sprint(v...)
	logger.Error("ConfigController: [Err] ", a)
}
