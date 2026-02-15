git:
	git add .
	git commit -m "$(m)"
	git push

test:
	go test -v ./tests/...