# Task Division

[Node] <---> [control server]

* Node: high performance machine used for OS research
* Control server: an external machine providing kernel images for users.

# process for booting nodes

1. Node boots and requests management kernel from control server (pxe)
2. Management kernel sets up the system
3. Reboot, make sure the user's kernel is booted. Make sure boot order is changed in some way so the management kernel does not load again. (ipmi?)

# Management kernel's jobs:

1. Check the system integrity
2. Clear disk
3. Repartition based on the user's needs (image definition)
4. Writes the user's kernel images to disk
5. Write the user's persistent storage to disk

# Image:

Generic images + specialized linux-focused utilities

* Disk images (with one partition partition)
* Partition list (with partition metadata)
* Metadata about persistence, architecture (arm, x86), boot information etc,
