# Commands
Commands in Lily consist of the following fields:

| Name        | Description     | Type   |
| -           | -               | -      |
| Command ID  | The command ID. | string |
| Auth        | The authentication data. | security.auth.Auth |
| Arguments   | The command arguments. | Any |

Lily commands accept two types of authentication: user and session. User authentication takes a username and password, while session authentication takes a username and session ID. Most commands allow both session and user authentication, however, some commands require only a certain type. For example, login commands require user authentication as they need a username and password.

Responses consist of the following fields:

| Name            | Description     | Type   |
| -               | -               | -      |
| Response Code   | The response code. | int |
| Response String | A string containing the response message, if necessary. | string |
| Data            | The response data. | Any |

## Basic Commands

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

**Chunk Returns:** None

### Login
> Login to the server. This command, if successful, will create a session and return the resultant session ID. This command requires user authentication.

**Parameters:** 

> - `sessionExpiration` (type `time.Duration`)
> 
>   If the server allows clients to specify the expiration time, this argument will specify the expiration time. If the server does not allow clients to set the expiration time, this argument does nothing. If the server allows never-expiring sessions and the value for the session expiration time is 0, the session will never expire. Returns an error if the session expiration time is 0 and the server does not allow it.

**Chunk Arguments:** None

**Returns:** 

> - `sessionID` (type `uuid.UUID`)
> 
>   The new session ID, if the login was successful.

**Chunk Returns:** None



