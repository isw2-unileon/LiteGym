FROM postgres:15-alpine

COPY initdb/ /docker-entrypoint-initdb.d/

RUN chmod -R a+rX /docker-entrypoint-initdb.d
