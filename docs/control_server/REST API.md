# Control API
For the BAAS project there exists only one singular way to interface with the control server which is making use of the public facing REST API. These define the functions which can be used to manipulate the server, create images and manage data. In this implementation the REST architecture defined by [Roy Fielding's dissertation](https://www.ics.uci.edu/~fielding/pubs/dissertation/rest_arch_style.htm) and explained on [RESTful API.net](https://restfulapi.net/) is used. All messages to and from the server are in the JSON data format.

## Design
Currently there are two related HTTP interfaces: one user facing and one internal for the mangagement OS. In the future, the aim is to have a singular interface which is used for both parts. In this section the focus is on the latter since this is the main interface and by far the most complex.

There are exists six root uri paths where all other resources have been clustered under, they are as follows:

- `/machine(s)` which defines the functions to manipulate the machines that the images are run on. <br>
- `/user(s)` which allows for the management and creation of users. <br>
- `/image(s)` is used to access the created images. <br>
- `/v1/boot` is only used for the iPXE server. <br>
- `/static` are the static images and irrelevant for users. <br>
- `/log` where the debug messages from the management OS are sent to. <br>

Each of these routes define their own unique resources which they manage and for each there is a direct correspondence to the related entity in the database. For each of these you can expect at least the basic CRUD functions with some extra user-friendly functionality. All of them, besides the creating user, requires an authentication token since not all features are available to everyone. You can find a full compendium of the endpoints at the end of this page.

In general of all of the resources look like the following:

- `GET /nouns` returns all instances of this object. <br>
- `POST /noun` creates a new instance of the object. <br>
- `GET /noun/name` fetches the instance with the id name. <br>
- `PUT /noun/name` updates the instance with the id name. <br>
- `METHOD /noun/name/verb` performs the specified action for that instance <br>

Different resources may be nested in groups of two arbitrarily deep into other resources, for example, `/machine/[mac]/disk/[uuid]/file/[name]`.

Some endpoints may require a user to be logging in, as indicated by the permissions field in the documentation below, which means that the `session-name` cookie must be set to the right value. This can be done by simply [logging in](logging_in.md), copying the relevant cookie value and using it in your requests. For example, using cURL you want to prefix your commands with: `--cookie "session-name=[some base64 string]"`.


## Endpoint compendium
In this section an overview is given of every single on the defined endpoints together with an example on how to call it, what parameters it takes and what it returns. This section is divided in the same way as the resources defined above.

### Machines
Here the management of the machines is done, where machine is any computer which the images can be run on. This can be used to register new servers to the system, upload new disks or get information about the actions the machine should take.

#### Get information about a machine
Allows a user to get information about a specific machine which is identified by it's MAC address. Please note that different MAC addresses may refer to the exact same machine.

**Request:** `GET /machine/[mac]` <br>
**Body:** None <br>
**Response:** <br>
- *Name:* Human-readable name. <br>
- *Architecture:* Architecture of the machine <br>
- *MacAddresses*: MAC Addresses which are associated with this machine. <br>
**Permission**: All <br>
**Example curl command**: `curl localhost:8080/machine/00:11:22:33:44:55:66` <br>
**Example response:**
```json
{
    "name": "Machine 1",
    "Architecture": "x86_64",
    "MacAddresses": [{
		 "Mac": "00:11:22:33:44:55:66
	}]
}
```

#### Getting all machines
Receives information about every currently registered machine.

**Request:** `GET /machines` <br>
**Body**: None <br>
**Response:** A list of machine objects described above. <br>
**Permission**: All <br>
**Example curl command:** `curl localhost:8080/machines` <br>
**Example response:**
```json
[{
	"name": "Machine 1",
    "Architecture": "x86_64",
    "MacAddresses": [{
		"Mac": "00:11:22:33:44:55:66"
    }]
  },
  {
	"name": "Machine 1",
    "Architecture": "x86_64",
    "MacAddresses": [{
		"Mac": "42:DE:AD:BE:EF:42"
    }]
}]
```

#### Update machine
Change the information of a machine, this also used to create a machine.

**Request:** `PUT /machine` <br>
**Body:** <br>
- *Name:* Human-readable for the machine. <br>
- *Architecture:* Architecture of the machine, typically x86\_64 <br>
- *Managed:* Unknown <br>
- *MacAddresses:* A list of MAC addresses associated with the system in the form of `{"mac": "value"}"`. <br>
**Response:** None
**Permissions:** Administrators <br>
**Example body:**
```json
{
	"name": "Hello World",
	"Architecture": "x86_64",
	"Managed": true,
	"DiskUUIDs": null,
	"MacAddresses": [{
		"Mac": "52:54:00:d9:71:15",
		"MachineModelID": 12
	}]
}
```
**Example curl command:** `curl -X PUT localhost:4848/machine -H 'Content-Type: application/json' -d '{"name": "Test", "Architecture": "x86_64", "Managed": true, "DiskUUIDs": null, "MacAddress": [{"Mac": "52:54:00:d9:71:12", "MachineModelID": 12}]}'` <br>

#### Get the boot information
Fetches configuration for the next boot for this particular machine using a SQL based FIFO queue.

**Request:** `GET /machine/mac/boot` <br>
**Body:** None <br>
**Response:** <br>
- *Prev* A description of the last boot configuration. <br>
- *Next* Boot configuration that is going to be downloaded to the server. <br>
- *\*.Ephemeral:* Unused <br>
- *\*.Disks:* List of the disks on the machine. <br>
- *\*.Disks.UUID:* Image UUID. <br>
- *\*.Disks.Version:* Version of image that should be downloaded. <br>
- *\*.Disks.Image:* Nested object describing where the image is flashed to. <br>
- *\*.Disks.Image.DiskType:* Format of the image file. <br>
- *\*.Disks.Image.DiskTransferStrategy:* Strategy for the downloading the images. <br>
- *\*.Disks.Image.Location:* Disk UUID or disk file on a GNU/Linux machine <br>
**Permissions:** Management OS <br>
**Example curl request**:` curl localhost:4848/machine/42:DE:AD:BE:EF:42/boot` <br>
**Example response:**
```json
{
   "Prev":{
      "Ephemeral":false,
      "Disks":[
         {
            "UUID":"0d0a60c4-7718-4d2e-9aef-3abbfd4209ea",
            "Version":5,
            "Image":{
               "DiskType":0,
               "DiskTransferStrategy":1,
               "Location":"/dev/sda"
            }
         }
      ]
   },
   "Next":{
      "Ephemeral":false,
      "Disks":[
         {
            "UUID":"4563e2a1-c77a-4427-a47a-acf7695f6329",
            "Version":2,
            "Image":{
               "DiskType":0,
               "DiskTransferStrategy":0,
               "Location":"/dev/sdk2"
            }
         }
      ]
   }
}
```

#### Add another configuration to a machine queue
Push a boot configuration to the queue in a machine's FIFO boot queue. In this future this should probably be machine agnostic.

**Request:** `POST /machine/[mac]/boot` <br>
**Body:** <br>
- *Version:* The version of the image that you want to boot <br>
- *ImageUUID:* UUID associated with the image <br>
- *Update:* Should the changes to the image by synced? <br>
**Response:** <br>
- *MachineModelID:* Machine that the image should be flashed to. <br>
- *Version:* Image iteration that is downloaded to the machine. <br>
- *ImageUUID:* UUID for the image. <br>
- *Update:* Should the changes be synced to the disk. <br>
**Permissions:** All <br>
**Example curl request:** `curl -X POST "localhost:4848/machine/52:54:00:d9:71:93/boot" -H 'Content-Type: application/json' -d '{"Version": 1636116090, "ImageUUID": "74368cec-7903-4233-87b7-564195619dce", "update": true}' `<br>
**Example response:**
```json
{
  "MachineModelID": 1,
  "Version": 1636116090,
  "ImageUUID": "74368cec-7903-4233-87b7-564195619dce",
  "Update": true
}
```

### Users
Users are the access control mechanism which is used in the BAAS project. There are exists three kinds of users: administrators, moderators and users. Users can only access system images and modify their own images. Moderators can modify assigned system images. Administrators can modify any part of the program.

#### Create a new user
Add user to the system.

**Request:** `POST /user` <br>
**Body:** <br>
- *name:* Requested username <br>
- *email:* Email of the user <br>
- *role:* One of user, moderator or administrator <br>
**Response:** Successfully created user <br>
**Permissions:** Administrators/System <br>
**Example curl request:** `curl -X POST "localhost:4848/user -h 'Content-Type: application/json' -d {"name": "William Narchi", "email": "w.narchi1@student.tudelft.net", "role": "user"}`


#### Login using GitHub
Starts the OAuth2 process as described in [logging in](logging_in.md)

**Request:** `GET /user/login/github` <br>
**Body:** None <br>
**Response:** A link to Github's Authentication page <br>
**Example request:** `curl "localhost:4848/user/login/github"`

#### Fetch a particular user
Returns information about a particular user.

**Request:** `GET /user/[name]` <br>
**Body:** None <br>
**Response:** <br>
- *Name:* Username of the user.  <br>
- *Email:* Associated email adres of the university. <br>
- *Role:* Assigned permissions <br>
**Permissions:** All <br>
**Example curl request:** `curl "localhost:4848/user/Jan"` <br>
**Response:**
```json
{
   "Name": "Jan",
   "Email:" "j.w.dijkstra@tudelft.nl",
   "Role:" "admin"
}
```

#### Fetch the currently logged in user
Finds the user information for the user that is currently logged into the session.

**Request:** `GET /user/me` <br>
**Body:** None <br>
**Response:** Same as [fetch a particular user](#fetch-a-particular-user) <br>
**Example request:** `curl "localhost:4848/user/me" --cookie "session-name=[value]"`


#### Get all registered users
Gives a list of every user which is currently registered with the system.

**Request:** `GET /users` <br>
**Body:**  None  <br>
**Response:** A list of user objects described above. <br>
**Permissions:** All <br>
**Example curl request:** `curl "localhost:4848/users" ` <br>
**Example response:**
```json
[
  {
    "Name": "Jan",
    "Email": "j.w.dijkstra@tudelft.nl",
    "Role": "admin",
    "Images": null
  },
  {
    "Name": "William",
    "Email": "j.cena@student.tudelft.nl",
    "Role": "user",
    "Images": null
  }
]
```

#### Create a new image
Creates a new image entity and file.

**Request:** `POST /user/[name]/image` <br>
**Body:** <br>
- *Disk UUID:* Either a [disk UUID](https://wiki.archlinux.org/title/persistent_block_device_naming#by-uuid) or a device block name. <br>
- *Name*: A human-readable name for this image.
**Response:** <br>
- *Name:* Human-readable name of the image.
- *Versions:* A list of objects with a Version attribute containing the version number.
- *UUID:* Identifiying unique ID for the image.
- *DiskUUID:* Where the image should be flashed to.
- *UserModelID:* Identifies the owner of the image.
**Permissions:** User in question or administrator <br>
**Example curl request:** `curl -x POST "localhost:4848/user/[name]/image" -h 'Content-Type: application/json' -d '{"name": "Fedora", "DiskUUID": "30DF-844C"}'` <br>
**Response:**
```json
{
   "Name": "Fedora",
   "Versions": [{
	 "Version": "5",
	 "ImageModelID": 0
	}],
	"UUID": "eed13670-5974-4c98-b044-347e1f630bc5",
	"DiskUUID": "30DF-844C",
	"UserModelID": 0
}
```

#### Find all the images made by a user
Returns every image created by the human without versions.

**Request:** GET /user/[name]/images <br>
**Body:** None <br>
**Response:** A list of image objcts described above <br>
**Permissions:** User in question or administrator <br>
**Example curl request:** `curl "localhost:4848/user/Jan/images"` <br>
**Example response:**
```json
[
  {
    "Name": "Arch Linux",
    "Versions": [
      {
        "Version": 1635860188,
        "ImageModelID": 1
      },
      {
        "Version": 1635888918,
        "ImageModelID": 1
      },
      {
        "Version": 1635890557,
        "ImageModelID": 1
      },
      {
        "Version": 1635890643,
        "ImageModelID": 1
      }
    ],
    "UUID": "87f58936-9540-4dad-aba6-253f06142166",
    "DiskUUID": "30DF-844C",
    "UserModelID": 2
  },
  {
    "Name": "Arch Linux",
    "Versions": [
      {
        "Version": 1636115517,
        "ImageModelID": 2
      }
    ],
    "UUID": "5c62c7c9-95a2-4b19-9df5-1a74320373bd",
    "DiskUUID": "30DF-844C",
    "UserModelID": 2
  }
]

```

#### Get all the images from a user with a particular name
Find all the images which are associated with user which share the same human-readable name. This is a convenience function which can be used for searching through the images since, at the moment, image names are not unique.

**Request:** GET /user/[username]/images/[name] <br>
**Body:** None <br>
**Response:** A list of image objects described above filtered on name. <br>
**Permissions:** User in question or administrator <br>
**Example curl request:** `curl "localhost:4848/user/Jan/images/Gentoo"` <br>
**Example response:** <br>
```json
[
  {
    "Name": "Gentoo",
    "Versions": null,
    "UUID": "57bf0cd3-c2bf-4257-acdd-b7f1c8633fcf",
    "DiskUUID": "30DF-844C",
    "odelID": 1
  }
]
```


### Images
Represents the images used for the BAAS project. Endpoints can be found in the `/image/` resource pool, but also as a part of the `/user` pool. In particular, the creation of user system images are typically in the latter rather than the former.

#### Get image info
Offers the underlying image file to the user.

**Request:** `GET /image/[UUID]` <br>
**Body:** None <br>
**Permissions:** User in question or administrator <br>
**Example curl request:** `curl "localhost:4848/image/42:DE:AD:BE:EF:42"` <br>
**Response:** <br>
- *Name:* Human-readable name associated with the image. <br>
- *Versions*: A list of versions available for this image. Versions are JSON objects with a Version attribute. <br>
- *UUID:* UUID of the image itself. <br>
- *DiskUUID:* UUID for the disk partition on the target machine that it is flashed to. <br>
- *UserModelID:* ID of the user who created the image. <br>
**Example response:**
```json
{
  "Name": "Arch Linux",
  "Versions": [
    {
      "Version": 1636116090,
      "ImageModelID": 3
    },
    {
      "Version": 1636585720,
      "ImageModelID": 3
    }
  ],
  "UUID": "74368cec-7903-4233-87b7-564195619dce",
  "DiskUUID": "30DF-844C",
  "UserModelID": 2
}
```

#### Get the latest version of an image
Convenience function which offers the latest version of the specified image.

**Request:** `GET /image/[name]/latest` <br>
**Body:** None. <br>
**Response:** None <br>
**Permissions:** User in question or the system. <br>
**Example curl request:** `curl "localhost:4848/image/42:DE:AD:BE:EF:42/latest" --output /tmp/image.img` <br>

#### Download a particular version of an image.
Offers the file associated with a particular version of the image to the user.

**Request:** `GET /image/[UUID]/[version]` <br>
**Body:** None <br>
**Response:** None
**Permissions:** User in question or system <br>
**Example curl request:** `curl "localhost:4848/image/42:DE:AD:BE:EF:42/5" --output /tmp/dead_12.img` <br>

#### Upload a new version of an image
Updates the image with either an entirely new file or a modified version of the original image.

**Request:** `POST /image/[UUID]` <br>
**Body:** Multi-Part image file with the image. <br>
**Response:** Successfuly uploaded image: 5
**Permissions:** User in question or the system. <br>
**Example curl request:** `curl  -X POST localhost:4848/image/87f58936-9540-4dad-aba6-253f06142166 -H "Content-Type: multipart/form-data" -F "file=@/tmp/test3.img"` <br>
