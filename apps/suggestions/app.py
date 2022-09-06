import os

from flask import Flask
from suggestions.views import index, get_by_article_id
from suggestions.containers import Container
from tracing import OTLPProvider
from opentelemetry import trace


def create_app() -> Flask:
    app = Flask(__name__)

    # init container for dependency injection
    container = Container()
    container.config.data_source.file_name.from_env("DATA_SOURCE_FILENAME")

    # init dataprovider on startup
    container.suggestions_service.init()

    # init tracer
    provider = OTLPProvider("suggestions-service", os.getenv("TRACING_BACKEND_URL"))
    trace.set_tracer_provider(provider.provider)

    # setup handlers
    app.add_url_rule("/suggestions/v1", "index", index)
    app.add_url_rule(
        "/suggestions/v1/<article_id>", "get_by_article_id", get_by_article_id
    )

    return app


app = create_app()

if __name__ == "__main__":
    app.run(host=os.getenv("HOST"), port=os.getenv("PORT"))
