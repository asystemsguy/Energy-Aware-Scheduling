kubectl create namespace athena
kubectl label namespace athena istio-injection=enabled
#kubectl apply -f /home/satish/real_world_applications/robot-shop/K8s/descriptors -n athena
kubectl apply -f test/endtoendtest/python/robo_shop/descriptors/ -n athena
