module shopass/services/track

go 1.25.0

replace shopass/services/price => ../price

replace shopass/services/track => ./

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.10.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	shopass/services/price v0.0.0-00010101000000-000000000000 // indirect
)
