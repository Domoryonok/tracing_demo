import os
import uuid
import json
import random
from faker import Faker
import click

fake = Faker()


@click.group()
def cli():
    pass


def generate_articles(number_of_articles):
    articles = {}

    for _ in range(number_of_articles):
        article_id = uuid.uuid4().hex
        articles[article_id] = {
            "id": article_id,
            "author": fake.name(),
            "title": fake.sentence(nb_words=4),
            "text": fake.text(),
        }
    return articles


def generate_suggestions(articles_keys, number_of_suggestions):
    suggestions = {}

    def get_suggestions(key, articles_keys):
        suggestions = random.choices(articles_keys, k=number_of_suggestions)
        if key not in suggestions:
            return suggestions

        return get_suggestions(key, articles_keys)

    for key in articles_keys:
        suggestions[key] = get_suggestions(key, articles_keys)

    return suggestions


@cli.command()
@click.option("-na", help="Number of articles.", default=10)
@click.option("-ns", help="Number of suggestions per article.", default=3)
def generate(na, ns):
    if ns >= na:
        raise Exception(
            "Number of suggestions per article cannot be equal or greater then number of articles"
        )

    articles_file_name = os.path.join("apps/articles/mocks", "articles.json")
    suggestions_file_name = os.path.join("apps/suggestions/mocks", "suggestions.json")

    articles = generate_articles(na)
    suggestions = generate_suggestions(list(articles.keys()), ns)

    with open(articles_file_name, "w") as file:
        file.write(json.dumps(articles, indent=4))

    with open(suggestions_file_name, "w") as file:
        file.write(json.dumps(suggestions, indent=4))


if __name__ == "__main__":
    cli()
