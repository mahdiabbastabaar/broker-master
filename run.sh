go build -o main
docker build . -t broker
docker-compose up
