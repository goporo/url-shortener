# URL Shortener Service

A robust URL shortener service built with Golang that converts long URLs into short, easy-to-share links. The service includes rate limiting, usage statistics tracking, and a RESTful API.

## Features

- **URL Shortening**: Convert long URLs to short codes using base62 encoding
- **Rate Limiting**: Prevents API abuse with configurable request limits
- **Usage Statistics**: Track clicks, referrers, and other metrics for each shortened URL
- **RESTful API**: Complete API for managing shortened URLs
- **Persistence**: Data stored in a database for reliability

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET    | `/urls` | Retrieve all shortened URLs |
| POST   | `/urls` | Create a new shortened URL |
| GET    | `/urls/:shortCode` | Redirect to the original URL |
| PUT    | `/urls/:shortCode` | Update an existing shortened URL |
| DELETE | `/urls/:shortCode` | Delete a shortened URL |
| GET    | `/urls/:shortCode/stats` | Get usage statistics for a specific URL |

## How It Works

### URL Shortening Process

The service uses base62 encoding (a-z, A-Z, 0-9) to generate short codes for URLs. This approach allows for:
- Creating shorter URLs than with hexadecimal encoding
- Generating up to 62^n unique URLs where n is the length of the short code
- Avoiding confusing characters like 'O' and '0'

### Rate Limiting

To prevent abuse, the service implements rate limiting on API requests:
- Configurable limits by IP address or API key
- Sliding window algorithm for fair usage calculation
- Clear rate limit headers in API responses

### Usage Statistics

For each shortened URL, the service tracks:
- Total clicks
- Unique visitors
- Referrer sources
- Geographic data
- Browser/device information
- Timestamp of visits

## API Usage Examples

### Create a Shortened URL

```bash
curl -X POST http://localhost:8080/urls \
  -H "Content-Type: application/json" \
  -d '{"original_url": "https://example.com/very/long/url/that/needs/shortening"}'
```

Response:
```json
{
  "original_url": "https://example.com/very/long/url/that/needs/shortening",
  "short_code": "aBc123",
  "short_url": "http://localhost:8080/urls/aBc123",
  "created_at": "2023-05-20T15:30:45Z"
}
```

### Get Statistics for a URL

```bash
curl -X GET http://localhost:8080/urls/aBc123/stats
```

Response:
```json
{
  "short_code": "aBc123",
  "original_url": "https://example.com/very/long/url/that/needs/shortening",
  "created_at": "2023-05-20T15:30:45Z",
  "stats": {
    "total_clicks": 42,
    "unique_visitors": 24,
    "referrers": {
      "google.com": 15,
      "twitter.com": 12,
      "direct": 15
    },
    "daily_clicks": [
      {"date": "2023-05-20", "clicks": 10},
      {"date": "2023-05-21", "clicks": 32}
    ]
  }
}
```

## Installation

### Prerequisites

- Go 1.16 or higher
- Database (PostgreSQL, MySQL, or Redis)
- Docker (optional)

### Local Setup

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/url-shortener.git
   cd url-shortener
   ```

2. Install dependencies:
   ```
   go mod download
   ```

3. Configure the application (edit config.yaml or use environment variables)

4. Run the application:
   ```
   go run main.go
   ```

### Docker Setup

```bash
docker build -t url-shortener .
docker run -p 8080:8080 url-shortener
```

## Configuration

The service can be configured via environment variables or a config file:

```yaml
server:
  port: 8080
  host: "localhost"

database:
  type: "postgres"
  connection: "postgresql://user:password@localhost:5432/urlshortener"

ratelimit:
  requests: 100
  period: "1h"

shortcode:
  length: 6
```

## Implementation Details

### Base62 Encoding

The URL shortening service uses base62 encoding to generate compact short codes. The implementation converts a numeric ID to a character string using the following character set:

```
0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz
```

Here's a simplified example of the encoding function:

```go
func base62Encode(number uint64) string {
    const base = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
    length := uint64(len(base))
    var encodedBuilder strings.Builder
    
    for number > 0 {
        encodedBuilder.WriteByte(base[number%length])
        number = number / length
    }
    
    encoded := encodedBuilder.String()
    return reverse(encoded)
}
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.
