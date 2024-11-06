# Backend Election
A backend for election system built with Go.

## Business Features
- Authentication & Authorization
- Voter Registration
- Candidat Registration
- Vote Transaction
- Get Eletion Result
- Manage Peer Registration

## Technical Features
- Concurrency Limit: Control the maximum number of concurrent requests.
- Rate Limiter: Protect your API from abuse by limiting request rates.
- JWT Authentication: Secure your API with JSON Web Tokens.
- RBAC Authorization: Implement role-based access control for fine-grained permissions.
- Dependency Injection Pattern: Promote modular and testable code.
- Structured Logging: Enhanced logging for errors and information.
- Environment Configuration: Option to use OS environment variables or a .env file for configuration.
- Redis Caching: Improve performance with caching.
- Graceful Shutdown: Ensure all requests complete before shutting down the server.
- CORS Handling: Manage Cross-Origin Resource Sharing.
- Clean Architecture: Maintainable and organized code structure.
- Panic Recovery Handling: Safeguard against server crashes.
- Context Error Handling: Manage request timeouts and cancellations.
- Database Migrations: Version control your database schema.
- API Testing: Ensure your API functions as expected.
- Swagger Documentation: Auto-generate API documentation for easy reference.
- Idempotent Request Handling: Ensure repeated requests yield the same result.
- Docker Support: Pre-configured Dockerfile for easy deployment.
- Matching Bimetric Fingerprint

## Getting Started
### Prerequisites
- Go version 1.23.1 or later
- Docker (optional, for containerization)
- PostgreSQL for the database

### Installation
1. Clone the repository:

```bash
git clone https://github.com/jacky-htg/backend-election.git
cd backend-election
```

2. Install dependencies:

```bash
go mod tidy
```

3. Create a .env file or set environment variables based on the provided configuration template.

4. Run the application: From your root app directory, run the command:

```bash
go run main.go
```

### API Documentation
API documentation is automatically generated and can be accessed at http://localhost:port/swagger/doc.json.

if you want to login using seed data, you can try with this payload:
```json
{
    "email": "rijal.asep.nugroho@gmail.com",
    "password": "qwertyuiop!1Q"
}
```
  
## Folder Structure
The folder structure is organized to follow the principles of Clean Architecture, ensuring that the application remains maintainable and scalable:

```bash
go-rest-api-skeleton/
├── cmd/                # CLI command for migration and etc
├── docs/               # Generated APi Doc for Swagger
├── internal/           # Application internals (domain logic, services)
│   ├── dto/            # Data transfer object to transform request into model and transform model into response
│   ├── handler/        # HTTP handlers (controllers)
│   ├── middleware/     # Middleware functions
│   ├── model/          # Data models and entities
│   ├── pkg/            # Utility functions
│   ├── repository/     # Database repository and interfaces
│   ├── route/          # API route definitions
│   └── usecase/        # Business logic and services
├── log/                # Directory for log application
├── migrations/         # Database migration scripts
├── tests/              # API tests
├── .env.example        # Example environment variables file
├── Dockerfile          # Docker configuration file
├── main.go             # Main Entrypoint for this app
└── README.md           # This README file
```

## Creating a New API
To create a new API endpoint, follow these steps:

1. Define the Model: Create a new model in the internal/model directory representing your data structure.

Example: `internal/model/candidate.go`

```go
package model

type Candidate struct {
    ID    int     `json:"id"`
    Name  string  `json:"name"`
}
```

2. Create the DTO: to transfer model with request and response payload

Example: `internal/dto/candidate_dto.go`

```go
package dto

import (
	"errors"
	"backend-election/internal/model"
)

type AddcandidateRequest struct {
	Name  string  `json:"name"`
}

func (p *AddcandidateRequest) Validate() error {
	if len(p.Name) == 0 {
		return errors.New("name is required")
	}
	return nil
}

func (p *AddcandidateRequest) ToEntity() model.Candidate {
	return model.Candidate{
		Name:  p.Name,
	}
}

type CandidateResponse struct {
	ID    int64   `json:"id"`
	Name  string  `json:"name"`
}

func (p *CandidateResponse) FromEntity(candidate model.Candidate) {
	p.ID = candidate.ID
	p.Name = candidate.Name
}
```

3. Create the Handler: Implement the handler functions in the internal/handler directory.

Example: `internal/handler/candidate.go`

```go
package handler

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"backend-election/internal/dto"
	"backend-election/internal/model"
	"backend-election/internal/pkg/httpresponse"
	"backend-election/internal/pkg/logger"
	"backend-election/internal/repository"

	"github.com/bytedance/sonic"
	"github.com/julienschmidt/httprouter"
	"go.opentelemetry.io/otel"
)

// Candidates handler
type Candidates struct {
	Log *logger.Logger
	DB  *sql.DB
}

// @Security Bearer
// @Summary Add Candidate
// @Description Add Candidate
// @Tags Candidates
// @Accept  json
// @Produce  json
// @Param request body dto.AddCandidateRequest true "Candidate to add"
// @Param Authorization header string true "Bearer token"
// @Success 201 {object} dto.CandidateResponse
// @Router /candidates [post]
func (h *Candidates) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var ctx = r.Context()
	
	defer r.Body.Close()
	var httpres = httpresponse.Response{}
	var candidateRequest dto.AddCandidateRequest
	err := sonic.ConfigDefault.NewDecoder(r.Body).Decode(&candidateRequest)
	if err != nil {
		h.Log.Error(err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if err := candidateRequest.Validate(); err != nil {
		h.Log.Error(err)
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	var candidateRepo = repository.CandidateRepository{Log: h.Log, Db: h.DB}
	candidateRepo.CandidteEntity = candidateRequest.ToEntity()
	if err := candidateRepo.Save(ctx); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var response dto.CandidateResponse
	response.FromEntity(candidateRepo.CandidateEntity)
	httpres.SetMarshal(ctx, w, http.StatusCreated, response, "")
}

```

4. Create Respository

Example: `internal/repository/candidate_repository.go`

```go
package repository

import (
	"context"
	"database/sql"
	"backend-election/internal/model"
)

type CandidateRepository struct {
	Log           *logger.Logger
	Db            *sql.DB
	CandidateEntity model.Candidate
}

func (r *CandidateRepository) Save(ctx context.Context) error {
	// Implementasi simpan candidate ke database
	return nil
}
```

5. Define Routes: Add new routes in the internal/route directory to map HTTP requests to your handler functions.

Example: `internal/route/route.go`

```go
func ApiRoute(log *logger.Logger, db *database.Database, cache *redis.Cache) *httprouter.Router {
    // .... existing code
    candidateHandler := handler.Candidates{Log: log, DB: db.Conn}
    router.POST("/candidates", mid.WrapMiddleware(privateMiddlewares, candidateHandler.Create))
    // .... existing code
}
```

6. Create swagger documentation with command `swag init`. Attention to install `go install github.com/swaggo/swag/cmd/swag@latest` before you run `swag init`. 

7. Testing: Write tests for your new API endpoint in the tests directory to ensure it behaves as expected.

## Creating Migration Scripts
When adding a new database migration for your product model, follow these steps and naming conventions:

1. Naming Convention: Use the following prefixes for your migration scripts:

- For functions: `1.001_fn_random_bigint.sql`
- For tables: `2.001_t_access.sql`
- For seeding: `3.001_seed.sql`

After the prefix, include a three-digit serial number followed by a descriptive name for the migration.

2. Create a Migration for Candidates: For the candidates table, create a migration file in the migrations directory named `2.006_t_candidates.sql`.

Example: `migrations/2.006_t_candidates.sql`

```sql
CREATE TABLE candidates (
    id int8 DEFAULT int64_id('candidates'::text, 'id'::text) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NULL,
);
```

3. Functions or Seeds: If you need to create a function or a seed, follow the same naming conventions for your migration files.

Example function file: `1.001_fn_add_tax.sql`

```sql
CREATE OR REPLACE FUNCTION add_tax(price NUMERIC)
RETURNS NUMERIC AS $$
BEGIN
    RETURN price * 1.1;  -- Adds a 10% tax
END;
$$ LANGUAGE plpgsql;
```

Example seed file: `3.001_seed.sql`

```sql
INSERT INTO candidates (name) VALUES
('Alice'),
('Bob');
```

4. Running migration command using `go run cmd/main.go migrate`

## Running Tests
To run the API tests, use the following command:

```bash
go test ./...
```

## Contributing
Contributions are welcome! Please fork the repository and submit a pull request for any enhancements or fixes.

## License
This project is licensed under GNU GPL V3 License. See the [LICENSE](./LICENSE) file for details.