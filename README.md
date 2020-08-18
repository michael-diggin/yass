# yass (Yet Another Storage Solution)
A distributed key-value storage solution written in golang using gRPC.

This is still very much a work in progress...

### Roadmap

1. Currently it works locally with an in memory cache (map of strings)

2. Aim is to improve the project, with persistent data storage.  
The idea would be that this repo is cloned, the backend deployed as a service (eg k8s) and the client side is set up as a package to connect to it.

