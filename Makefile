all: main

templ:
	go generate

tailwind.css: templ
	npx tailwindcss@latest -i ./static/global.css -o ./static/tailwind.css --minify

main: templ tailwind.css
	go build -o main main.go
