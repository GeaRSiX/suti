#!/usr/bin/env sh
# if nothing prints, the test passed

diff="diff -bs"
fail=0

go build -o dati dati.go

if [ -e dati ]; then
	./dati -cfg ../examples/dati.cfg -r ../examples/template/html.hmpl > out.html
	$diff out.html ../examples/out.html
	if [ $? -ne 0 ]; then fail=1; else rm out.html; fi

	./dati -cfg ../examples/dati.cfg -r ../examples/template/txt.mst > out.txt
	$diff out.txt ../examples/out.txt
	if [ $? -ne 0 ]; then fail=1; else rm out.txt; fi

	rm dati

	if [ $fail -eq 1 ]; then echo "TEST FAIL"; else echo "TEST PASS"; fi
fi
