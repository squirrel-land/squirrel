.PHONY : image

image : squirrel-worker
	docker build -t squirrel-worker .

squirrel-worker :
	CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w' github.com/squirrel-land/squirrel/squirrel-worker
