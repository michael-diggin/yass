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

if [[ "$(docker images -q local-server 2> /dev/null)" == "" ]] || [[ "$(docker images -q local-watchtower 2> /dev/null)" == "" ]]; then
  echo "Docker images not found locally"
  local_build=true
fi

# build the docker images
if [ "$local_build" = true ]; then
    echo "Building docker images..."
    docker build -t local-server -f server/Dockerfile .
    docker build -t local-watchtower -f watchtower/Dockerfile .
fi

# firstly create the docker network if it doesn't exist
if [[ "$(docker network inspect yass-net >/dev/null 2>&1)" ]]; then
    echo "Creating docker network \`yass-net\`..."
    docker network create yass-net
fi

# watchtower needs to be deployed first
echo "deploying watchtower..."
docker run -d --name watchtower -p 8010:8010 --network yass-net local-watchtower -f "node_data.txt" >/dev/null


# next the 3 server nodes
for i in {0..2};
do
echo "deploying server-$i..." 
docker run -d --name server-$i --env POD_NAME=server-$i -p 808$i:808$i --network yass-net local-server -g "watchtower:8010" -p 808$i >/dev/null;
done

echo -e "\nYass Servers accessible on:"
for i in {0..2};
do echo "$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' server-$i):808$i";
done