#!/bin/bash -e
# I don't use schemaspy to generate an ER diagram anymore, I use DBeaver's ER
# diagram feature. DBeaver also serves as a GUI alternative to psql for running
# SQL commands against your database.

mkdir -p .schemaspy_output
docker run -v "$PWD/.schemaspy_output:/output" schemaspy/schemaspy:snapshot -t pgsql -host host.docker.internal -port 5432 -db skylab_devx -u bokwoon
