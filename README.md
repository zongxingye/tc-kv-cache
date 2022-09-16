# tc-kv-cache
## docker本地调试
1. 执行dockerfile的编译命令：sudo docker build -t tckvcache .

2. 建立network ，执行命令 
docker volume create tckvcache1
docker volume create tckvcache2 
docker volume create tckvcache3
3. 运行三个容器
docker run -it  --name kv11 --network kvcache -d -v tckvcache1:/app/test  tckvcache
docker run -it  --name kv12 --network kvcache -d -v tckvcache2:/app/test  tckvcache
docker run -it  --name kv13 --network kvcache -d -v tckvcache3:/app/test  tckvcache
4. 进入docker bash并调用updatecluster接口（无需顺序，需要注意自己的ip）：
   curl --location --request POST 'http://172.21.0.4:8080/updateCluster' \
   --header 'Content-Type: application/json' \
   --data-raw '{
   "hosts": [
   "172.21.0.2",
   "172.21.0.3",
   "172.21.0.4"
   ],
   "index": 3
   }'
   curl --location --request POST 'http://172.21.0.3:8080/updateCluster' \
   --header 'Content-Type: application/json' \
   --data-raw '{
   "hosts": [
   "172.21.0.2",
   "172.21.0.3",
   "172.21.0.4"
   ],
   "index": 2
   }'
   curl --location --request POST 'http://172.21.0.2:8080/updateCluster' \
   --header 'Content-Type: application/json' \
   --data-raw '{
   "hosts": [
   "172.21.0.2",
   "172.21.0.3",
   "172.21.0.4"
   ],
   "index": 1
   }'
5. 目前没有init接口，只能调用其他的接口