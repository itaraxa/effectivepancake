#!/bin/bash

go build -o cmd/server/server cmd/server/*.go
go build -o cmd/agent/agent cmd/agent/*.go