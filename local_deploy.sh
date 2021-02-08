#! /bin/bash
# Simple utility script to spin up the watchtower and 3 server nodes on docker

local_build=false

while [[ "$#" -gt 0 ]]; do
    case $1 in
        -b|--build) local_build=true; shift ;;
        *) echo "Unknown parameter passed: $1"; exit 1 ;;
    esac
    shift
done

# build the docker images
if [ "$local_build" = true ]; then
    docker build -t local-server -f server/Dockerfile .
    docker build -t local-watchtower -f watchtower/Dockerfile .
fi

if [[ "$(docker images -q local-server 2> /dev/null)" == "" ]]; then
  echo "Docker images not found locally, please run \`local_deploy.sh -b\`"
  exit 1
fi

if [[ "$(docker images -q local-watchtower 2> /dev/null)" == "" ]]; then
  echo "Docker images not found locally, please run `local_deploy.sh -b`"
  exit 1
fi

# firstly create the docker network
docker network create yass-net

# watchtower needs to be deployed first
docker run -d --name watchtower -p 8010:8010 --network yass-net local-watchtower -f "node_data.txt"

# next the 3 server nodes
docker run -d --name server-0 --env POD_NAME=server-0 -p 8080:8080 --network yass-net local-server -g "watchtower:8010" -p 8080
docker run -d --name server-1 --env POD_NAME=server-1 -p 8081:8081 --network yass-net local-server -g "watchtower:8010" -p 8081
docker run -d --name server-2 --env POD_NAME=server-2 -p 8082:8082 --network yass-net local-server -g "watchtower:8010" -p 8082

echo -e "\nAccessible on:"
for i in {0..2};
do echo "$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' server-$i):808$i";
done