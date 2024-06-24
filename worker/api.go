package worker

import "github.com/go-chi/chi"

type Api struct {
	Address string
	Port    int
	Worker  *Worker
	Router  *chi.Mux
}
