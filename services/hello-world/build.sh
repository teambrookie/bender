CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o hello-world .
docker build -t hello-world .