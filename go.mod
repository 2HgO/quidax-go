module github.com/2HgO/quidax-go

go 1.22.4

require (
	github.com/Masterminds/squirrel v1.5.4
	github.com/creasty/defaults v1.8.0
	github.com/go-playground/validator/v10 v10.22.1
	github.com/go-sql-driver/mysql v1.8.1
	github.com/google/uuid v1.6.0
	github.com/gorilla/schema v1.4.1
	github.com/lucsky/cuid v1.2.1
	github.com/tigerbeetle/tigerbeetle-go v0.16.11
	go.uber.org/fx v1.23.0
	go.uber.org/zap v1.26.0
	golang.org/x/crypto v0.28.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	go.uber.org/dig v1.18.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
)

replace github.com/tigerbeetle/tigerbeetle-go => ../tigerbeetle-go
