# lily
A lightweight secure network file server written in Go.

## What is Lily?

Lily is a secure network file server written in Go. Lily supports users and permissions, and allows for a wide variety of commands. Lily is designed to be safe and efficient, making it perfect for large-scale file servers.

Lily has a wide variety of features, including but not limited to:
- Multiple file stores
- Custom Lily protocol
- Efficient handling of large amounts of data
- Access permissions and security, including whitelists and blacklists
- Multithreaded safety, allowing for efficient and safe handling of multiple requests at a time
- User and session authentication
- Settings are able to be customized easily at runtime

See [Lily](./LILY.md) for more information.

## Setup

### Building and Installing
To install, run `make install` or `go install` in the root directory. Lily can also be built using `make` or `go build`. 

### Creating a Drive
To create a drive, run `lily drive init <name> <absolutePathToDrive>`. Replace `<name>` with the name of the new drive, and `<absolutePathToDrive>` with the absolute path to the drive directory on the filesystem. This will create a new drive file in the current directory, named `.your-drive-name.lilyd`.

### Creating a Server
To create a new server, start by creating a config file. This config file will only be used once (when the server file is being created) and can be discarded afterwards. An example config file is as follows:
```
[config]
name: server-name

[drives]
drive-name: /absolute/path/to/drive/file

[certs]
certFiles: /absolute/path/to/certificate
keyFiles: /absolute/path/to/key
```

Create the new server by running `lily config init <pathToConfig>`. Replace `<pathToConfig>` with the path to your config file. This should create a new server file in the current directory, named `.server.lily`. To start the server, run `lily serve`. This should find the server file in your current directory, load the drive files, and begin the server.

## Usage

You can access a Lily server using the Go API, with `github.com/cubeflix/lily/client`.
