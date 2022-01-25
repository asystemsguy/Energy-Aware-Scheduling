kill $(lsof -t -i:8090)
kubectl port-forward svc/server-a-service -n athena 8090:8090 &
#rm -rf results.txt
sleep 10
for i in 1 2
do 
echo $i

for j in 1 2 3 4 5 6 7 8 9 10
do 
echo $j
temp=$(ab -n $j -c $j -s 5000 -e test.csv -m GET -g bench http://localhost:8090/a)
t=$(echo $temp | cut -d ' ' -f76)
echo "$j $t" >> results.txt
#sleep 5
done 

for k in 10 9 8 7 6 5 4 3 2 1
do 
echo $k
temp1=$(ab -n $k -c $k -s 5000 -e test.csv -m GET -g bench http://localhost:8090/a)
t1=$(echo $temp1 | cut -d ' ' -f76)
echo "$k $t1" >> results.txt
#sleep 5
done

done
#echo $temp
#sed -n '4p' $temp.txt
#s=$('echo $temp | awk -F: '{print $3}'')
#echo $s
