FROM docker.io/library/python:3.13.5-slim-bookworm

RUN pip install flask && rm -rf /root/.cache

ADD ./main.py /main.py

EXPOSE 8080

WORKDIR /

ENTRYPOINT ["flask", "--app", "main", "run"]