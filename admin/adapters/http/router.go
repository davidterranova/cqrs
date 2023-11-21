package http

import (
	"github.com/davidterranova/cqrs/admin"
	"github.com/davidterranova/cqrs/eventsourcing"
	"github.com/gorilla/mux"
)

func New[T eventsourcing.Aggregate](root *mux.Router, app *admin.App[T]) *mux.Router {
	aggregateHandler := NewAggregateHandler[T](app)

	root.HandleFunc("/v1/aggregates/{aggregate_id}", aggregateHandler.LoadAggregate).Methods("GET")
	root.HandleFunc("/v1/aggregates/{aggregate_id}:republish", aggregateHandler.RepublishAggregate).Methods("POST")

	eventHandler := NewEventHandler[T](app)

	root.HandleFunc("/v1/events", eventHandler.ListEvent).Methods("GET")

	return root
}
