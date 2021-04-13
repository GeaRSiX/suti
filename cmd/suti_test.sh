#!/bin/bash
# if nothing prints, the test passed

diff="diff -bs"

go build -o suti suti.go

if [ -e suti ]; then
	./suti -cfg ../examples/suti.cfg -r ../examples/template/html.hmpl > out.html
	$diff out.html ../examples/out.html
	rm out.html

	./suti -cfg ../examples/suti.cfg -r ../examples/template/txt.mst > out.txt
	$diff out.txt ../examples/out.txt
	rm out.txt

	rm suti

	echo "if files are identical, TEST PASS"
fi
