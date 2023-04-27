
import os
import csv
import collections
from pathlib import Path

# Read input file
with open("FIMFILEA.OUT", "r") as file:
    reader = csv.reader(file, delimiter='\t')
    data = [row for row in reader]

# Initialize directory counter
dir_counter = collections.Counter()

# Analyze the file paths and count changes
for row in data:
    date_code, file_path, sha256, has_changed = row
    if has_changed.lower() == "true":
        dir_path = os.path.dirname(file_path)
        dir_counter[dir_path] += 1

# Calculate the total number of changes
total_changes = sum(dir_counter.values())

# Calculate the percentage of changes for each directory
dir_percentages = {}
for dir_path, change_count in dir_counter.items():
    dir_percentages[dir_path] = (change_count / total_changes) * 100

# Sort directories by the percentage of changes
sorted_dirs = sorted(dir_percentages.items(), key=lambda x: x[1], reverse=True)

# Write the analysis to the output file
with open("analysis.txt", "w") as output:
    writer = csv.writer(output, delimiter='\t')
    for dir_path, percentage in sorted_dirs:
        writer.writerow([dir_path, f"{percentage:.2f}%"])
