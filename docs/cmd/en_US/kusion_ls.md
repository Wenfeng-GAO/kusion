## kusion ls

List all Projects and Stacks

### Synopsis

List all project and stack information in the current directory or the specify directory. The default output is in a human friendly format, and it also supports a variety of formatted structure output.

```
kusion ls [WORKDIR] [flags]
```

### Examples

```
  # List all project and stack information in the current directory
  kusion ls
  
  # List all project and stack information in the specify directory
  kusion ls ./path/to/project_dir
  
  # List all project and stack information in the specify directory,
  # and output in a Tree format
  kusion ls ./path/to/project_dir --format=tree
  
  # List all project and stack information in the specify directory,
  # and output in a JSON format
  kusion ls ./path/to/project_dir --format=json
  
  # List all project and stack information in the specify directory,
  # and output in a YAML format
  kusion ls ./path/to/project_dir --format=yaml
  
  # List all project and stack by level, and output in a Tree format
  kusion ls ./path/to/project_dir --format=tree --level=1
```

### Options

```
      --format string   Output format of the project information. valid values: json, yaml, tree, human (default "human")
  -h, --help            help for ls
  -L, --level int       Max display depth of the project and stack tree. One of 0,1,2 (default 2)
```

### SEE ALSO

* [kusion](kusion.md)	 - Kusion is the platform engineering engine of KusionStack

###### Auto generated by spf13/cobra on 28-Sep-2023
