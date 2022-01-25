#kill $(lsof -t -i:8090)
#kubectl port-forward svc/server-a-service -n athena 8090:8090 &
#sleep 10
#ab -n 10 -c 10 -s 5000 -m GET -g bench http://localhost:8090/a
# GET /api/cart/add/anonymous-6/C3P0/1                               2     0(0.00%)      47      14      81  |      14    0.00
 #GET /api/cart/add/anonymous-7/HAL-1/1                              2     0(0.00%)      12      12      13  |      12    0.00
 #GET /api/cart/cart/anonymous-6                                     1     0(0.00%)      45      45      45  |      45    0.00
 #GET /api/cart/cart/anonymous-7                                     1     0(0.00%)      77      77      77  |      77    0.00
 #GET /api/cart/update/anonymous-6/C3P0/2                            1     0(0.00%)       7       7       7  |       7    0.00
 #GET /api/cart/update/anonymous-7/HAL-1/2                           1     0(0.00%)       7       7       7  |       7    0.00
 #GET /api/catalogue/categories                                      2     0(0.00%)       9       9      10  |       9    0.00
 #GET /api/catalogue/product/C3P0                                    2     0(0.00%)       7       7       8  |       7    0.00
 #GET /api/catalogue/product/HAL-1                                   2     0(0.00%)      38       9      67  |       9    0.00
 #GET /api/catalogue/products                                        2     0(0.00%)      36       9      63  |       9    0.00
 #GET /api/ratings/api/fetch/C3P0                                    0   2(100.00%)       0       0       0  |       0    0.00
 #GET /api/ratings/api/fetch/HAL-1 


ab -n 1 -c 1 -s 5000 -m GET -g bench http://localhost:8080/api/cart/add/anonymous-6/C3P0/1


