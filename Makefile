install-tools:
	if [ ! $$(which go) ]; then \
		echo "goLang not found."; \
		echo "Try installing go..."; \
		exit 1; \
	fi
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.15.1
	if [ ! $$( which migrate ) ]; then \
		echo "The 'migrate' command was not found in your path. You most likely need to add \$$HOME/go/bin to your PATH."; \
		exit 1; \
	fi



tidy:
	go mod tidy

test: tidy
	go test ./...

build:
	mkdir -p ./bin
	GOOS=linux go build -o bin/api ./cmd/api/api.go

package:
	docker  build -t $(tag) . 

run:database
	go mod tidy
	if [ ! -f '.env' ]; then \
		cp .env.example .env; \
	fi
	go run ./cmd/api/api.go




create-migration: ## usage: make name=new create-migration
	migrate create -ext sql -dir ./db/migrations -seq $(name)

database:
	# if [ "$$( docker container inspect -f '{{.State.Running}}' cliqets-api-db )" != "true" ]; then \
	# 	docker run -d --name sigma-api-db  -e POSTGRES_USER=sigma -e POSTGRES_PASSWORD=sigma@123 -e POSTGRES_DB=sigma-auto -e PGDATA=/var/lib/postgresql/data/pgdata/cliqets -v $$(pwd)/pg_data:/var/lib/postgresql/data/cliqets  -p 5432:5432 -it postgres:14; \
	# fi

	# if [ "$$( docker container inspect -f '{{.State.Running}}' pg-admin )" != "true" ]; then \
	# 	docker run -d --name pg-admin --rm -e PGADMIN_DEFAULT_PASSWORD=12345678 -e PGADMIN_DEFAULT_EMAIL=silasmagho18@gmail.com --add-host=host.docker.internal:host-gateway -p 3000:80 -it dpage/pgadmin4; \
	# fi
	docker-compose up -d
	
gen:
	go mod tidy
	go generate ./...
