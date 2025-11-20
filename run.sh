source env.sh
go build -v -o build/main
ret=$?
if [ ! $ret -eq 0 ]; then 
  echo -e "Exit code $ret"
  exit $?
fi
./build/main