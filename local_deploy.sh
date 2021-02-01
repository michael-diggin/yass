#! /bin/bash
# Simple utility script to spin up the watchtower and 3 server nodes on docker

# build the docker images
docker build -t yass-server:0.2 -f server/Dockerfile .
docker build -t yass-watchtower:0.1 -f watchtower/Dockerfile .

# firstly create the docker network
docker network create yass-net

# watchtower needs to be deployed first
docker run -d --name watchtower -p 8010:8010 --network yass-net yass-watchtower:0.1 -f "node_data.txt"

# next the 3 server nodes
docker run -d --name server-0 --env POD_NAME=server-0 -p 8080:8080 --network yass-net yass-server:0.2 -g "watchtower:8010" -p 8080
docker run -d --name server-1 --env POD_NAME=server-1 -p 8081:8081 --network yass-net yass-server:0.2 -g "watchtower:8010" -p 8081
docker run -d --name server-2 --env POD_NAME=server-2 -p 8082:8082 --network yass-net yass-server:0.2 -g "watchtower:8010" -p 8082

echo -e "\nAccessible on:"
for i in {0..2};
do echo "$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' server-$i):808$i";
done