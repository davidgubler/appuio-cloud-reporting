# APPUiO Cloud Reporting

## Use APPUiO Global instance

```sh
# Follow the login instructions to get a token
oc login --server=https://api.cloudscale-lpg-2.appuio.cloud:6443

kubectl -n appuio-reporting port-forward svc/reporting-db 5432 &

DB_USER=$(kubectl -n appuio-reporting get secret/reporting-db-superuser -o jsonpath='{.data.user}' | base64 --decode)
DB_PASSWORD=$(kubectl -n appuio-reporting get secret/reporting-db-superuser -o jsonpath='{.data.password}' | base64 --decode)

# Check for pending migrations
export DB_URL="postgres://${DB_USER}:${DB_PASSWORD}@localhost/reporting?sslmode=disable"
go run ./cmd/migrate -show-pending

# Connect to the database's interactive terminal
export PGPASSWORD=$(kubectl -n appuio-reporting get secret/reporting-db-superuser -o jsonpath='{.data.password}' | base64 --decode)
psql -U "${DB_USER}" -w -h localhost reporting
```

## Local Installation

```sh
SUPERUSER_PW=$(ruby -rsecurerandom -e "puts SecureRandom.alphanumeric(40)")

kubectl create ns appuio-reporting
kubectl -nappuio-reporting create secret generic reporting-db-superuser --from-literal=user=reporting-db-superuser "--from-literal=password=${SUPERUSER_PW}"
kubectl -nappuio-reporting apply -k manifests/base
```

## Usage

### Run Test Query

```sh
kubectl -n appuio-reporting port-forward svc/reporting-db 5432 &

DB_USER=$(kubectl -n appuio-reporting get secret/reporting-db-superuser -o jsonpath='{.data.user}' | base64 --decode)
DB_PASSWORD=$(kubectl -n appuio-reporting get secret/reporting-db-superuser -o jsonpath='{.data.password}' | base64 --decode)
export DB_URL="postgres://${DB_USER}:${DB_PASSWORD}@localhost/reporting?sslmode=disable"

go run ./cmd/testreport
```

### Migrate to Most Recent Schema

```sh
kubectl -n appuio-reporting port-forward svc/reporting-db 5432 &

DB_USER=$(kubectl -n appuio-reporting get secret/reporting-db-superuser -o jsonpath='{.data.user}' | base64 --decode)
DB_PASSWORD=$(kubectl -n appuio-reporting get secret/reporting-db-superuser -o jsonpath='{.data.password}' | base64 --decode)
export DB_URL="postgres://${DB_USER}:${DB_PASSWORD}@localhost/reporting?sslmode=disable"

go run ./cmd/migrate -show-pending

go run ./cmd/migrate
```

### Connect to the Database

```sh
kubectl -n appuio-reporting port-forward svc/reporting-db 5432 &

DB_USER=$(kubectl -n appuio-reporting get secret/reporting-db-superuser -o jsonpath='{.data.user}' | base64 --decode)
export PGPASSWORD=$(kubectl -n appuio-reporting get secret/reporting-db-superuser -o jsonpath='{.data.password}' | base64 --decode)

psql -U "${DB_USER}" -w -h localhost reporting
```
