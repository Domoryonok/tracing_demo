FROM python:3.9-slim-buster

WORKDIR /suggestions-microservice

COPY . .

RUN pip3 install -r requirements/requirements.txt

CMD ["python3", "app.py"]
