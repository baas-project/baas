# Project Structure
```
/control_server   # Control server that schedules machines
/docs             # Documentation for the project
/management_os    # The OS+programs that is ran when (re)provisioning a machine
    └─ programs   # The program that is run inside of the management OS
    └─ build      # Scripts for building and packaging the management OS
        └─ linux  # The kernel that is built for the mangement OS (submodule)
/pkg              # Common go code that is shared between all components of this project

```
