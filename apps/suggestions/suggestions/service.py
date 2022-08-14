import json
from typing import List, Dict
from tracing import Trace


@Trace()
class SuggestionsService:
    def __init__(self, filename) -> None:
        self.suggestions: Dict[str, List[str]] = {}
        with open(filename) as file:
            self.suggestions = json.load(file)

    def get_suggestions(self) -> Dict[str, List[str]]:
        return self.suggestions

    def get_suggestions_for_article(self, article_id: str) -> List[str]:
        return self.suggestions.get(article_id)
