# Velox

Velox is a somewhat serious programming language based off of C which implements easy-to-use features and fast run times.

## Compiling

The compiler is currently being written in Go, however once Velox is stable enough I plan to rewrite the compiler in Velox.

To compile, simply run `go run src/main.go /path/to/your/file.vl`. If you want to take a look at the build artifacts, pass `--preserve` into the command!

## Features

This section displays working, up to date features of the language!

- Global Constants
`#define CONSTANT 123`

- Basic functions
```c
int mul(int a, int b) {
    return a * b;
}
```

- Types
    - `int, float`
    - More will be added soon

- Conditionals
```c
if (x > y) {
    ...
} else if (x === y) {
    ...
} else {
    ...
}
```

- Loops
    - Only while loops will be supported. This is to simplify how logic works in Velox.
```c
while (x < y) {
    x ++;
}
```

- I/O
    - You may pass an int or a float into this function, no format specifier is supported yet. This is purely for debugging at this point in time.
```c
printf(x);
```

## Known Issues
- Comparison operators on float data type fall over