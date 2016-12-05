# docker-plugin-hostnic

docker-plugin-hostnic is a docker network plugin which can binding a special host nic to a container.

## QuickStart

1. Make sure you are using Docker 1.9 or later (test with 1.12)
2. Build docker-plugin-hostnic and run, or directly run docker-plugin-hostnic docker image.

    docker pull qingcloud/docker-plugin-hostnic

    docker run -v /run/docker/plugins:/run/docker/plugins -v /etc/docker/hostnic:/etc/docker/hostnic --network host --privileged qingcloud/docker-plugin-hostnic docker-plugin-hostnic

3. Create hostnic network，the subnet and gateway argument should be same as hostnic.

    docker network create -d hostnic --subnet=192.168.1.0/24 --gateway 192.168.1.1 hostnic

4. Run a container and binding a special hostnic. Mac-address argument is for identity the hostnic. Please ensure that the ip argument do not conflict with other hostnic.

    docker run -it --ip 192.168.1.5 --mac-address 52:54:0e:e5:00:f7 --network hostnic ubuntu:14.04 bash


## Additional Notes:

1. If the ip argument is not passed when running container, docker will assign a ip to the container, so please pass the ip  argument and ensure that the ip do not conflict with other hostnic.
2. Network config will save to /etc/docker/hostnic/config.json，if plugin container removed and create again, network config can recover from the config.
3. If your host only have one nic, please not use this plugin. If you binding the only one nic to container, your host will lost network.
