package http

import (
	"net/http"

	"github.com/davidterranova/cqrs/admin"
	"github.com/davidterranova/cqrs/eventsourcing"
	"github.com/davidterranova/cqrs/xhttp"
	"github.com/google/uuid"
)

type EventHandler[T eventsourcing.Aggregate] struct {
	app *admin.App[T]
}

func NewEventHandler[T eventsourcing.Aggregate](app *admin.App[T]) *EventHandler[T] {
	return &EventHandler[T]{
		app: app,
	}
}

func (h *EventHandler[T]) ListEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filter := make([]eventsourcing.EventQueryOption, 0)
	aggregateId, err := xhttp.QueryParamUUID(r, "aggregate_id")
	if err != nil {
		xhttp.WriteError(r.Context(), w, http.StatusBadRequest, "failed to parse aggregate_id", err)
		return
	}
	if aggregateId != uuid.Nil {
		filter = append(filter, eventsourcing.EventQueryWithAggregateId(aggregateId))
	}

	strAggregateType, err := xhttp.QueryParamStr(r, "aggregate_type")
	if err != nil {
		xhttp.WriteError(r.Context(), w, http.StatusBadRequest, "failed to parse aggregate_type", err)
		return
	}
	if strAggregateType != "" {
		filter = append(filter, eventsourcing.EventQueryWithAggregateType(eventsourcing.AggregateType(strAggregateType)))
	}

	eventType, err := xhttp.QueryParamStr(r, "event_type")
	if err != nil {
		xhttp.WriteError(r.Context(), w, http.StatusBadRequest, "failed to parse event_type", err)
		return
	}
	if eventType != "" {
		filter = append(filter, eventsourcing.EventQueryWithEventType(
			eventsourcing.EventType(eventType),
		))
	}

	published, err := xhttp.QueryParamBool(r, "published")
	if err != nil {
		xhttp.WriteError(r.Context(), w, http.StatusBadRequest, "failed to parse published", err)
		return
	}
	if published != nil {
		filter = append(filter, eventsourcing.EventQueryWithPublished(*published))
	}

	events, err := h.app.ListEvent(
		r.Context(),
		eventsourcing.NewEventQuery(filter...),
	)
	if err != nil {
		xhttp.WriteError(ctx, w, http.StatusInternalServerError, "failed to list events", err)
		return
	}

	xhttp.WriteObject(ctx, w, http.StatusOK, fromEventInternalSlice(events))
}
