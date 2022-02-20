# Project Structure
```
control_server     # Control server that schedules machines
    ├─ api         # Code which defines the REST interface
    ├─ disks       # Storage of disk images
    ├─ pixieserver # Code to run a PXE server
    └─ static      # Miscellaneous program data like an initramfs image or a kernel

/docs              # Documentation for the project
/management_os     # The OS+programs that is ran when (re)provisioning a machine
    ├─ programs    # The program that is run inside of the management OS]
    ├─ config      # Configuration files for the program
    └─ build       # Scripts for building and packaging the management OS

/pkg               # Common go code that is shared between all components of this project
    ├─ api         # Structures which are shared across the network
    ├─ compression # Interfaces to various compression algorithms
    ├─ database    # Database interface and concrete implementation
	├── Sqlite     # Database implementation for sqlite.
    ├─ fs          # Functions that manipulate the filesystem
    ├─ httplog     # API to accept log messages from the management server
    ├─ model       # Database models
    └─ util        # Miscellaneous functions and structures.
```
