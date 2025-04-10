module github.com/pyculiar/chirpy

go 1.24.2

require (
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
	internal/auth v1.0.0
	internal/database v1.0.0
)

require github.com/google/uuid v1.6.0

require golang.org/x/crypto v0.37.0 // indirect

replace internal/database => ./internal/database

replace internal/auth => ./internal/auth
