# Response Codes
Responses in Lily consist of a response code and a response string, an optional string containing an error message.

### **Code:** 0

**Description:**
> Operation completed successfully. Returns an empty response string.

### **Code:** 1

**Description:**
> Invalid command ID.

### **Code:** 2

**Description:**
> Unhandled command error.

### **Code:** 3

**Description:**
> Invalid request.

### **Code:** 4

**Description:** 
> Connection timed out or connection error.

### **Code:** 5

**Description:** 
> Invalid protocol version.

### **Code:** 6

**Description:**
> Invalid or expired authentication.

### **Code:** 7

**Description:**
> Rate limit reached. Please try again later.

### **Code:** 8

**Description:**
> Out of memory. Please try again later.

### **Code:** 9

**Description:**
> Server could not successfully generate a unique session ID. Please try again.

### **Code:** 10

**Description:**
> Invalid expiration time. Server does not allow non-expiring sessions.

### **Code:** 11

**Description:**
> Per-user session limit reached.

### **Code:** 8

**Description:**
> Insufficient clearance.

### **Code:** 9

**Description:**
> User does not exist.

### **Code:** 10

**Description:**
> Invalid lengths of arguments. Generally occurs when the arguments for a command are multiple arrays that must all have the same length.

### **Code:** 11

**Description:**
> Incompatible server and client versions.

### **Code:** 12

**Description:**
> User already exists.

### **Code:** 13

**Description:**
> Session does not exist.

### **Code:** 14

**Description:**
> Drive does not exist.

### **Code:** 15

**Description:**
> Drive already exists.

### **Code:** 16

**Description:**
> Invalid number of workers.

### **Code:** 17

**Description:**
> Invalid timeout interval.

### **Code:** 18

**Description:**
> Invalid log file path.

### **Code:** 19

**Description:**
> Invalid log level.

### **Code:** 20

**Description:**
> Invalid rate limit.

### **Code:** 21

**Description:**
> Invalid server file path.

### **Code:** 22

**Description:**
> Invalid host and port.

### **Code:** 23

**Description:** 
> Invalid drive file path.