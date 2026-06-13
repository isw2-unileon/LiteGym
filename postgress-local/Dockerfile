FROM postgres:15-alpine

COPY --chown=postgres:postgres initdb/ /docker-entrypoint-initdb.d/

RUN chmod -R 755 /docker-entrypoint-initdb.d
