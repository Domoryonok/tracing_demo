package articles

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"go.opentelemetry.io/otel/trace"
)

type getArticleRequest struct {
	ID               ArticleID
	IncludeSuggested bool
}

type getArticleResponse struct {
	Article Article `json:"article"`
}

type getArticlesRequest struct {
	IncludeSuggested bool
}

type getArticlesResponse struct {
	Articles []Article `json:"articles"`
}

func makeGetArticleEndpoint(as Service, tracer trace.Tracer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		ctx, span := tracer.Start(ctx, "get article endpoint")
		defer span.End()

		req := request.(getArticleRequest)
		article, err := as.GetArticle(ctx, req.ID)
		if err != nil {
			return nil, err
		}

		if req.IncludeSuggested {
			s, err := as.GetSuggested(ctx, req.ID)
			if err != nil {
				return nil, err
			}

			article.SuggestedArticles = s
		}

		return getArticleResponse{
			Article: article,
		}, nil
	}
}

func makeGetArticlesEndpoint(as Service, tracer trace.Tracer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		ctx, span := tracer.Start(ctx, "get articles endpoint")
		defer span.End()

		req := request.(getArticlesRequest)

		articles, err := as.GetArticles(ctx)
		if err != nil {
			return nil, err
		}

		if req.IncludeSuggested {
			for idx := range articles {
				s, err := as.GetSuggested(ctx, articles[idx].ID)
				if err != nil {
					return nil, err
				}

				articles[idx].SuggestedArticles = s
			}
		}

		return getArticlesResponse{
			articles,
		}, nil
	}
}
