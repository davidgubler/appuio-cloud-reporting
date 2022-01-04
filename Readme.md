# APPUiO Cloud Reporting

## Installation

```sh
SUPERUSER_PW=$(ruby -rsecurerandom -e "puts SecureRandom.alphanumeric(40)")

kubectl create ns appuio-invoicing
kubectl -nappuio-invoicing create secret generic invoicing-db-superuser --from-literal=user=invoicing-db-superuser "--from-literal=password=${SUPERUSER_PW}"
kubectl -nappuio-invoicing apply -f database-sts.yaml
```

## Usage

```sh
kubectl -n appuio-invoicing port-forward svc/invoicing-db 5432 &

DB_USER=$(kubectl -n appuio-invoicing get secret/invoicing-db-superuser -o jsonpath='{.data.user}' | base64 --decode)
DB_PASSWORD=$(kubectl -n appuio-invoicing get secret/invoicing-db-superuser -o jsonpath='{.data.password}' | base64 --decode)
export DB_URL="postgres://${DB_USER}:${DB_PASSWORD}@localhost/${DB_USER}?sslmode=disable"

go run ./cmd/migrate
```
