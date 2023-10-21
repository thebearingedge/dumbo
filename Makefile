include .env

MAKEFLAGS += --no-print-directory
.DEFAULT_GOAL := test

.PHONY: migrate admin tdd test cover

migrate:
	docker compose run --rm migrate $(filter-out $@,$(MAKECMDGOALS))

admin:
	docker compose up --detach admin

tdd:
	DATABASE_URL=$(DATABASE_URL) gow -c test

test:
	DATABASE_URL=$(DATABASE_URL) go test -count=1 -v

cover:
	DATABASE_URL=$(DATABASE_URL) go test -count=1 -v -coverprofile .coverage/dumbo.out
	go tool cover -html=.coverage/dumbo.out -o .coverage/dumbo.html

stop:
	docker compose down -v --remove-orphans

# we're doing this to use make as a task runner instead of a build system
# this last recipe allows us to pass extra arguments to recipes
# e.g. `make migrate version` to print the current migration number
# https://stackoverflow.com/a/6273809/3100405
%:
	@:
