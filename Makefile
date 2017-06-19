build:
	GOOS=linux CGO_ENABLED=0 go build -o sweeptaken
	docker build -t opiuman/sweeptaken .
	rm -f sweeptaken