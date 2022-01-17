# APPUiO Cloud Reporting

## Use APPUiO Global instance

```sh
# Follow the login instructions to get a token
oc login --server=https://api.cloudscale-lpg-2.appuio.cloud:6443

# Forward database and thanos to local host
kubectl -n appuio-reporting port-forward svc/reporting-db 5432 &
kubectl --as=cluster-admin -n appuio-thanos port-forward svc/thanos-query 9090 &

# Check for pending migrations
DB_USER=$(kubectl -n appuio-reporting get secret/reporting-db-superuser -o jsonpath='{.data.user}' | base64 --decode)
DB_PASSWORD=$(kubectl -n appuio-reporting get secret/reporting-db-superuser -o jsonpath='{.data.password}' | base64 --decode)
export DB_URL="postgres://${DB_USER}:${DB_PASSWORD}@localhost/reporting?sslmode=disable"
go run ./cmd/migrate -show-pending

# Run a query
go run ./cmd/report ping "2022-01-17T09:00:00Z"

# Connect to the database's interactive terminal
DB_USER=$(kubectl -n appuio-reporting get secret/reporting-db-superuser -o jsonpath='{.data.user}' | base64 --decode)
export PGPASSWORD=$(kubectl -n appuio-reporting get secret/reporting-db-superuser -o jsonpath='{.data.password}' | base64 --decode)
psql -U "${DB_USER}" -w -h localhost reporting
```

## Local Installation

```sh
SUPERUSER_PW=$(pwgen 40 1)

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

## Local Development

Local development assumes a locally installed PostgreSQL database.

```sh
createdb appuio-cloud-reporting-test
export DB_URL="postgres://localhost/appuio-cloud-reporting-test?sslmode=disable"

go run ./cmd/migrate
go run ./cmd/migrate -seed
go test ./...
```

### IDE Integration

To enable IDE Test/Debug support, `DB_URL` should be added to the test environment.

#### VSCode

```sh
mkdir -p .vscode
touch .vscode/settings.json
jq -s '(.[0] // {}) | ."go.testEnvVars"."DB_URL" = $ENV."DB_URL"' .vscode/settings.json > .vscode/settings.json.i
mv .vscode/settings.json.i .vscode/settings.json
```
