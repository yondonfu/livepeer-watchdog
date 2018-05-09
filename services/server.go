package services

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
)

type Watcher struct {
	TelNo      string `json:"telNo"`
	Transcoder string `json:"transcoder"`
}

type WebServer struct {
	httpPort string
	db       *DB
}

func NewWebServer(httpPort string, db *DB) *WebServer {
	return &WebServer{
		httpPort: httpPort,
		db:       db,
	}
}

func (ws *WebServer) Start() {
	router := mux.NewRouter()

	router.HandleFunc("/watchers", ws.allWatchers).Methods("GET")
	router.HandleFunc("/watchers", ws.registerWatcher).Methods("POST")
	router.HandleFunc("/watchers/{telNo}", ws.unregisterWatcher).Methods("DELETE")

	go func() {
		glog.Infof("Starting HTTP server at http://localhost:%v", ws.httpPort)

		err := http.ListenAndServe(fmt.Sprintf(":%v", ws.httpPort), router)
		if err != nil {
			glog.Error(err)
		}
	}()
}

func (ws *WebServer) allWatchers(w http.ResponseWriter, r *http.Request) {
	watchers, err := ws.db.AllWatchers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(watchers)
}

func (ws *WebServer) registerWatcher(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var res Watcher
	err := decoder.Decode(&res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = ws.db.RegisterWatcher(res.TelNo, res.Transcoder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	watchers, err := ws.db.AllWatchers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(watchers)
}

func (ws *WebServer) unregisterWatcher(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	err := ws.db.UnregisterWatcher(params["telNo"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	watchers, err := ws.db.AllWatchers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(watchers)
}
