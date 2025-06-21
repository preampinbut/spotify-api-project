# Backend (WIP)

this is the new backend writen in go using
[OAuth2](https://github.com/golang/oauth2)

.env

```
BASE_URL= # http://localhost:3000 https://example.com
CLIENT_ID= # your spotify client_id
PORT=8888

DB_SCHEMA= # mongodb+srv mongodb
DB_HOST=
DB_OPTIONS=retryWrites=true&w=majority&appName=Cluster0
DB=
DB_COLLECTION=

DB_URL= # ${DB_SCHEMA}://${DB_HOST}/${DB_OPTIONS}
DB_USERNAME=
DB_PASSWORD=

DATABASE_URL= # your database connection string
DATABASE_URL= # ${DB_SCHEMA}://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOST}/${DB}
```
