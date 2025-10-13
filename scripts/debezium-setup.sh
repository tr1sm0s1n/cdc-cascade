#!/bin/sh

echo "Waiting for Debezium Connect to be ready..."
until curl -f http://$DEBEZIUM_HOST:$DEBEZIUM_PORT/ >/dev/null 2>&1; do
	echo "Waiting for Debezium..."
	sleep 5
done

echo "Debezium is ready! Registering PostgreSQL connector..."

curl -i -X POST -H "Accept:application/json" -H "Content-Type:application/json" \
	$DEBEZIUM_HOST:$DEBEZIUM_PORT/connectors/ -d '{
  "name": "postgres-connector",
  "config": {
    "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
    "database.hostname": "'"${DB_HOST}"'",
    "database.port": "'"${DB_PORT}"'",
    "database.user": "'"${DB_USER}"'",
    "database.password": "'"${DB_PASS}"'",
    "database.dbname": "'"${DB_NAME}"'",
    "database.server.name": "'"${DB_HOST}"'",
    "table.include.list": "public.sinners",
    "plugin.name": "pgoutput",
    "publication.name": "dbz_publication",
    "slot.name": "debezium_slot",
    "topic.prefix": "'"${DB_HOST}"'"
  }
}'

echo -e "\n\nDebezium connector registered successfully!"
