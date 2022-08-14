package articles

type ArticleID string

type Article struct {
	ID                ArticleID `json:"id"`
	Author            string    `json:"author"`
	Title             string    `json:"title"`
	Text              string    `json:"text"`
	SuggestedArticles []Article `json:"suggested,omitempty"`
}
