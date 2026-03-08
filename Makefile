

git:
	git add .
	git commit -m "$(m)"
	git push

test:
	find . -type f ! -path "*/testdata/*" \( -name "*.mp4" -o -name "*.webm" -o -name "*.mkv" \) -delete
	mkdir -p tests/cut/output tests/concatenate/output tests/speed/output tests/multi_video/output tests/filters/output tests/composite/output tests/stack/output tests/text/output
	go test -v -count=1 -p 1 ./tests/...