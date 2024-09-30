package api

import (
    "log"
    "net/http"
	"github.com/ANUB1s-xio/TidalFlow/framework/internal"
    "github.com/gorilla/mux"
)

func Run() {
    r := mux.NewRouter()

    // API endpoints to fetch application data
    r.HandleFunc("/applications", internal.GetApplications).Methods("GET")
    r.HandleFunc("/applications/{name}", internal.GetApplicationByName).Methods("GET")

    log.Fatal(http.ListenAndServe(":8080", r))
}
