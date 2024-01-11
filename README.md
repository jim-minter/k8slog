# k8slog

Jim Minter, 11 January 2024

```sh
go install github.com/jim-minter/k8slog@latest
k8slog -n namespace [-c container_regexp] [-s source_regexp] [-since duration] | less -S
```

Tabulate JSON container logs.  Logs can be filtered by container regexp, "source" JSON tag regexp and recency.  The `-S` option of `less` can be used to browse the results.

If present, the following fields are moved to lead the tabulated columns:
- timestamp
- level
- filename
- linenumber
- msg
