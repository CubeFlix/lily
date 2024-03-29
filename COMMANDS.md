# Commands
Commands in Lily consist of the following fields (see [the protocol](./PROTOCOL.md#command)):

| Name        | Description     | Type   |
| -           | -               | -      |
| Auth        | The authentication data. | security.auth.Auth |
| Command ID  | The command ID. | string |
| Arguments   | The command arguments. | Any |
| Chunks          | The chunk data. | Chunks |

Lily commands accept two types of authentication: user and session. User authentication takes a username and password, while session authentication takes a username and session ID. Most commands allow both session and user authentication, however, some commands require only a certain type. For example, login commands require user authentication as they need a username and password.

Responses consist of the following fields:

| Name            | Description     | Type   |
| -               | -               | -      |
| Chunks          | The chunk data. | Chunks |
| Response Code   | The response code. | int |
| Response String | A string containing the response message, if necessary. | string |
| Data            | The response data. | Any |

## General Commands

### Ping
> Ping the server. This command does not require authentication, and clients may provide an empty user authentication object.

**Parameters:** None

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Info
> Get basic server information. This command does not require authentication, and clients may provide an empty user authentication object.

**Parameters:** None

**Chunk Arguments:** None

**Returns:** 

> - `name` (type `string`)
> 
>   The name of the server.
> 
> - `version` (type `string`)
> 
>   The server version.
> 
> - `drives` (type `[]string`)
> 
>   A list of the drives on the server.
> 
> - `defaultSessionExpiration` (type `time.Duration`)
> 
>   The default session expiration time.
>  
> - `allowChangeSessionExpiration` (type `bool`)
> 
>   If clients are allowed to modify or specify the expiration time.
> 
> - `allowNonExpiringSessions` (type `bool`)
> 
>   If the server allows clients to specify a never-expiring session.
> 
> - `timeout` (type `time.Duration`)
> 
>   The timeout duration for receiving requests.
> - `limit` (type `time.Duration`)
>   
>   The rate limit interval.
> - `maxLimitEvents` (type `int`)
>   
>   The maximum number of events per limit interval.

**Chunk Returns:** None

### Login
> Login to the server. This command, if successful, will create a session and return the resultant session ID. This command requires user authentication.

**Parameters:** 

> - `sessionExpiration` (type `time.Duration`)
> 
>   If the server allows clients to specify the expiration time, this argument will specify the expiration time. If the server does not allow clients to set the expiration time, this argument does nothing. If the server allows never-expiring sessions and the value for the session expiration time is 0, the session will never expire. Returns an error if the session expiration time is 0 and the server does not allow it.

### Logout
> Log out of the server. This command, if successful, will remove the associated session. This command requires session authentication.

**Parameters:** None

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

## Administrative Commands
These following commands all require administrator privileges to execute.

### Get All Users
> Get a list of all users.

**Parameters:** None

**Chunk Arguments:** None

**Returns:** 

> - `users` (type `[]string`)
> 
>   The list of users.

**Chunk Returns:** None

### Get User Information
> Get information about users, given a list of users. Returns the user's name, clearance level, and bcrypt password hash as a UserInfo object. If a given user does not exist, the command returns an error.

**Parameters:** 

> - `users` (type `[]string`)
> 
>   The list of users.

**Chunk Arguments:** None

**Returns:** 

> - `info` (type `[]UserInfo`)
> 
>   The list of user information.

**Chunk Returns:** None

### Set User Clearance
> Set users' clearance level, given a list of users. If a given user does not exist, it returns an error. If the lengths of the two parameters are not the same, it returns an error.

**Parameters:** 

> - `users` (type `[]string`)
> 
>   The list of users to modify.
> - `clearances` (type `[]int`)
> 
>   The list of new clearance levels. 

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Set User Password

> Set users' password, given a list of users. If a given user does not exist, it returns an error. If the lengths of the two parameters are not the same, it returns an error.

**Parameters:** 

> - `users` (type `[]string`)
> 
>   The list of users to modify.
> - `passwords` (type `[]string`)
> 
>   The list of new passwords. 

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Create Users

> Create users, given a list of users, passwords, and clearance levels. If a given user already exists, it returns an error. If the lengths of the three parameters are not the same, it returns an error.

**Parameters:** 

> - `users` (type `[]string`)
> 
>   The list of users to create.
> - `passwords` (type `[]string`)
> 
>   The list of passwords for the new users.
> - `clearances` (type `[]int`)
> 
>   The list of clearance levels for the new users. 

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Delete Users

> Delete users, given a list of users. If a given user does not exist, it returns an error. 

**Parameters:** 

> - `users` (type `[]string`)
> 
>   The list of users to delete.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Get All Sessions

> Get a list of all sessions active on the server, returning their IDs.

**Parameters:** None

**Chunk Arguments:** None

**Returns:** 

> - `ids` (type `[]uuid.UUID`)
> 
>   The list of session IDs.

**Chunk Returns:** None

### Get All User Sessions

> Get a list of all sessions for a specific user, returning their IDs.

**Parameters:** 

> - `user` (type `string`)
> 
>   The user to get.

**Chunk Arguments:** None

**Returns:** 

> - `ids` (type `[]uuid.UUID`)
> 
>   The list of session IDs.

**Chunk Returns:** None

### Get Session Info

> Get information about sessions, given a list of session IDs. Returns the session's ID, username, next expiration time, and default expiration time as a SessionInfo object. If a given session ID does not exist, it returns an error.

**Parameters:** 

> - `ids` (type `[]uuid.UUID`)
> 
>   The list of session IDs.

**Chunk Arguments:** None

**Returns:** 

> - `sessions` (type `[]SessionInfo`)
> 
>   The list of session information.

**Chunk Returns:** None

### Expire All Sessions

> Expire all expired sessions on the server.

**Parameters:** None

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Expire Sessions

> Expire sessions, given a list of session IDs. If a given session ID does not exist, it returns an error.

**Parameters:** 

> - `ids` (type `[]uuid.UUID`)
> 
>   The list of session IDs.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Get All Settings

> Get all the Lily server settings.

**Parameters:** None

**Chunk Arguments:** None

**Returns:** 

> - `host` (type `string`)
> 
>   The host string.
> - `port` (type `int`)
> 
>   The port.
> - `drives` (type `[]string`)
> 
>   A list of drives.
> - `driveFiles` (type `map[string]string`)
> 
>   A map of drive names and drive files.
> - `numWorkers` (type `int`)
> 
>   The number of server workers.
> - `mainCronInterval` (type `time.Duration`)
> 
>   The main cron interval duration.
> - `sessionCronInterval` (type `time.Duration`)
> 
>   The session expiration cron interval duration.
> - `networkTimeout` (type `time.Duration`)
> 
>   The network timeout duration.
> - `verbose` (type `bool`)
> 
>   If the server is verbose.
> - `logToFile` (type `bool`)
> 
>   If the server should log to a file.
> - `logJSON` (type `bool`)
> 
>   If the server should log JSON output.
> - `logLevel` (type `string`)
> 
>   The threshold logging level to log.
> - `logFile` (type `string`)
> 
>   The path to the log file. If the server does not log to a file, this is empty.
> - `limit` (type `time.Duration`)
>   
>   The rate limit interval.
> - `maxLimitEvents` (type `int`)
>   
>   The maximum number of events per limit interval.

**Chunk Returns:** None

### Set Host and Port

> Set the Lily server host and port. This WILL NOT update the active server, but will update after the server is restarted.

**Parameters:** 
> - `host` (type `string`)
> 
>   The new host.
> - `port` (type `int`)
> 
>   The new port.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Add Drive

> Add a drive to the Lily server. Accepts a drive name and absolute drive file path. If the drive is invalid, this returns an error. If the path is not absolute, this returns an error.

**Parameters:** 

> - `name` (type `string`)
> 
>   The new drive name.
> - `path` (type `string`)
> 
>   The drive file path.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Rename Drive

> Rename a drive on the Lily server. If the new name is taken or invalid, this returns an error.

**Parameters:** 

> - `drive` (type `string`)
> 
>   The drive to rename.
> - `newName` (type `string`)
> 
>   The new drive name.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Remove Drive

> Remove a drive on the Lily server. Accepts a drive name. If the drive name is invalid, this returns an error. This DOES NOT remove the drive or drive file from the host filesystem, instead, it removes the drive from the server.

**Parameters:** 
> - `drive` (type `string`)
> 
>   The drive to remove.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Set Num Workers

> Set the number of workers. This does not update the server. If the number of workers is invalid, this returns an error.

**Parameters:** 
> - `numWorkers` (type `int`)

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Set Cron Intervals

> Set the cron intervals (main and session).

**Parameters:** 
> - `mainInterval` (type `time.Duration`)
> 
>   The main interval.
> - `sessionInterval` (type `time.Duration`)
> 
>   The session interval.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Set Timeout Interval

> Set the timeout interval. If the interval time is invalid, this returns an error.

**Parameters:** 
> - `timeout` (type `time.Duration`)
> 
>   The timeout interval. 

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Set Logging Settings

> Set the logging settings. This WILL NOT update the active server, but will update after the server is restarted. If log level is invalid, this returns an error.

**Parameters:** 
> - `verbose` (type `bool`)
> 
>   If the server should log.
> - `logToFile` (type `bool`)
> 
>   If the server should log to a file.
> - `logJSON` (type `bool`)
> 
>   If the server should log JSON.
> - `logLevel` (type `string`)
> 
>   The threshold logging level. Should be `debug`, `info`, `warning`, or `fatal`.
> - `logPath` (type `string`)
> 
>   The path for the file to log to. If the server does not log to a file, this should be an empty string.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Set Rate Limit

> Set the rate limit. This WILL NOT update the active server, but will update after the server is restarted. 

**Parameters:** 
> - `limit` (type `time.Duration`)
> 
>   The new rate limit interval.
> - `maxLimitEvents` (type `int`)
> 
>   The max number of events per interval.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Shutdown

> Shutdown the Lily server and save.

**Parameters:** None

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Get Memory Usage

> Get memory usage stats.

**Parameters:** None

**Chunk Arguments:** None

**Returns:**

> - `alloc` (type `int64`)
> 
>   Total bytes of allocated heap objects.
> - `total` (type `int64`)
> 
>   Total cumulative bytes allocated for heap objects. Does not decrease as objects are freed.
> - `sys` (type `int64`)
> 
>   Total bytes of memory obtained from the OS. 

**Chunk Returns:** None

## User Commands

### Set Password

> Set the password for the currently logged in user. Requires user or session authentication.

**Parameters:** 

> - `password` (type `string`)
> 
>   The new password.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

## Session Commands

### Reauthenticate

> Reauthenticate the session. This will update the expiration. This command requires session authentication. If the authentication is invalid, this will return an error.

**Parameters:** None

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Set Expiration Time

> Set the expiration time for the session. If the server does not allow setting the expiration time, this command does not return an error, and instead does nothing.

**Parameters:** 

> - `sessionExpiration` (type `time.Duration`)
> 
>   The new expiration time.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

## Filesystem Commands

### Create Dirs

> Create directories. Requires drive modify clearance.

**Parameters:** 

> - `drive` (type `string`)
> 
>   The name of the drive.
> - `paths` (type `[]string`)
> 
>   The directories to create.
> - `settings` (type `[]BSONAccessSettings`)
> 
>   Optional. The settings for the paths.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Create Dir Tree

> Create a directory tree. Requires drive modify clearance.

**Parameters:** 

> - `drive` (type `string`)
> 
>   The name of the drive.
> - `parent` (type `string`)
> 
>   The parent directory.
> - `paths` (type `[]string`)
> 
>   The directories to create. Must be in the parent directory.
> - `parentSettings` (type `BSONAccessSettings`)
> 
>   Optional. The settings for the parent.
> - `settings` (type `[]BSONAccessSettings`)
> 
>   Optional. The settings for the paths.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### List Dir

> List the contents of a directory.

**Parameters:** 

> - `drive` (type `string`)
> 
>   The name of the drive.
> - `path` (type `string`)
> 
>   The path.

**Chunk Arguments:** None

**Returns:** 

> - `list` (type `map[string]PathStatus`)
> 
>   The contents of the directory.

**Chunk Returns:** None

### Rename Dirs

> Rename directories within their respective directories.

**Parameters:** 

> - `drive` (type `string`)
> 
>   The name of the drive.
> - `paths` (type `[]string`)
> 
>   The paths to rename.
> - `newNames` (type `[]string`)
>  
>   The new names for the paths.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Move Dirs

> Move directories throughout the drive.

**Parameters:** 

> - `drive` (type `string`)
> 
>   The name of the drive.
> - `paths` (type `[]string`)
> 
>   The paths to move.
> - `dests` (type `[]string`)
>  
>   The new destinations.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Delete Dirs

> Delete directories.

**Parameters:** 

> - `drive` (type `string`)
> 
>   The name of the drive.
> - `paths` (type `[]string`)
> 
>   The paths to delete.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Create Files

> Create files. Requires drive modify clearance.

**Parameters:** 

> - `drive` (type `string`)
> 
>   The name of the drive.
> - `paths` (type `[]string`)
> 
>   The files to create.
> - `settings` (type `[]BSONAccessSettings`)
> 
>   Optional. The settings for the paths.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Read Files

> Read files. Requires access clearance.

**Parameters:** 

> - `drive` (type `string`)
> 
>   The name of the drive.
> - `paths` (type `[]string`)
> 
>   The files to read.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** Chunked files.

### Write Files

> Write files. Requires modify clearance.

**Parameters:** 

> - `drive` (type `string`)
> 
>   The name of the drive.
> - `paths` (type `[]string`)
> 
>   The files to write.

**Chunk Arguments:** Chunked files.

**Returns:** None

**Chunk Returns:** None

### Rename Files

> Rename files. Requires modify clearance.

**Parameters:** 

> - `drive` (type `string`)
> 
>   The name of the drive.
> - `paths` (type `[]string`)
> 
>   The files to rename.
> - `newNames` (type `[]string`)
> 
>   The new names.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Move Files

> Move files. Requires modify clearance.

**Parameters:** 

> - `drive` (type `string`)
> 
>   The name of the drive.
> - `paths` (type `[]string`)
> 
>   The files to move.
> - `dests` (type `[]string`)
> 
>   The destinations.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Delete Files

> Delete files. Requires modify clearance.

**Parameters:** 

> - `drive` (type `string`)
> 
>   The name of the drive.
> - `paths` (type `[]string`)
> 
>   The files to delete.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Stat

> Status command. Does not require access clearance.

**Parameters:** 

> - `drive` (type `string`)
> 
>   The name of the drive.
> - `paths` (type `[]string`)
> 
>   The files to get the status of.

**Chunk Arguments:** None

**Returns:** 

> - `stat` (type `map[string]PathInfo`)
> 
>   The statuses of the paths.

**Chunk Returns:** None

### Rehash Files

> Rehash files. Require modify clearance.

**Parameters:** 

> - `drive` (type `string`)
> 
>   The name of the drive.
> - `paths` (type `[]string`)
> 
>   The files to rehash.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Verify Hashes

> Verify hashes for files. Require access clearance.

**Parameters:** 

> - `drive` (type `string`)
> 
>   The name of the drive.
> - `paths` (type `[]string`)
> 
>   The files to verify.

**Chunk Arguments:** None

**Returns:** 

> - `results` (type `map[string]bool`)
> 
>   The hash results.

**Chunk Returns:** None

### Get Path Settings

> Get a path's settings. Requires access clearance. If the path does not exist, this returns an error.

**Parameters:** 

> - `drive` (type `string`)
> 
>   The name of the drive.
> - `path` (type `string`)
> 
>   The path.

**Chunk Arguments:** None

**Returns:** 

> - `settings` (type `BSONAccessSettings`)
> 
>   The access settings.

**Chunk Returns:** None

### Set Path Settings

> Set the path's access settings. Requires modify clearance.

**Parameters:** 

> - `settings` (type `BSONAccessSettings`)
> 
>   The new access settings.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Set Path Clearances

> Set the access and modify clearances for a path. Requires modify clearance. If the clearances are invalid, this returns an error.

**Parameters:** 

> - `drive` (type `string`)
> 
>   The name of the drive.
> - `path` (type `string`)
> 
>   The path.
> - `access` (type `int`)
> 
>   The new access clearance.
> - `modify` (type `int`)
> 
>   The new modify clearance.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Add to Path Access Whitelist

> Add users to a path's access whitelist. Requires modify clearance. If a given username already exists, it will be skipped.

**Parameters:** 

> - `drive` (type `string`)
> 
>   The name of the drive.
> - `path` (type `string`)
> 
>   The path.
> - `users` (type `[]string`)
> 
>   The usernames to add.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Remove from Path Access Whitelist

> Remove users from a path's access whitelist. Requires modify clearance. If a given username does not exist, this returns an error.

**Parameters:** 

> - `drive` (type `string`)
> 
>   The name of the drive.
> - `path` (type `string`)
> 
>   The path.
> - `users` (type `[]string`)
> 
>   The usernames to remove.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Add to Path Access Blacklist

> Add users to a path's access blacklist. Requires modify clearance. If a given username already exists, it will be skipped.

**Parameters:** 

> - `drive` (type `string`)
> 
>   The name of the drive.
> - `path` (type `string`)
> 
>   The path.
> - `users` (type `[]string`)
> 
>   The usernames to add.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Remove from Path Access Blacklist

> Remove users from a path's access blacklist. Requires modify clearance. If a given username does not exist, this returns an error.

**Parameters:** 

> - `drive` (type `string`)
> 
>   The name of the drive.
> - `path` (type `string`)
> 
>   The path.
> - `users` (type `[]string`)
> 
>   The usernames to remove.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Add to Path Modify Whitelist

> Add users to a path's modify whitelist. Requires modify clearance. If a given username already exists, it will be skipped.

**Parameters:** 

> - `drive` (type `string`)
> 
>   The name of the drive.
> - `path` (type `string`)
> 
>   The path.
> - `users` (type `[]string`)
> 
>   The usernames to add.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Remove from Path Modify Whitelist

> Remove users from a path's modify whitelist. Requires modify clearance. If a given username does not exist, this returns an error.

**Parameters:** 

> - `drive` (type `string`)
> 
>   The name of the drive.
> - `path` (type `string`)
> 
>   The path.
> - `users` (type `[]string`)
> 
>   The usernames to remove.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Add to Path Modify Blacklist

> Add users to a path's modify blacklist. Requires modify clearance. If a given username already exists, it will be skipped.

**Parameters:** 

> - `drive` (type `string`)
> 
>   The name of the drive.
> - `path` (type `string`)
> 
>   The path.
> - `users` (type `[]string`)
> 
>   The usernames to add.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None

### Remove from Path Modify Blacklist

> Remove users from a path's modify blacklist. Requires modify clearance. If a given username does not exist, this returns an error.

**Parameters:** 

> - `drive` (type `string`)
> 
>   The name of the drive.
> - `path` (type `string`)
> 
>   The path.
> - `users` (type `[]string`)
> 
>   The usernames to remove.

**Chunk Arguments:** None

**Returns:** None

**Chunk Returns:** None