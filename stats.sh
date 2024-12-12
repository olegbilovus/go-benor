#!/bin/bash

array_N=(10 15 25 50 75 100 125 150 175 200 250 500 700 900 1000)

array_S=(1 5 10 15 20 25 30 35 40 45 50 55 60 65 70 75 80 85 90 95 100)

>&2 echo "n,f,fCount,S,maxS,decision,countViEQ0,countViEQ1"

total=$(expr "${#array_N[@]}" \* "${#array_S[@]}" \* 3 \* 10 )

>&2 echo "total: $total"

for n in "${array_N[@]}"
do

  for S in "${array_S[@]}"
  do

    array_F=(0 $(expr $n / 4) $(expr $n / 2 - 1) )
    for f in "${array_F[@]}"
    do

      for i in {1..10}
      do

        start=$(date +%s.%N)

        ./go-benor -n "$n" -f "$f" -S "$S" --csv --odds 1.0 >> results.csv
        if [ $? -ne 0 ]; then
                    echo "Error"
                    exit 1
                fi

        dur=$(echo "$(date +%s.%N) - $start" | bc)
        >&2 echo "n: $n, S: $S, f: $f, i: $i ($dur s)"

      done

    done

  done

done
