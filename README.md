# Monkey Language

Implementation of Monkey Programming Language from [Thorsten Ball](https://thorstenball.com/)'s [Writing An Interpreter In Go](https://interpreterbook.com/) and [Writing A Compiler In Go](https://compilerbook.com/) Books.

## Performance Results 

### Hardware Overview:

```md
Model Name:	MacBook Air
Model Identifier:	MacBookAir10,1
Chip:	Apple M1
Total Number of Cores:	8 (4 performance and 4 efficiency)
Memory:	8 GB
```

### Results
```
➜  monkey git:(main) ✗ go build -o bench benchmark/main.go 
➜  monkey git:(main) ✗ ./bench -engine=eval 
engine=eval, result=9227465, duration=12.911341834s
➜  monkey git:(main) ✗ ./bench -engine=vm  
engine=vm, result=9227465, duration=4.276731125s
```

