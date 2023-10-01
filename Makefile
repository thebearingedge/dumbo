include .env

MAKEFLAGS += --no-print-directory
.DEFAULT_GOAL := test

DATABASE_URL := "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable"

.PHONY: up down admin tdd test cover

up:
	docker compose run migrate up

down:
	docker compose run migrate down

admin:
	docker compose up --detach admin

tdd:
	DATABASE_URL=$(DATABASE_URL) gow -c test ./...

test:
	DATABASE_URL=$(DATABASE_URL) go test -count=1 -v ./...

cover:
	DATABASE_URL=$(DATABASE_URL) go test -count=1 -v ./... -coverprofile .coverage/dumbo.out
	go tool cover -html=.coverage/dumbo.out -o .coverage/dumbo.html
