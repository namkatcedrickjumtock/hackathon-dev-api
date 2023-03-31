# Postgresql & PgAdmin powered by compose

## Requirements

* docker >= 17.12.0+
* docker-compose

## to install docker-compose

* Run this command `sudo apt-get update`
                   `sudo apt-get install docker-compose-plugin`
                   `sudo apt  install docker-compose`

## Quick Start

* Run this command `docker-compose up -d`

## Environments

This Compose file contains the following environment variables:

* `POSTGRES_USER` the default value is **postgres**
* `POSTGRES_PASSWORD` the default value is **postgres**
* `PGADMIN_PORT` the default value is **5050**
* `PGADMIN_DEFAULT_EMAIL` the default value is **silasmagho18@gmail.com**
* `PGADMIN_DEFAULT_PASSWORD` the default value is **12345678**

## Access to postgres

* `localhost:5432`
* **Username:** postgres (as a default)
* **Password:** postgres (as a default)

## Access to PgAdmin

* **URL:** `http://localhost:5050`
* **Username:** silasmagho18@gmail.com (as a default)
* **Password:** 12345678 (as a default)

## Add a new server in PgAdmin

* **Host name/address** `postgres`
* **Port**  `5432`
* **Username** as `POSTGRES_USER`, by default: `postgres`
* **Password** as `POSTGRES_PASSWORD`, by default `12345678`

[reference](https://github.com/khezen/compose-postgres/pull/23/files)
