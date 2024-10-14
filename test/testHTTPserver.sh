#!/bin/bash

interval=1

curls=(
"curl -v -X POST 'http://localhost:8080'"
"curl -v -X POST 'http://localhost:8080/update/counter/testCounter/100'"
"curl -v -X POST 'http://localhost:8080/update/unknown/testCounter/100'"
"curl -v -X POST 'http://localhost:8080/update/gauge/test1/88.88' -H 'Content-Type: text/html' -d ''"
"curl -v -X POST 'http://localhost:8080/update/gauge/test1/88.102' -H 'Content-Type: text/html' -d ''"
"curl -v -X POST 'http://localhost:8080/update/gauge/test2/-32.102' -H 'Content-Type: text/html' -d ''"
"curl -v -X POST 'http://localhost:8080/update/counter/test1/4' -H 'Content-Type: text/html' -d ''"
"curl -v -X POST 'http://localhost:8080/update/counter/test1/5' -H 'Content-Type: text/html' -d ''"
)

for curl_cmd in "${curls[@]}"
do
  eval "$curl_cmd"
  sleep $interval
done
