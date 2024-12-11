go-benor simulates the BenOr algorithm.

## Usage

After installing GO on your machine, you can just clone this repository and run `go run .` inside
the repository.

By default, the initial values will be generated randomly, but you can use the `-v` option to write
your desired initial values.

For more details on the available options, read below or type `go run . --help`.

```
Usage of go-benor:
  -S int
        number of phases (default 10)
  -f int
        max number of stops (default 1)
  -n int
        number of processes (default 3)
  -v string
        initial values of the processes. Example: 1 0 1 1
  -verbose
        print all the messages sent and received in real time

```

## Stops

go-benor will randomly stop the processes up to `f` processes. The stop can happen at any time
before starting a new phase.

## Termination probability

There is a [geogebra file](terminationProbability.ggb) to view the relation between `n` and `S`. It
can tell how big the `S` should be to get a good probability of termination.

## Output

By default, go-benor will show a progress bar of the computation. The total number of computations
is given by $n * S$.

After the computation is done, it will print in order:

1. initial values
2. decisions: 
    - 1
    - 0
    - -1 if a process did not decide
    - "stopped" if the process stopped
3. info:
    - the given input
    - number of processes needed for a majority: $\lfloor \frac{n}{2} \rfloor+ 1$
    - termination probability: $1 - (1 - \frac{1}{2^n})^S$
    - fCount: number of stopped processes
    - How many phases were needed to decide.

The option `--verbose` can be used to view every sent and received messages and when a process
starts a new phase. Do not use this when using big `n` or `S` because the console outputs will slow
down everything else. The progress bar will not be visible.

### Output example for `go run .`
```
 100% |██████████████████████████████████████████████████████| (30/30, 58173 it/s)        
----- INIT VALUES -----
v_0: 0
v_1: 1
v_2: 1
----- DECISIONS -----
P_0 decided: 0
P_1 decided: 0
P_2 stopped
----- INFO -----
n: 3, f: 1, S: 10, majority: 2, termProb:73.69%, fCount: 1
Decided after 10/10 (100.00%) phases.
```


## Resources usage

For a big `n` or `S`, >= 10000, go-benor will start using a lot of memory and computation. At some
point, the progress bar may seem to be stuck, but it is the Garbage Collector in GO doing its
things.



