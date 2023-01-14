# mk 🐧
| make files and make it easy

mk creates directories for all paths ending on slash (`/`) and normal files otherwise.
Make sure to place the file in a directory in your `PATH` and make it executable with `chmod +x mk`

Usage:
       mk <path> [<path>...]

Examples
```
# create a file
mk README.md

# create a directory
mk src/

# create a file and directory at the same time
mk src/main.py

# poweruser example
mk pyproject.toml src/main.py src/__init__.py tests/unit/ tests/integ/ 
```