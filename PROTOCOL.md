# Protocol
## Requests
Lily requests contain three parts: authentication information, request information, and chunk data. Each request consists of the following fields:

| Name        | Description     | Type   |
| -           | -               | -      |
| Header  | The request header information. | [Header](#header) |
| Auth        | The authentication data. | [Authentication](#authentication) |
| Command     | The command data.        | [Command](#command) |
| Chunks      | The chunk data.          | [Chunks](#chunks)   |
| Footer      | The request footer       | [Footer](#footer)   |

Note that all strings with arbitrary length are encoded as a 32-bit unsigned int containing the length of the string, followed by the string itself. All numbers in Lily use little-endian. Arrays with arbitrary length will also encode the length of the array as 32-bit unsigned int, followed by the array data.

## Header
The header information is the same for both requests and responses. It consists of a single UTF-8 encoded string: `LILY` and the Lily protocol version, which is encoded as a string of arbitrary length. If the protocol version for the client and server do not match, the Lily server will respond with an error. The Lily server header is to ensure that the data being received is encoded properly.

| Name        | Description     | Type   |
| -           | -               | -      |
| Header  | The request header. | `"LILY"` |
| Version     | The Lily protocol version. | `string` |

## Footer
The footer information is the same for all request fields. It consists of a single UTF-8 encoded string: `END`. Footers are used to ensure that the command is encoded properly and that the information received is not corrupted.

## Authentication
The authentication information encoding varies depending on the type of authentication: user or session. Lily servers can identify the type of authentication by the authentication type field: a string of length 1 that can either be "U" for or and "S" for session authentication.

### User Authentication
| Name        | Description     | Type   |
| -           | -               | -      |
| Type  | The authentication type. Here it is the string `U`. | `string` (length 1) |
| Username | The username. | `string` |
| Password | The password. | `string` |
| Footer   | The authentication data footer. | [Footer](#footer) |

### Session Authentication
| Name        | Description     | Type   |
| -           | -               | -      |
| Type  | The authentication type. Here it is the string `S`. | `string` (length 1) |
| Username | The username. | `string` |
| Session ID | The session ID. It is a UUID and is thus encoded as a 16-byte byte array. | `[]byte` (length 16) |
| Footer   | The authentication data footer. | [Footer](#footer) |

## Command

The command data consists of the name of the command, followed by the command arguments, which is encoded differently depending on the command.

| Name        | Description     | Type   |
| -           | -               | -      |
| Name  | The command name. | `string` |
| Args  | The command arguments. | Any |
| Footer | The command footer. | [Footer](#footer) |

## Chunks

Lily servers use chunks in order to handle large amounts of file data. The chunk data encoded at the end of each Lily request/response can store multiple streams of information. This allows the server and client to differentiate between multiple files being transferred on the same request/response. The chunk data starts with a list of all chunk streams that will be transferred. Then, the chunks themselves will be encoded. Note that the chunk streams must be encoded in order.

| Name        | Description     | Type   |
| -           | -               | -      |
| Stream info  | The list of all chunk streams. | `[]ChunkStreamInfo` (see [Chunk Stream Info](#chunk-stream-info)) |
| Chunks  | The chunks themselves. Note that these are not encoded as an array, rather, they are encoded one after the other | Stream of [Chunks](#chunk) |
| Footer | The chunks' footer. | [Footer](#footer) |

### Chunk Stream Info
| Name        | Description     | Type   |
| -           | -               | -      |
| Name        | The name of the chunk. | `string` |
| Num Chunks      | The number of chunks associated with the chunk stream. | `uint32` |

### Chunk

| Name        | Description     | Type   |
| -           | -               | -      |
| Name        | The name of the chunk. | `string` |
| Length      | The length of the chunk data. | `uint64` |
| Data        | The chunk data.               | `[]byte` (length Length) |
| Footer      | The chunk footer. | [Footer](#footer)