go build -o main
docker build . -t broker
sudo service postgresql start
docker-compose up
