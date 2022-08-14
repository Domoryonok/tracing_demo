from flask import jsonify
from dependency_injector.wiring import inject, Provide
from suggestions.service import SuggestionsService
from suggestions.containers import Container
from tracing import Trace


@Trace()
@inject
def index(
    suggestions_service: SuggestionsService = Provide[Container.suggestions_service],
):
    return jsonify(suggestions_service.get_suggestions())


@Trace()
@inject
def get_by_article_id(
    article_id,
    suggestions_service: SuggestionsService = Provide[Container.suggestions_service],
):
    suggestions = suggestions_service.get_suggestions_for_article(article_id)
    if suggestions:
        return jsonify(suggestions), 200

    return jsonify(errors=[f"there are no suggestions for `{article_id}` article"]), 404
