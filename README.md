# Gimmie

Do you forget your tar flags? Tired of struggling to decompress your tar/gz/bz2 files? Struggle no more!
Gimmie determines the compression format and extracts your file for you.

## Example
```
~/go/src/gimmie $ ./gimmie test.tar.bz2
[+] Found bz2 compression
[+] Found a tar archive

~/go/src/gimmie $ ls test -lh
total 4.0K
-rw-rw-r-- 1 v33ps v33ps 4 Mar  2 01:28 test.txt
~/go/src/gimmie $ cat test/test.txt
hi
```
## Build Instructions
```
go build -o gimmie
```

## TODO
* finish tests and add them
* clean up code and make things shiny
