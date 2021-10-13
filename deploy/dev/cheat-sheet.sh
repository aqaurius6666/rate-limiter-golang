NETWORK_NAME=medicalchain-dev
docker-compose --project-name=medicalchain-dev -f deploy/dev/docker-compose.yaml up -d
docker exec -it medicalchain-dev_authservice_1 sh
docker logs medicalchain-dev_authservice_1 -f