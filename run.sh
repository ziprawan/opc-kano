go build -ldflags="-w -s" -v -o build/main
ret=$?
if [ ! $ret -eq 0 ]; then 
  echo -e "Exit code $ret"
  exit $ret
fi
./build/main