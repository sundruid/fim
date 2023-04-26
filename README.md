# fim
File integrity monitor written in Golang and supporting multi-threaded scanning.

Usage: fim [-v] [-h]

-v: Verbose mode
-h: Help

This is a multi-threaded application that performs a file integrity 
check using sha256 hashing and comparing since last scan. 

Hash changes will be recorded in FIMFILEA.OUT as TRUE after second 
scan if hashes do not match. FALSE for files that have not changed 
since last scan. 

You can exclude directories or files using the absolute file path 
in the exclude.config file. Directories listed will exclude all 
subdirectories. Will not follow symlinks.

This code supports multi-platform compiling.
