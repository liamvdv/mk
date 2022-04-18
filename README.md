# mk
| make files and make it easy

A single command replacing `touch`, `mkdir` and `ln`. \
Just specify a path and mk will take care of creating the missing directories **and** files. 

Usage:
       mk <path> [-s- <target> | -h- <target>] ...

make handles creating files, directories and symbolic links. \
It unifies the unix commands 'touch' 'mkdir' and 'ln'.

Paths ending on / (slash) indicate a directory, else it's a file.
```bash
$ mk main.go  # create file main.go
$ mk docs/	  # create docs folder	
```
Missing directories are created by default. 
```bash
$ mk non/existent/parent/and/file.txt   # create missing dirs and file.txt
```
Create symbolic links with -s- and hard links with -h-. Think of labeled edges.
```bash
$ mk symname -s- target/file   # create symlink:  symname -s-> target/file 
$ mk aliasa -h- target/file    # create hardlink: aliasa -h-> target/file
```