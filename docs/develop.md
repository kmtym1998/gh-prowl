# develop

```sh
# install dependencies
go mod tidy

# build extension binary
go build

# install to your extension. Run this in the same directory where you executed 'go build'.
gh extension install .

# execute binary
gh prowl
```
