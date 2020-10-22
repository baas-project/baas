# Management OS Reprovision flow
1. Boot into management OS
2. Inform the control_server that we have booted and receive `MachineSetup` info
3. Clean up previous session (save disk state etc.)
4. Set everything up for next session (restore disk state etc.)
