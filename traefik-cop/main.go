package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/spf13/viper"
	"github.com/gorilla/mux"
)

type UAConfig struct {
	Path string `json:"path"`
	UserAgent string `json:"useragent"`
	URL string `json:"url"`
}

type NewRoute struct {
	Name string `json:"name"`
	Address string `json:"address"`
	Host string `json:"host"`
	UserAgent string `json:"useragent"`
	RedirectURL string `json:"redirect_url"`
}

type DeleteRoute struct {
	Name string `json:"name"`
}

type ResponseMsg struct {
	Response string
}

func deleteRouteHandler(w http.ResponseWriter, r *http.Request) {
	var response ResponseMsg
	var deleteRoute DeleteRoute
	err := json.NewDecoder(r.Body).Decode(&deleteRoute)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	router := fmt.Sprintf("http.routers.%s", deleteRoute.Name)
	rtrName := fmt.Sprintf("%s", deleteRoute.Name)
	mwName := fmt.Sprintf("%s-redirect", deleteRoute.Name)
	svcName := fmt.Sprintf("%s-svc", deleteRoute.Name)

	traefikCop := loadConfig()

	if !traefikCop.IsSet(router) {
		http.Error(w, fmt.Sprintf("%s route does not exist", deleteRoute.Name), http.StatusInternalServerError)
		return
	}

	log.Println("Deleting route, service and middleware for: %s", deleteRoute.Name)

	middlewares := traefikCop.GetStringMap("http.middlewares")
	delete(middlewares, mwName)
	routers := traefikCop.GetStringMap("http.routers")
	delete(routers, rtrName)
	services := traefikCop.GetStringMap("http.services")
	delete(services, svcName)

	traefikCop.Set("http.middlewares", middlewares)
	traefikCop.Set("http.routers", routers)
	traefikCop.Set("http.services", services)

	err = traefikCop.WriteConfig()
	if err != nil {
		log.Println("Failed to write config after deleting route")
	}

	response.Response = fmt.Sprintf("%s route data deleted", deleteRoute.Name)
	json.NewEncoder(w).Encode(response)
}

func newRouteHandler(w http.ResponseWriter, r *http.Request) {
	var response ResponseMsg
	var newRoute NewRoute
	err := json.NewDecoder(r.Body).Decode(&newRoute)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rtrName := fmt.Sprintf("%s", newRoute.Name)
	mwName := fmt.Sprintf("%s-redirect", newRoute.Name)
	svcName := fmt.Sprintf("%s-svc", newRoute.Name)
	hostRule := fmt.Sprintf("Host(`%s`)", newRoute.Host)

	traefikCop := loadConfig()

	if traefikCop.IsSet(fmt.Sprintf("http.routers.%s", rtrName)) {
		http.Error(w, fmt.Sprintf("%s name already exists", rtrName), http.StatusInternalServerError)
		return
	}

	newService := fmt.Sprintf("http.services.%s.loadbalancer.servers", svcName)
	servers := []map[string]string {
		map[string]string {"url": newRoute.Address},
	}
	traefikCop.Set(newService, servers)

	newRtrEntryPoint := fmt.Sprintf("http.routers.%s.entrypoints", rtrName)
	traefikCop.Set(newRtrEntryPoint, "web")

	newRtrMW := fmt.Sprintf("http.routers.%s.middlewares", rtrName)
	traefikCop.Set(newRtrMW, mwName)

	newRtrRule := fmt.Sprintf("http.routers.%s.rule", rtrName)
	traefikCop.Set(newRtrRule, hostRule)

	newRtrSvc := fmt.Sprintf("http.routers.%s.service", rtrName)
	traefikCop.Set(newRtrSvc, svcName)

	newMWURL := fmt.Sprintf("http.middlewares.%s.plugin.uaredirect.url", mwName)
	traefikCop.Set(newMWURL, newRoute.RedirectURL)

	newMWUA := fmt.Sprintf("http.middlewares.%s.plugin.uaredirect.useragent", mwName)
	traefikCop.Set(newMWUA, newRoute.UserAgent)

	response.Response = fmt.Sprintf("Added new route: %s", newRoute.Name)
	json.NewEncoder(w).Encode(response)

	log.Println("Saving new route: ", newRoute.Name)
	err = traefikCop.WriteConfig()
	if err != nil {
		log.Println("Failed to write config updates")
	} else {
		log.Println("Updated user-agent and wrote changes to disk!")
	}
}

func setUserAgentHandler(w http.ResponseWriter, r *http.Request) {
	var response ResponseMsg

	if r.Method == http.MethodPost {
		log.Println("Received a POST for setting the user agent")

		var uaConfig UAConfig
		err := json.NewDecoder(r.Body).Decode(&uaConfig)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		traefikCop := loadConfig()

		if !traefikCop.IsSet(uaConfig.Path) {
			http.Error(w, "Invalid Path to UA", http.StatusInternalServerError)
			return
		}

		traefikCop.Set(uaConfig.Path, uaConfig.UserAgent)
		response.Response = fmt.Sprintf("changed UserAgent from to %s", uaConfig.UserAgent)
		json.NewEncoder(w).Encode(response)

		log.Println("Saving changes to disk")
		err = traefikCop.WriteConfig()
		if err != nil {
			log.Println("Failed to write config updates")
		} else {
			log.Println("Updated user-agent and wrote changes to disk!")
		}

		return
	}
	http.Error(w, "Method Not Implemented", http.StatusInternalServerError)
}

func dumpAll(w http.ResponseWriter, r *http.Request) {
	traefikCop := loadConfig()
	json.NewEncoder(w).Encode(traefikCop.Get("http"))
}

func loadConfig() *viper.Viper {
	traefikCop := viper.New()
	traefikCop.SetConfigName("config")
	traefikCop.SetConfigType("yaml") 
	traefikCop.AddConfigPath("config/")
	log.Println("Loading in configuration file")
	err := traefikCop.ReadInConfig()
	if err != nil {
		log.Fatalln("Failed to read config")
	}
	return traefikCop
}

func main() {

	rtr := mux.NewRouter()
	rtr.HandleFunc("/deleteroute", deleteRouteHandler)
	rtr.HandleFunc("/addroute", newRouteHandler)
	rtr.HandleFunc("/useragent", setUserAgentHandler)
	rtr.HandleFunc("/settings", dumpAll)

	srv := &http.Server{
		Handler: rtr,
		Addr: "0.0.0.0:7000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
