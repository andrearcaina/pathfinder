# CLI Documentation

The list of commands and flags you can pass into the Pathfinder CLI. You can also always run `pathfinder -h` or `pathfinder --help` for more information.

## Commands
- `pathfinder version`: Displays the current version of Pathfinder.
- `pathfinder scan`: Scans the codebase depending on the provided flags.

## Flags for `pathfinder scan`
- `-b <int>` or `--buffer-size <int>`: Sets the buffer size for reading files in KB. Default is 4.
- `-d` or `--dependencies`: Scans for dependencies in the codebase. Default is false.
- `-f <string>` or `--format <string>`: Output format. Options; JSON
- `-g` or `--git`: Scan for git information. Default is false.
- `-h` or `--help`: Displays help information about the commands and flags.
- `-i` or `--hidden`: Includes hidden files in the scan. Default is false.
- `-m <int>` or `--max-depth <int>`: Sets the maximum directory depth to scan. Default is -1 (which means unlimited).
- `-o <string>` or `--output <string>`: Specifies the output file name
- `-p <string>` or `--path <string>`: Specifies the path to scan. Default is the current directory.
- `-R` or `--recursive`: Enables recursive scanning of directories. Default is false.
- `-t` or `--throughput`: Enables throughput mode to see scanning speed for each worker. Default is false.
- `-w <int>` or `--workers <int>`: Sets the number of concurrent workers 
for scanning. Default is 16.
