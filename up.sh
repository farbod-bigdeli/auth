docker build --pull --rm -f "Dockerfile" -t auth:0.0.5 .

docker save auth:0.0.5 > auth.tar.gz

scp *.tar.gz centos@YOUR_SERVER:tmp/

ssh centos@YOUR_SERVER docker load < auth.tar.gz