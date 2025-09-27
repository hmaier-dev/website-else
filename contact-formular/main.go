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
  const port = "3000"
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
  router.HandleFunc("/contact", HandlerContactRequest).Methods("POST")
  router.HandleFunc("/mailbox", HandlerMailboxRequest).Methods("GET")
	return &Server{Router: router}
}

func HandlerContactRequest(w http.ResponseWriter, r *http.Request){
	name := r.FormValue("name")
	email := r.FormValue("email")
	message := r.FormValue("message")
	approval := r.FormValue("approval")
	log.Println(name, email, message, approval)
}
func HandlerMailboxRequest(w http.ResponseWriter, r *http.Request){
	log.Println("Mailbox was requested :))")

}
