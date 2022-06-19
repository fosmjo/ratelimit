# URL Shortening

## API Design
**URL Shortening**
```sh
POST /api/v1/data/shorten
```
- request:
  - longUrl: string
- response:
  - shortUrl: string


**URL Redirecting**
```sh
GET /api/v1/:shortURL
```
- response:
  - status: 301
  - location header: longUrl