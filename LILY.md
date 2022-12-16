# Lily
Lily is a secure network file server written in Go. Lily supports users and permissions, and allows for a wide variety of commands. Lily is designed to be safe and efficient, making it perfect for large-scale file servers.

Lily has a wide variety of features, including but not limited to:
- Multiple file stores
- Custom Lily protocol
- Efficient handling of large amounts of data
- Access permissions and security, including whitelists and blacklists
- Multithreaded safety, allowing for efficient and safe handling of multiple requests at a time
- User and session authentication
- Settings are able to be customized easily at runtime

Each Lily server consists of one or more *drives*, which function as individual file stores. Each drive corresponds to a directory on the host's filesystem, but is entirely managed by Lily at runtime. Every directory and file is secured with a set of *access settings*, which control access and modify permissions for the object and the settings themselves. Access settings allow administrators to set *clearance levels*, a number from 1-5 that is assigned to each user. Using access/modify clearance levels, along with whitelists and blacklists, Lily admins can control access permissions easily and effectively. Server settings and user settings can also be modified by users with level 5 access: administrator permissions.

Lily clients and servers run via the Lily protocol, which itself runs on TLS. The Lily protocol encodes important information about the Lily server, the request/response, and authentication. Lily requests and responses also include *chunks*, which allow clients and servers to send large amounts of data in chunks, improving memory usage.

Internally, the Lily server keeps a tree of each directory and file of every drive. Each directory and file keeps a lock in order to synchronize accessing and modifying the directories and files' settings, along with the directories and files themselves. The server also keeps a list of sessions and a list of users, which are all updated at runtime. Lastly, the server runs a cron job which verifies that the stored tree is valid, and updates the drive and server files if they are changed. If a Lily command updates anything within the server or its drives, it will mark the server or the drive as "dirty", meaning that it needs to be saved on the next cron cycle.