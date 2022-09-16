docker volume create tckvcache1
docker volume create tckvcache2
docker volume create tckvcache3
docker run -it  --name kv11 --network kvcache -d -v tckvcache1:/app/test  tckvcache

docker run -it  --name kv12 --network kvcache -d -v tckvcache2:/app/test  tckvcache
docker run -it  --name kv13 --network kvcache -d -v tckvcache3:/app/test  tckvcache
#mkdir "~/kvcache1" && cd "~/kvcache2"
#docker run -it --name kv12 -v $(pwd):/app --network kvcache -d mykvcachev2
#mkdir "~/kvcache1" && cd "~/kvcache2"
#docker run -it --name kv13 -v $(pwd):/app --network kvcache -d mykvcachev2