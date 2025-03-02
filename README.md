# URL Shortener API

A URL shortening service with API documentation using Swagger.

## Features

- Create short URLs
- Redirect to original URLs
- Update existing short URLs
- Delete short URLs
- View URL statistics
- Get list of all shortened URLs
- API documentation with Swagger UI

## Getting Started

### Prerequisites

- Go 1.16 or higher
- PostgreSQL database

### Installation

1. Clone the repository
2. Install dependencies:

```bash
go mod download
```

3. Set up environment variables in a `.env` file:

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=url_shortener
PORT=8080
```

4. Generate Swagger documentation:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init
```

5. Run the application:

```bash
go run main.go
```

### API Documentation

Access the Swagger UI documentation at: 

```
http://localhost:8080/swagger/index.html
```

## API Endpoints

- `GET /`: Home page with link to API documentation
- `GET /urls`: Get all shortened URLs
- `POST /urls`: Create a new short URL
- `GET /urls/:shortCode`: Redirect to original URL
- `PUT /urls/:shortCode`: Update an existing short URL
- `DELETE /urls/:shortCode`: Delete a short URL
- `GET /urls/:shortCode/stats`: Get statistics for a short URL

## License

This project is licensed under the MIT License - see the LICENSE file for details.
