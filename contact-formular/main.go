package main

import (
	"fmt"
	"log"
	"net/http"
	"github.com/gorilla/mux"
)

type Server struct {
	Router *mux.Router
}

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
  const port = "8080"
  host := fmt.Sprintf("0.0.0.0:%s", port)
  srv := NewServer()
	log.Printf("Starting tool on %s \n", host)
	err := http.ListenAndServe(host, srv.Router)
	if err != nil {
		log.Fatal("cannot listen and server", err)
	}
}

func NewServer() *Server {
	router := mux.NewRouter()
  router.HandleFunc("/booking/request", HandlerBookingRequest).Methods("POST")
	return &Server{Router: router}
}

func HandlerBookingRequest(w http.ResponseWriter, r *http.Request){

}
