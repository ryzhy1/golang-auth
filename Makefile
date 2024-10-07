run:
	go run cmd/sso/main.go config=./config/dev.yaml
migrate:
	go run cmd/migrate.go