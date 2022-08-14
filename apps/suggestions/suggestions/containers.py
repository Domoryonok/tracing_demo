from dependency_injector import containers, providers
from suggestions.service import SuggestionsService


class Container(containers.DeclarativeContainer):
    wiring_config = containers.WiringConfiguration(modules=["suggestions.views"])
    config = providers.Configuration()

    suggestions_service = providers.Resource(
        SuggestionsService, filename=config.data_source.file_name
    )
