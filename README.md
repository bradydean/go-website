# go-website

Simple todo list website using the GoTTH stack.

## tools

- [Go](https://golang.org/)
- [Tailwind](https://tailwindcss.com/)
- [Templ](https://templ.guide/)
- [HTMX](https://htmx.org/)
- [jet](https://github.com/go-jet/jet)
- [pgx](https://github.com/jackc/pgx)
- [migrate](https://github.com/golang-migrate/migrate)

## setup

1. Populate environment variables, or use a `.env` file.

The Auth0 application must be a Regular Web Application with the following settings:

- Allowed Callback URLs: `http://localhost:8000/callback`
- Allowed Logout URLs: `http://localhost:8000`

```sh
export DATABASE_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable
export AUTH0_DOMAIN=your.auth0.com
export AUTH0_CLIENT_ID=your-auth0-client-id
export AUTH0_CLIENT_SECRET=your-auth0-client-secret
export AUTH0_CALLBACK_URL=http://localhost:8000/callback
export SESSION_KEY=$(openssl rand -base64 32)
export PORT=8000
```

2. Run the following to setup the database.

```sh
go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate -path ./migrations -database ${DATABASE_URL} up
```

3. Run the following to start the server.

```sh
go build -o go-website main.go
./go-website
```

## development

### dependencies

- [Go](https://golang.org/)
- [Node/NPM](https://nodejs.org/)

### shortcuts

1. Generate `tailwind.css` and Templ components, Jet schema builder.

```sh
go generate
```

2. Format files

```sh
go fmt ./...
go run github.com/a-h/templ/cmd/templ fmt ./internal/pkg/components
```
