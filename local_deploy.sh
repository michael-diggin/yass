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

if [[ "$(docker images -q local-server 2> /dev/null)" == "" ]]; then
  echo "Docker images not found locally"
  local_build=true
fi

# build the docker images
if [ "$local_build" = true ]; then
    echo "Building docker images..."
    docker build -t local-server -f server/Dockerfile .
fi

# firstly create the docker network if it doesn't exist
if [[ "$(docker network inspect yass-net >/dev/null 2>&1)" ]]; then
    echo "Creating docker network \`yass-net\`..."
    docker network create yass-net
fi

# next the 3 server nodes
for i in {0..2};
do
echo "deploying yass-$i..." 
docker run -d --name yass-$i --env POD_NAME=yass-$i:808$i -p 808$i:808$i --network yass-net local-server -join "yass-0:8080,yass-1:8081,yass-2:8082" -p 808$i >/dev/null;
done

echo -e "\nYass Servers accessible on:"
for i in {0..2};
do echo "$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' yass-$i):808$i";
done