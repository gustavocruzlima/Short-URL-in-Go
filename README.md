# Short URL in Go

Create a short URL from input and saves on Mongo DB

## Running

Run the command bellow in terminal

```bash
go run shortener.go
```

## inputs

To create a short URL

```json
{
	"url": "test.com"
}
```

in localhost:8080/short

to retrieve a shortened URL 

```json
{
	"short": "shorturl.com/g6NGDvjZ",
}
```

in localhost:8080/search
