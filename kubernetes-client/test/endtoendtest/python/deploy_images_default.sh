kubectl apply -f test/endtoendtest/python/kubeconfig/initconf.yaml
kubectl label namespace athena istio-injection=enabled
kubectl apply -f test/endtoendtest/python/app1/server_a_d.yaml
kubectl apply -f test/endtoendtest/python/app2/server_b_d.yaml
kubectl apply -f test/endtoendtest/python/app3/server_c_d.yaml
kubectl apply -f test/endtoendtest/python/app4/server_d_d.yaml 
kubectl apply -f test/endtoendtest/python/app5/server_e_d.yaml  
