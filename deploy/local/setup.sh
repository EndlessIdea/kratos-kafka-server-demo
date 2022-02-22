#!/bin/sh
docker-compose -f docker_kafka.yaml up -d
docker exec broker \
kafka-topics --bootstrap-server broker:9092 \
             --create \
             --topic location
docker exec broker \
kafka-topics --bootstrap-server broker:9092 \
             --create \
             --topic vehicle
docker exec --interactive --tty broker \
kafka-console-producer --bootstrap-server broker:9092 \
                       --topic location
docker exec --interactive --tty broker \
kafka-console-producer --bootstrap-server broker:9092 \
                       --topic vehicle