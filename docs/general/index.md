# General Project Information
In this section the general structures and implementation of the BAAS project is explained. For example, the way that images are generated and how they are structured.

```plantuml
[Control Server] -right- UserAPI : Offers
[Control Server] -down- ImageAPI : Offers
[Control Server] -up- MachineAPI : Offers
[Control Server] -down- BootAPI : Offers


[Management OS] ..> BootAPI : Consumes
[Management OS] ..> ImageAPI : Consumes
[Web Server] ..> UserAPI : Consumes
[Web Server] ..> ImageAPI : Consumes
[Web Server] ..> BootAPI : Consumes
```

## Documentation Index
1. [Image Layout](images.md)
2. [Architecture](architecture.md)
3. [Project Structure](project_structure.md)
