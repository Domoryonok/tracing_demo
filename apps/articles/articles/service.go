package articles

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Service interface {
	GetArticle(cxt context.Context, id ArticleID) (Article, error)
	GetSuggested(cxt context.Context, id ArticleID) ([]Article, error)
	GetArticles(cxt context.Context) ([]Article, error)
}

type service struct {
	articles               map[ArticleID]*Article
	client                 *http.Client
	suggestionsServiceHost string
	tracer                 trace.Tracer
}

func (s *service) GetArticle(ctx context.Context, id ArticleID) (Article, error) {
	_, span := s.tracer.Start(ctx, "get article")
	defer span.End()

	span.SetAttributes(
		attribute.String("article.id", string(id)),
	)

	article, found := s.articles[id]
	if !found {
		return Article{}, &NotFound{
			ArticleID: id,
		}
	}

	return *article, nil
}

func (s *service) GetSuggested(ctx context.Context, id ArticleID) ([]Article, error) {
	ctx, span := s.tracer.Start(ctx, "get suggestion")
	defer span.End()

	span.SetAttributes(
		attribute.String("article.id", string(id)),
	)

	// build request object with tracing context
	u, err := url.Parse(fmt.Sprintf("%s/suggestions/v1/%s", s.suggestionsServiceHost, id))
	if err != nil {
		return nil, err
	}

	req := &http.Request{
		Method: http.MethodGet,
		URL:    u,
	}
	req = req.WithContext(ctx)

	// fetch suggestions from the suggestions microservice
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	var suggestedArticlesIDs []ArticleID
	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(d, &suggestedArticlesIDs)
	if err != nil {
		return nil, err
	}

	var suggestedArticles []Article

	for _, id := range suggestedArticlesIDs {
		article, err := s.GetArticle(ctx, id)
		if err != nil {
			return nil, err
		}

		suggestedArticles = append(suggestedArticles, article)
	}

	return suggestedArticles, nil
}

func (s *service) GetArticles(ctx context.Context) ([]Article, error) {
	_, span := s.tracer.Start(ctx, "get articles")
	defer span.End()

	var articles []Article

	for _, article := range s.articles {
		articles = append(articles, *article)
	}

	return articles, nil
}

func NewService(suggestionsServiceHost, articlesDataPath string, client *http.Client, tracer trace.Tracer) (Service, error) {
	var articles map[ArticleID]*Article

	// store articles list in memory
	file, err := os.ReadFile(articlesDataPath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(file), &articles)
	if err != nil {
		return nil, err
	}

	return &service{
		articles:               articles,
		client:                 client,
		suggestionsServiceHost: suggestionsServiceHost,
		tracer:                 tracer,
	}, nil
}
