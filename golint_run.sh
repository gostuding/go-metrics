echo "Run golangci"
~/go/bin/golangci-lint run -c $PWD/golangci-lint/.golangci.yml > ./golangci-lint/report-unformatted.json
cd ./golangci-lint
echo "Convert"
python3 ./recomp.py
rm ./report-unformatted.json
echo "Done"