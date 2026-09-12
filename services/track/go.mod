module shopass/services/track

go 1.25.0

toolchain go1.25.13

replace shopass/services/price => ../price

replace shopass/services/track => ./

require (
	github.com/jackc/pgx/v5 v5.11.0
	github.com/lib/pq v1.12.3
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	golang.org/x/text v0.40.0 // indirect
)
