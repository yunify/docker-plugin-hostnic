# docker-plugin-hostnic

## Usage

    docker pull qingcloud/docker-plugin-hostnic
    docker run -v /run/docker/plugins:/run/docker/plugins --network host --privileged qingcloud/docker-plugin-hostnic docker-plugin-hostnic
    docker network create -d hostnic --subnet=192.168.1.0/24 --gateway 192.168.1.1 hostnic
    docker run -it --ip 192.168.1.5 --network hostnic ubuntu bash


