a=0
while [ $a -lt 10 ]
do 
	echo "Add article article$1-$a to feed $1"
	python3 super_user.py -i 15337 -p $1,"article$1-$a"
	sleep 5
	a=`expr $a + 1`
done
