# Control API
For the BAAS project there exists only one singular way to interface with the control server which is making use of the public facing REST API. These define the functions which can be used to manipulate the server, create images and manage data. In this implementation the REST architecture defined by [Roy Fielding's dissertation](https://www.ics.uci.edu/~fielding/pubs/dissertation/rest_arch_style.htm) and explained on [RESTful API.net](https://restfulapi.net/) is used. All messages to and from the server are in the JSON data format.

## Design
Currently there are two related HTTP interfaces: one user facing and one internal for the management OS. In the future, the aim is to have a singular interface which is used for both parts. In this section the focus is on the latter since this is the main interface and by far the most complex.

There are exists six root uri paths where all other resources have been clustered under, they are as follows:

- `/machine(s)` which defines the functions to manipulate the machines that the images are run on.
- `/user(s)` which allows for the management and creation of users.
- `/user/[name]/image_setups` are the image setups owned by a user
- `/image(s)` is used to access the created images.
- `/v1/boot` is only used for the iPXE server.
- `/static` are the static images and irrelevant for users.
- `/log` where the debug messages from the management OS are sent to.

Each of these routes define their own unique resources which they manage and for each there is a direct correspondence to the related entity in the database. For each of these you can expect at least the basic CRUD functions with some extra user-friendly functionality. All of them, besides the creating user, requires an authentication token since not all features are available to everyone. You can find a full compendium of the endpoints at the end of this page.

In general of all of the resources look like the following:

- `GET /nouns` returns all instances of this object.
- `POST /noun` creates a new instance of the object.
- `GET /noun/name` fetches the instance with the id name.
- `PUT /noun/name` updates the instance with the id name.
- `METHOD /noun/name/verb` performs the specified action for that instance

Different resources may be nested in groups of two arbitrarily deep into other resources, for example, `/machine/[mac]/disk/[uuid]/file/[name]`.

Some endpoints may require a user to be logging in, as indicated by the permissions field in the documentation below, which means that the `session-name` cookie must be set to the right value. This can be done by simply [logging in](logging_in.md), copying the relevant cookie value and using it in your requests. For example, using cURL you want to prefix your commands with: `--cookie "session-name=[some base64 string]"`.


## Endpoint compendium
In this section an overview is given of every single on the defined endpoints together with an example on how to call it, what parameters it takes and what it returns. This section is divided in the same way as the resources defined above.

### Machines
Here the management of the machines is done, where machine is any computer which the images can be run on. This can be used to register new servers to the system, upload new disks or get information about the actions the machine should take.

#### Get information about a machine
Allows a user to get information about a specific machine which is identified by it's MAC address. Please note that different MAC addresses may refer to the exact same machine.

**Request:** `GET /machine/[mac]`<br>
**Body:** None<br>
**Response:**<br>
- *Name:* Human-readable name.<br>
- *Architecture:* Architecture of the machine<br>
- *Managed:* A boolean indicating that BAAS manages the machine<br>
- *MacAddress*: MAC Address which is associated with this machine.<br>

**Permission**: All<br>
**Example curl command**: `curl localhost:8080/machine/00:11:22:33:44:55:66`<br>
**Example response:**<br>
```json
{
    "Name": "Machine 1",
    "Architecture": "X86_64",
	"Managed": true,
    "MacAddress": [{
		 "Address": "00:11:22:33:44:55:66"
	}]
}
```

#### Getting all machines
Receives information about every currently registered machine.

**Request:** `GET /machines`<br>
**Body**: None<br>
**Response:** A list of machine objects described above.<br>
**Permission**: All<br>
**Example curl command:** `curl localhost:8080/machines`<br>
**Example response:**<br>
```json
[{
	"Name": "Machine 1",
    "Architecture": "x86_64",
	"Managed": true,
    "MacAddress": [{
		"Address": "00:11:22:33:44:55:66"
    }]
  },
  {
	"Name": "Machine 2",
    "Architecture": "x86_64",
	"Managed": false,
    "MacAddress": [{
		"Address": "42:DE:AD:BE:EF:42"
    }]
}]
```

#### Create machine
Creates a new machine in the database and creates the base image for
the machine.

**Request:** `POST /machine`<br>
**Body:**<br>
- *Name:* Human-readable for the machine.<br>
- *Architecture:* Architecture of the machine, typically x86\_64<br>
- *Managed:* Boolean indicating that BAAS is managing the machine.<br>
- *MacAddresses:* A list of MAC addresses associated with the system
  in the form of `{"Address": "value"}"`.<br>

**Response:** None<br>
**Permissions:** Administrators<br>
**Example body:**<br>
```json
{
	"Name": "Hello World",
	"Architecture": "x86_64",
	"Managed": true,
	"MacAddress": [{
		"Address": "52:54:00:d9:71:15",
		"MachineModelID": 12
	}]
}
```
**Example curl command:** `curl -X POST localhost:4848/machine -H 'Content-Type: application/json' -d '{"name": "Test", "Architecture": "x86_64", "Managed": true, "MacAddress": {"Address": "52:54:00:d9:71:12"}1}'`

#### Update machine
Change the information of a machine, this also used to create a machine.

**Request:** `PUT /machine`<br>
**Body:**<br>
- *Name:* Human-readable for the machine.<br>
- *Architecture:* Architecture of the machine, typically x86\_64<br>
- *Managed:* Unknown<br>
- *MacAddress:* The MAC address associated with the system in the form
  of `{"mac": "value"}"`.<br>

**Response:** The same as the given body<br>
**Permissions:** Administrators<br>
**Example body:**<br>
```json
{
	"Name": "Hello World",
	"Architecture": "x86_64",
	"Managed": true,
	"MacAddress": [{
		"Mac": "52:54:00:d9:71:15",
	}]
}
```
**Example curl command:** `curl -X PUT localhost:4848/machine -H 'Content-Type: application/json' -d '{"name": "Test", "Architecture": "x86_64", "Managed": true, "MacAddress": {"Address": "52:54:00:d9:71:12"}}'`

#### Get the boot information
Fetches configuration for the next boot for this particular machine using a SQL based FIFO queue.

**Request:** `GET /machine/mac/boot`<br>
**Body:** None<br>
**Response:**<br>
- *Name:* The name of the image setup.<br>
- *Images:* A list of the images associated with the setup and the
  machine image<br>
- *User:* Username of the image setup<br>
- *UUID:* UUID for the image setup<br>

**Permissions:** Management OS<br>
**Example curl request**:` curl localhost:4848/machine/42:DE:AD:BE:EF:42/boot`<br>
**Example response:**<br>
```json
{
  "Name": "Linux Kernel 2",
  "Images": [
        {
      "Image": {
        "Name": "52:54:00:d9:71:93",
        "Versions": [
          {
            "Version": 0,
            "ImageModelUUID": "6c63d514-7314-4d80-bf04-12f1dfa005c5"
          }
        ],
        "UUID": "6c63d514-7314-4d80-bf04-12f1dfa005c5",
        "Username": "",
        "DiskCompressionStrategy": "none",
        "ImageFileType": "raw",
        "Type": "machine",
        "Checksum": "80654151"
      },
      "UUIDImage": "",
      "VersionNumber": 0,
      "Update": false
    }
  ],
  "User": "ValentijnvdBeek",
  "UUID": "f02dc9d1-833e-45e9-9d28-87a5390cbee3"
}
```

#### Add another configuration to a machine queue
Push a boot configuration to the queue in a machine's FIFO boot
queue. In this future this should probably be machine agnostic.

**Request:** `POST /machine/[mac]/boot`<br>
**Body:**<br>
- *MachineModelID:* Machine that the image should be flashed to.
- *SetupUUID:* UUID associated with the image setup
- *Update:* A boolean indicating whether the changes to images should
  be synced<br>

**Response:**<br>
- *MachineModelID:* Machine that the image should be flashed to.<br>
- *ImageUUID:* UUID for the image.<br>
- *Update:* Should the changes be synced to the disk.<br>

**Permissions:** System<br>
**Example curl request:** `curl "localhost:4848/machine/52:54:00:d9:71:93/boot" -H 'application/json' -d '{"Update": false, "SetupUUID": "2b59ff94-7fb6-4239-b2e6-82f1e30f4355", "MachineModelId": 1}' -H "type: system"`<br>
**Example response:**
```json
{
  "MachineModelID": 1,
  "ImageUUID": "74368cec-7903-4233-87b7-564195619dce",
  "Update": true
}
```

### Users
Users are the access control mechanism which is used in the BAAS
project. There are exists three kinds of users: administrators,
moderators and users. Users can only access system images and modify
their own images. Moderators can modify assigned system
images. Administrators can modify any part of the program.

#### Create a new user
Add user to the system.

**Request:** `POST /user`<br>
**Body:**<br>
- *Username:* User name of the new user.<br>
- *Name:* Name of the user.<br>
- *Email:* Email of the user<br>
- *Role:* One of user, moderator or administrator<br>

**Response:** Successfully created user<br>
**Permissions:** Administrators/System<br>
**Example curl request:** `curl -X POST "localhost:4848/user" -H 'Content-Type: application/json' -d '{"Username": "wnarchi", "Name": "William Narchi", "Email": "w.narchi1.obscured@student.tudelft.net", "Role": "user"}'`<br>

#### Login using GitHub
Starts the OAuth2 process as described in [logging in](logging_in.md)

**Request:** `GET /user/login/github`<br>
**Body:** None<br>
**Response:** A link to Github's Authentication page<br>
**Example request:** `curl "localhost:4848/user/login/github"`<br>

#### Fetch a particular user
Returns information about a particular user.

**Request:** `GET /user/[name]`<br>
**Body:** None<br>
**Response:**<br>
- *Username:* Username of the user.<br>
- *Name:* UTF-8 string for the user's name.<br>
- *Email:* Associated email adres of the university.<br>
- *Role:* Assigned permissions<br>

**Permissions:** All<br>
**Example curl request:** `curl "localhost:4848/user/jrellermeyer"`<br>
**Response:**
```json
{
   "Username": "jrellermeyer",
   "Name": "Jan",
   "Email": "j.w.dijkstra@tudelft.nl",
   "Role": "admin"
}
```

#### Fetch the currently logged in user
Finds the user information for the user that is currently logged into the session.

**Request:** `GET /user/me`<br>
**Body:** None<br>
**Response:** Same as [fetch a particular user](#fetch-a-particular-user)<br>
**Example request:** `curl "localhost:4848/user/me" --cookie "session-name=[value]"`<br>

##### Deletes a user
Removes an image setup from the database.

**Request:** `DELETE /user/[name]`<br>
**Body:** None<br>
**Response:** `Successfully deleted user`<br>
**Permissions:** All<br>
**Example curl request:** `curl -X DELETE "localhost:4848/user/ValentijnvdBeek"`<br>

##### Modify image setup
Modifies metadata related to the image setup. Cannot be used to change
the images.

**Request:** `PUT /user/[name]`<br>
**Body:** the wished modifications for the user<br>
**Response:** The modified user object<br>
**Permissions:** All<br>
**Example curl request:** `curl -X PUT "localhost:4848/user/ValentijnvdBeek" -d '{"Name": "Valentijn"}'`<br>
**Example response:**
```json
{
  "Username": "ValentijnvdBeek",
  "Name": "ValentijnvdBeek",
  "Email": "",
  "Role": "admin",
}
```

#### Get all registered users
Gives a list of every user which is currently registered with the system.

**Request:** `GET /users`<br>
**Body:**  None<br>
**Response:** A list of user objects described above.<br>
**Permissions:** All<br>
**Example curl request:** `curl "localhost:4848/users" `<br>
**Example response:**
```json
[
  {
    "Username": "jrellermeyer",
    "Name": "Jan",
    "Email": "j.w.dijkstra@tudelft.nl",
    "Role": "admin",
  },
  {
    "Username": "wnarchi",
    "Name": "William",
    "Email": "j.cena@student.tudelft.nl",
    "Role": "user",
  }
]
```

#### Create a new image
Creates a new image entity and file.

**Request:** `POST /user/[name]/image`<br>
**Body:**<br>
- *Name:* A human-readable name for the user.<br>
- *DiskCompressionStrategy:* How the image is compressed, can be one of: none, gzip or zstd.<br>
- *ImageFileType:* Filesystem type of the image, typically FAT32 or EXT4<br>
- *Type:* BAAS image type, one of: base, system, temporal and temporary<br>
- *Versioned:* Boolean value indicating that it is a versioned or a
  checksum-based image<br>

**Response:**
- *Name:* Human-readable name of the image.<br>
- *Versions:* A list of objects with a Version attribute containing the version number.<br>
- *UUID:* Identifiying unique ID for the image.<br>
- *Username:* Username of the user who owns the image.<br>
- *DiskCompressionStrategy:* Compression used on the disk<br>
- *ImageFileType:* Filesystem installed on the image.<br>
- *Type:* BAAS system type.<br>
- *Checksum:* Checksum in case of a non-versioned image.<br>

**Permissions:** User in question or administrator<br>
**Example curl request:** `curl -X POST "localhost:4848/user/ValentijnvdBeek/image" -H 'Content-Type: application/json' --cookie "session-name=$SECRET" -d '{"Name": "Fedora Research", "DiskCompressionStrategy": "none", "ImageFileType": "raw", "Type": "system"}'`<br>
**Example Response:**
```json
{
  "Name": "Fedora Research",
  "Versions": [
    {
      "Version": 0,
      "ImageModelUUID": "e5983f84-0fd6-4275-a1cf-a39da8236949"
    }
  ],
  "UUID": "e5983f84-0fd6-4275-a1cf-a39da8236949",
  "Username": "ValentijnvdBeek",
  "DiskCompressionStrategy": "none",
  "ImageFileType": "raw",
  "Type": "system",
  "Checksum": ""
}
```

#### Delete an image
Deletes an image and the associated files from the BAAS server. Please
note that it is also possible to do a soft-delete of an image by
setting the managed flag to off.

**Request:** `DELETE /image/{UUID}**`<br>
**Body:** None<br>
**Response:** An error message or *Successfully deleted image*<br>
**Permissions:** The user in question or any administrator<br>
**Example curl request:** `curl -X DELETE "localhost:4848/image/06995218-54f2-4a5d-9022-8324bae1971a"`<br>
**Example Response:** `Successfully deleted image`<br>

#### Update image
Changes the image stored in the database. Please note that it does
not, yet, handle reformatting or recompressioning the images. These
features could be added in the future.

**Request:** `PUT /image/{UUID}**`<br>
**Body:** None<br>
**Response:** An error message or the changed image<br>
**Permissions:** The user in question or any administrator<br>
**Example curl request:** `curl -X PUT "localhost:4848/image/06995218-54f2-4a5d-9022-8324bae1971a" -H 'Content-Type: application/json' -d {"Name": "RealVLC Research", "Type": "System" }`<br>
**Example Response:**
```json
{
  "Name": "RealVLC Research",
  "Versions": [],
  "UUID": "01018664-56c1-4d46-b6fe-fb5c5034a446",
  "Username": "ValentijnvdBeek",
  "DiskCompressionStrategy": "GZip",
  "ImageFileType": "raw",
  "Type": "System",
  "Checksum": ""
}
```

#### Generate a docker image
Takes a Dockerfile, generates an associated image and adds it as
another version to the database.

**Request:** `POST /image/[uuid]/docker`<br>
**Body:** Multipart file with the Dockerfile<br>
**Response:** `Successfully uploaded image: [version]`<br>
**Permissions:** Moderator, Admin and same user<br>
**Example curl request:** `curl -X POST "localhost:4848/image/06995218-54f2-4a5d-9022-8324bae1971a/docker" -F "file=@Dockerfile"`<br>
**Example response:** `Successfully uploaded image: 12`<br>

#### Find all the images made by a user
Returns every image created by the human without versions.

**Request:** `GET /user/[name]/images`<br>
**Body:** None<br>
**Response:** A list of image objcts described above<br>
**Permissions:** User in question or administrator<br>
**Example curl request:** `curl "localhost:4848/user/Jan/images"`<br>
**Example response:**
```json
[
  {
    "Name": "Fedora Research",
    "Versions": [
      {
        "Version": 0,
        "ImageModelUUID": "e5983f84-0fd6-4275-a1cf-a39da8236949"
      },
      {
        "Version": 1,
        "ImageModelUUID": "e5983f84-0fd6-4275-a1cf-a39da8236949"
      },
      {
        "Version": 2,
        "ImageModelUUID": "e5983f84-0fd6-4275-a1cf-a39da8236949"
      },
      {
        "Version": 3,
        "ImageModelUUID": "e5983f84-0fd6-4275-a1cf-a39da8236949"
      },
      {
        "Version": 4,
        "ImageModelUUID": "e5983f84-0fd6-4275-a1cf-a39da8236949"
      }
    ],
    "UUID": "e5983f84-0fd6-4275-a1cf-a39da8236949",
    "Username": "Jan",
    "DiskCompressionStrategy": "none",
    "ImageFileType": "raw",
    "Type": "system",
    "Checksum": ""
  },
  {
    "Name": "Gentoo",
    "Versions": [
      {
        "Version": 0,
        "ImageModelUUID": "64a85bf1-6b76-423c-be18-4c0e1922b778"
      },
      {
        "Version": 1,
        "ImageModelUUID": "64a85bf1-6b76-423c-be18-4c0e1922b778"
      }
    ],
    "UUID": "64a85bf1-6b76-423c-be18-4c0e1922b778",
    "Username": "Jan",
    "DiskCompressionStrategy": "None",
    "ImageFileType": "raw",
    "Type": "Temporal",
    "Checksum": ""
  },
  {
    "Name": "RealVLC",
    "Versions": [
      {
        "Version": 0,
        "ImageModelUUID": "50728327-5a4d-43fa-b61f-df771c4d4f0b"
      }
    ],
    "UUID": "50728327-5a4d-43fa-b61f-df771c4d4f0b",
    "Username": "Jan",
    "DiskCompressionStrategy": "ZSTD",
    "ImageFileType": "raw",
    "Type": "Temporal",
    "Checksum": ""
  },
  {
    "Name": "ENS",
    "Versions": [
      {
        "Version": 0,
        "ImageModelUUID": "b2aa291b-1ca1-4ca8-a59b-4cc57cb8ade9"
      }
    ],
    "UUID": "b2aa291b-1ca1-4ca8-a59b-4cc57cb8ade9",
    "Username": "Jan",
    "DiskCompressionStrategy": "ZSTD",
    "ImageFileType": "raw",
    "Type": "System",
    "Checksum": ""
  }
]
```

#### Get all the images from a user with a particular name
Find all the images which are associated with user which share the same human-readable name. This is a convenience function which can be used for searching through the images since, at the moment, image names are not unique.

**Request:** `GET /user/[username]/images/[name]`<br>
**Body:** None<br>
**Response:** A list of image objects described above filtered on name.<br>
**Permissions:** User in question or administrator<br>
**Example curl request:** `curl "localhost:4848/user/Jan/images/Gentoo"`<br>
**Example response:**

```json
[
  {
    "Name": "Gentoo",
    "Versions": [
      {
        "Version": 0,
        "ImageModelUUID": "64a85bf1-6b76-423c-be18-4c0e1922b778"
      },
      {
        "Version": 1,
        "ImageModelUUID": "64a85bf1-6b76-423c-be18-4c0e1922b778"
      }
    ],
    "UUID": "64a85bf1-6b76-423c-be18-4c0e1922b778",
    "Username": "Jan",
    "DiskCompressionStrategy": "None",
    "ImageFileType": "raw",
    "Type": "Temporal",
    "Checksum": ""
  }
]
```

### Images
Represents the images used for the BAAS project. Endpoints can be
found in the `/image/` resource pool, but also as a part of the
`/user` pool. In particular, the creation of user system images are
typically in the latter rather than the former.

#### Get image info
Offers the underlying image file to the user.

**Request:** `GET /image/[UUID]`<br>
**Body:** None<br>
**Permissions:** User in question or administrator<br>
**Example curl request:** `curl "localhost:4848/image/b2aa291b-1ca1-4ca8-a59b-4cc57cb8ade9`<br>
**Response:**<br>
- *Name:* Human-readable name associated with the image.<br>
- *Versions:* A list of versions available for this image. Versions are JSON objects with a Version attribute.<br>
- *UUID:* Image unique ID.<br>
- *Username:* User who owns the image<br>
- *DiskCompressionStrategy:* How the image is compressed.<br>
- *ImageFileType:* Image filesystem.<br>
- *Type:* Indicates the type of BAAS image.<br>
- *Checksum:* Checksum in case of non-versioned images.<br>

**Example response:**
```json
{
  "Name": "ENS",
  "Versions": [
    {
      "Version": 0,
      "ImageModelUUID": "b2aa291b-1ca1-4ca8-a59b-4cc57cb8ade9"
    }
  ],
  "UUID": "b2aa291b-1ca1-4ca8-a59b-4cc57cb8ade9",
  "Username": "ValentijnvdBeek",
  "DiskCompressionStrategy": "ZSTD",
  "ImageFileType": "raw",
  "Type": "System",
  "Checksum": ""
}
```

#### Get the latest version of an image
Convenience function which offers the latest version of the specified image.

**Request:** `GET /image/[name]/latest`<br>
**Body:** None.<br>
**Response:** None<br>
**Permissions:** User in question or the system.<br>
**Example curl request:** `curl "localhost:4848/image/42:DE:AD:BE:EF:42/latest" --output /tmp/image.img`

#### Download a particular version of an image.
Offers the file associated with a particular version of the image to the user.

**Request:** `GET /image/[UUID]/[version]`<br>
**Body:** None<br>
**Response:** None<br>
**Permissions:** User in question or system<br>
**Example curl request:** `curl "localhost:4848/image/42:DE:AD:BE:EF:42/5" --output /tmp/dead_12.img`

#### Upload a new version of an image
Updates the image with either an entirely new file or a modified version of the original image.

**Request:** `POST /image/[UUID]`<br>
**Body:** Multi-Part image file with the image.<br>
**Response:** Successfuly uploaded image: 5<br>
**Permissions:** User in question or the system.<br>
**Example curl request:** `curl  -X POST localhost:4848/image/87f58936-9540-4dad-aba6-253f06142166 -H "Content-Type: multipart/form-data" -F "newVersion=[false,true];file=@/tmp/test3.img"`

### Image setups
Although useful, simply being able to flash a singular image onto a
server is not a particularly novel feature. BAAS differs from other
solutions by allow for any configuration of images to a system where
each image may have particular semantic meanings to it. For example,
an user image contains only information owned by a user and should
always be updated while a system image is owned by the system and
should rarely be updated.

##### Create a new image setup
Adds a new image setup with no associated images to the database.

**Request:** `POST /user/[name]/image_setup`<br>
**Body:** None<br>
**Response:** `Successfully created image setup`<br>
**Permissions:** All<br>
**Example curl request:** `curl -X POST "localhost:4848/user/ValentijnvdBeek/image_setup"`

##### Deletes a image setup
Removes an image setup from the database.

**Request:** `DELETE /user/[name]/image_setup/[setup_uuid]`<br>
**Body:** None<br>
**Response:** `Successfully deleted image setup`<br>
**Permissions:** All<br>
**Example curl request:** `curl -X DELETE "localhost:4848/user/ValentijnvdBeek/image_setup/ae8567e6-f5b4-46b0-8afc-42f425d00194"`

##### Modify image setup
Modifies metadata related to the image setup. Cannot be used to change
the images.

**Request:** `PUT /user/[name]/image_setup/[setup_uuid]`<br>
**Body:** Modified image setup object<br>
**Response:** The modified image object<br>
**Permissions:** All<br>
**Example curl request:** `curl -X PUT "localhost:4848/user/ValentijnvdBeek/image_setup/ae8567e6-f5b4-46b0-8afc-42f425d00194" -d '{"Name": "GNU/Linux Research"}'`<br>
**Example response:**
```json
{
  "Name": "GNU/Linux Research",
  "Images": null,
  "Username": "ValentijnvdBeek",
  "UUID": "ae8567e6-f5b4-46b0-8afc-42f425d00194"
}
```

##### Get image setups by user
Gets all the image setups associated by user

**Request:** `GET /user/[name]/image_setups`<br>
**Body:** None<br>
**Response:** List of user images objects similar to [get image setup](#get-image-setup)<br>
**Permission:** User in question, system, moderator and administrator<br>
**Example curl request:** `curl -X GET "localhost:4848/user/ValentijnvdBeek/image_setups"`<br>
**Example Response:**
```json
[{
  "Name": "Linux Kernel 2",
  "Images": [
    {
      "Image": {
        "Name": "Fedora",
        "Versions": null,
        "UUID": "e9c62845-7a03-4d9d-8132-2dbf715d6159",
        "Username": "ValentijnvdBeek",
        "DiskCompressionStrategy": "none",
        "ImageFileType": "raw",
        "Type": "base",
        "Checksum": ""
      },
      "UUIDImage": "e9c62845-7a03-4d9d-8132-2dbf715d6159",
      "VersionNumber": 0,
      "Update": false
    }
  ],
  "User": "ValentijnvdBeek",
  "UUID": "f02dc9d1-833e-45e9-9d28-87a5390cbee3"
}]
```

##### Get image setup
Gets the data associated particular image setup, in particular
those images that are linked to it together with which version.

**Request:** `GET /user/[name]/image_setup/[UUID]`<br>
**Body:** None<br>
**Response:**<br>
- *User:* Username of the user who owns the setup.<br>
- *UUID:* UUID of the setup.<br>
- *Name:* The name of the image setup.<br>
- *Images:* A list of images and version numbers.<br>
- *Images.Image:* An object containing the image.<br>
- *Images.UUIDImage:* UUID of the image.<br>
- *Images.VersionNumber:* Version linked to the setup.<br>
- *Images.Update:* Should the image be updated after running<br>

**Permissions:** User in question, system, moderator and administrator<br>
**Example curl request:** `curl -X GET
"localhost:4848/user/ValentijnvdBeek/image_setup/f02dc9d1-833e-45e9-9d28-87a5390cbee3"`

**Example response:**
```json
{
  "Name": "Linux Kernel 2",
  "Images": [
    {
      "Image": {
        "Name": "Fedora",
        "Versions": null,
        "UUID": "e9c62845-7a03-4d9d-8132-2dbf715d6159",
        "Username": "ValentijnvdBeek",
        "DiskCompressionStrategy": "none",
        "ImageFileType": "raw",
        "Type": "base",
        "Checksum": ""
      },
      "UUIDImage": "e9c62845-7a03-4d9d-8132-2dbf715d6159",
      "VersionNumber": 0,
      "Update": false
    }
  ],
  "User": "ValentijnvdBeek",
  "UUID": "f02dc9d1-833e-45e9-9d28-87a5390cbee3"
}
```

##### Get images from the image setup
Finds all the images associated with a given image setup.

**Request:** `GET /user/[name]/image_setup/[uuid]/images`<br>
**Body:** None<br>
**Response:** A list of images<br>
**Permissions:** User in question, moderator and administrator.<br>
**Example curl request:** `curl "localhost:4848/user/ValentijnvdBeek/image_setup/2b59ff94-7fb6-4239-b2e6-82f1e30f4355/images" -h 'Content-Type: application/json' -d '{"Uuid": "3a760707-c160-40fa-81be-430b75131ddc", "Version": 3}'`<br>
**Example response:**
```json
{
  "Name": "GNU/Linux Research",
  "Images": [
    {
      "Image": {
        "Name": "EEMCS 2",
        "Versions": null,
        "UUID": "01018664-56c1-4d46-b6fe-fb5c5034a446",
        "Username": "ValentijnvdBeek",
        "DiskCompressionStrategy": "GZip",
        "ImageFileType": "raw",
        "Type": "Temporal",
        "Checksum": ""
      },
      "UUIDImage": "01018664-56c1-4d46-b6fe-fb5c5034a446",
      "VersionID": 3,
      "Update": false
    }
  ],
  "Username": "ValentijnvdBeek",
  "UUID": "ae8567e6-f5b4-46b0-8afc-42f425d00194"
}
```

##### Add image to image setup
Links an image to the given image setup.

**Request:** `POST /user/[name]/image_setup/[uuid]/images`<br>
**Body:**<br>
- *Uuid:* UUID of the image you want to link.<br>
- *Version:* Version that you would like to link.<br>

**Response:** The same response as getting the image setup.<br>
**Permissions:** User in question, moderator and administrator.<br>
**Example curl request:** `curl -X POST "localhost:4848/user/ValentijnvdBeek/image_setup/2b59ff94-7fb6-4239-b2e6-82f1e30f4355/images" -h 'Content-Type: application/json' -d '{"Uuid": "3a760707-c160-40fa-81be-430b75131ddc", "Version": 3}'`<br>
**Example body:** `{"Uuid": "3a760707-c160-40fa-81be-430b75131ddc", "Version": 3}`<br>
**Example response:** See get image setup

##### Remove image from image setup
Removes an image from the given image setup.

**Request:** `DELETE /user/[name]/image_setup/[uuid]/images`<br>
**Body:**<br>
- *Uuid:* UUID of the image you want to remove.<br>
- *Version:* Version that you would like to remove.<br>

**Response:** *Successfully deleted image from setup*<br>
**Permissions:** User in question, moderator and administrator.<br>
**Example curl request:** `curl -X DELETE "localhost:4848/user/ValentijnvdBeek/image_setup/2b59ff94-7fb6-4239-b2e6-82f1e30f4355/images" -h 'Content-Type: application/json' -d '{"Uuid": "3a760707-c160-40fa-81be-430b75131ddc", "Version": 3}'`<br>
**Example body:** `{"Uuid": "3a760707-c160-40fa-81be-430b75131ddc", "Version": 3}`<br>
**Example response:** *Successfully deleted image from setup*


##### Find an image setups based on name
Finds an image setup based on a name

**Request:** `GET /user/[name]/image_setup`<br>
**Body:** None.<br>
**Response:** <br>
A list containing the following:<br>
- *Name:* Name of the image setup.<br>
- *Images:* Possibly optional empty list of images associated with the
  setup.<br>
- *User:* Username of the user owning the setup.<br>
- *UUID:* UUID of the image setup<br>

**Permissions:** User in question, moderator and administrator<br>
**Example curl request:** `curl "localhost:4848/user/ValentijnvdBeek/image_setup"`<br>
**Example response:**
```json
[
  {
    "Name": "Linux Kernel 2",
    "Images": null,
    "User": "ValentijnvdBeek",
    "UUID": "f02dc9d1-833e-45e9-9d28-87a5390cbee3"
  }
]
```
