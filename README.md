# Libs
Repo with libraries for HMTM usage.

## Linters

To run linters, use next command:
```shell
 task -d scripts linters -v
```

## Tests

To run test, use next commands. Coverage info will be
recorded to ```coverage``` folder:
```shell
task -d scripts tests -v
```

To include integration tests, add `integration` flag:
```shell
task -d scripts tests integration=true -v
```

## Benchmarks

To run benchmarks, use next command:
```shell
task -d scripts bench -v
```
