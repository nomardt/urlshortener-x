# urlshortener-x

URL Shortening Service Features:
- JWT authentication
- URL backup available in either a local file or PostgreSQL database
- Concurrent URL deletion
- Custom logger
- Endpoints for text/plain, application/json MIME types

# Example usage
```bash
go run cmd/shortener/main.go
```
### OR
```bash
go run cmd/shortener/main.go -d 'postgres://pg:12345Secure!@localhost:5432/urlshortener?sslmode=disable' -a localhost:8888
```
