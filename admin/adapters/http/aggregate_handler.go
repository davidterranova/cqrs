package http

import (
	"net/http"

	"github.com/davidterranova/cqrs/admin"
	"github.com/davidterranova/cqrs/eventsourcing"
	"github.com/davidterranova/cqrs/xhttp"
)

type AggregateHandler[T eventsourcing.Aggregate] struct {
	add *admin.App[T]
}

func NewAggregateHandler[T eventsourcing.Aggregate](add *admin.App[T]) *AggregateHandler[T] {
	return &AggregateHandler[T]{
		add: add,
	}
}

func (h *AggregateHandler[T]) LoadAggregate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	aggregateId, err := xhttp.PathParamUUID(r, "aggregate_id")
	if err != nil {
		xhttp.WriteError(ctx, w, http.StatusBadRequest, "failed to parse aggregate_id", err)
		return
	}

	toVersion, err := xhttp.QueryParamInt(r, "to_version")
	if err != nil {
		xhttp.WriteError(ctx, w, http.StatusBadRequest, "failed to parse to_version", err)
		return
	}

	aggregate, err := h.add.LoadAggregate(ctx, aggregateId, toVersion)
	if err != nil {
		xhttp.WriteError(ctx, w, http.StatusInternalServerError, "failed to load aggregate", err)
		return
	}

	xhttp.WriteObject(ctx, w, http.StatusOK, aggregate)
}

func (h *AggregateHandler[T]) RepublishAggregate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	aggregateId, err := xhttp.PathParamUUID(r, "aggregate_id")
	if err != nil {
		xhttp.WriteError(ctx, w, http.StatusBadRequest, "failed to parse aggregate_id", err)
		return
	}

	err = h.add.RepublishAggregate(ctx, aggregateId)
	if err != nil {
		xhttp.WriteError(ctx, w, http.StatusInternalServerError, "failed to republish aggregate", err)
		return
	}

	xhttp.WriteObject(ctx, w, http.StatusOK, nil)
}
