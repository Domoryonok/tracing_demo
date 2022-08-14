package articles

import "fmt"

type NotFound struct {
	ArticleID ArticleID
}

func (nf *NotFound) Error() string {
	return fmt.Sprintf("article `%s` was not found", nf.ArticleID)
}
