#!/bin/bash
# if nothing prints, the test passed

diff="diff"

go run suti.go -cfg ../examples/suti.cfg -r ../examples/template/html.hmpl > out.html
$diff out.html ../examples/out.html
rm out.html

go run suti.go -cfg ../examples/suti.cfg -r ../examples/template/txt.mst > out.txt
$diff out.txt ../examples/out.txt
rm out.txt
