# fim
File integrity monitor written in Golang and supporting multi-threaded scanning.

v.2

Usage: fim [-v] [-vv] [-h]
-v: Verbose mode (errors only)
-vv: Super verbose mode (all files processed)
-h: Help

Example:
sudo fim -vv
sudo fim -v
sudo fim

This is a multi-threaded application that performs a file integrity 
check using sha256 hashing and comparing since last scan. 

Hash changes will be recorded in FIMFILEA.OUT as TRUE after second 
scan if hashes do not match. FALSE for files that have not changed 
since last scan. Hashes OS regular files only. 

You can exclude directories or files using the absolute file path 
in the exclude.config file. Directories listed will exclude all 
subdirectories. Will not follow symlinks.

This code supports multi-platform compiling.
