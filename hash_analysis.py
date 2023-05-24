#!/usr/bin/env python3

import os
import collections
from pathlib import Path

# Initialize directory counter
dir_counter = collections.Counter()

# Read input file
with open("FIMFILEA.OUT", "r") as file:
    lines = file.readlines()
    print(f"Total lines in the file: {len(lines)}")  # Debugging print

    for line in lines:
        parts = line.strip().split("\t")
        if len(parts) != 4:
            print(f"Skipping line due to incorrect number of fields: {line}")  # Debugging print
            continue

        date_time, file_path, sha256, has_changed = parts

        if has_changed == "TRUE":
            dir_path = os.path.dirname(file_path)
            dir_counter[dir_path] += 1

# Calculate the total number of changes
total_changes = sum(dir_counter.values())
print(f"Total changes found: {total_changes}")  # Debugging print

# Calculate the percentage of changes for each directory
dir_percentages = {}
for dir_path, change_count in dir_counter.items():
    dir_percentages[dir_path] = (change_count / total_changes) * 100

# Sort directories by the percentage of changes
sorted_dirs = sorted(dir_percentages.items(), key=lambda x: x[1], reverse=True)

# Print the analysis to stdout
for dir_path, percentage in sorted_dirs:
    print(f"{dir_path}\t{percentage:.2f}%")

# Print the number of directories analysed
print(f"Number of directories analysed: {len(sorted_dirs)}")  # Debugging print
