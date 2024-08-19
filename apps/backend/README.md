# Backend (WIP)

this is the new backend writen in go using
[Spotify-Wrapper](https://github.com/zmb3/spotify) and
[OAuth2](https://github.com/golang/oauth2)

config.yml

```yaml
client_id: "" # your spotify client_id
base_url: "" # http://localhost:3000 https://example.com
port: 8888
```

docker-compose.yaml

```yaml
services:
  backend:
    image: ghcr.io/momozahara/spotify-api-project-backend:latest
    volumes:
      - /path/to/dir:/app/
```
