import sys

filename = sys.argv[1]

value = []
with open(filename, 'r') as file:
    for line in file:
        columns = line.split()  # Split the line into columns based on whitespace
        value.append(int(columns[-1]))  # Or do whatever you want with the columns
        
value = sorted(value)
print(value[-5:])