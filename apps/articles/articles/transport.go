package articles

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// MakeHandler returns a handler for the handling service.
func MakeHandler(as Service, tracer trace.Tracer) http.Handler {
	r := mux.NewRouter()

	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
	}

	registerGetArticlesHandler := kithttp.NewServer(
		makeGetArticlesEndpoint(as, tracer),
		decodeGetArticlesRequest,
		encodeResponse,
		opts...,
	)

	registerGetArticleHandler := kithttp.NewServer(
		makeGetArticleEndpoint(as, tracer),
		decodeGetArticleRequest,
		encodeResponse,
		opts...,
	)

	r.Methods("GET").Path("/articles/v1/").Handler(registerGetArticlesHandler)
	r.Methods("GET").Path("/articles/v1/{articleID}/").Handler(registerGetArticleHandler)

	return r
}

func decodeGetArticleRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	_, span := otel.Tracer("").Start(ctx, "decodeGetArticleRequest")
	defer span.End()

	vars := mux.Vars(r)
	articleID, ok := vars["articleID"]
	if !ok {
		return nil, fmt.Errorf("articleID is missing in parameters")
	}
	includeSuggested, _ := strconv.ParseBool(r.URL.Query().Get("with_suggested"))
	return getArticleRequest{
		ID:               ArticleID(articleID),
		IncludeSuggested: includeSuggested,
	}, nil
}

func decodeGetArticlesRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	_, span := otel.Tracer("").Start(ctx, "decodeGetArticlesRequest")
	defer span.End()

	includeSuggested, _ := strconv.ParseBool(r.URL.Query().Get("with_suggested"))
	return getArticlesRequest{
		IncludeSuggested: includeSuggested,
	}, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	ctx, span := otel.Tracer("").Start(ctx, "encodeResponse")
	defer span.End()

	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

type errorer interface {
	error() error
}

func encodeError(ctx context.Context, err error, w http.ResponseWriter) {
	_, span := otel.Tracer("").Start(ctx, "encodeError")
	defer span.End()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	switch err.(type) {
	case *NotFound:
		w.WriteHeader(http.StatusNotFound)
	default:
		w.WriteHeader(http.StatusOK)
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
