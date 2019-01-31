# mnemosine
A stupid tool to import txt leaks and produce SQLite db for faster queries.

It is like a *Prêt-à-porter* dump. Rofl.

## To cross-compile (from MacOS for Windows systems)

```
brew install mingw-w64

export GOOS=windows
export GOARCH=amd64
export CGO_ENABLED=1
export CC=/usr/local/Cellar/mingw-w64/5.0.3_1/bin/x86_64-w64-mingw32-gcc
export CXX=/usr/local/Cellar/mingw-w64/5.0.3_1/bin/x86_64-w64-mingw32-g++
go build
```

## Usage example

To generate a sigle SQLite file from a single text file containing credentials.

```
.\mnemosine.exe -in-file leak-file.txt -out-file leak-file.db -skip-file leak-file-skipped.txt
```

To generate multiple SQLite files from multiple text files containing credentials.

```
./mnemosine -in-dir /your/path/containing/leaks/ -in-ext txt -skip-file skipped.txt
```