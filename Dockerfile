FROM ubuntu:20.04

WORKDIR /app

RUN apt update
RUN apt install -y python3 python3-pip

COPY ./requirements.txt ./
RUN pip3 install -r requirements.txt
COPY ./bhvcpy ./bhvcpy
COPY ./srv ./srv
COPY ./static ./static
COPY ./templates ./templates
COPY ./manage.py ./

EXPOSE 8000

CMD ["python3.8", "manage.py", "runserver", "--noreload"]