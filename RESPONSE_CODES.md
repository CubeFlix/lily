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
> Invalid request.

### **Code:** 3

**Description:**
> Invalid authentication type.

### **Code:** 4

**Description:**
> Invalid or expired authentication.

### **Code:** 5

**Description:**
> Invalid expiration time. Server does not allow non-expiring sessions.

### **Code:** 6

**Description:**
> Server could not successfully generate a unique session ID. Please try again.

### **Code:** 7

**Description:**
> Insufficient clearance.

### **Code:** 8

**Description:**
> User does not exist.

### **Code:** 9

**Description:**
> Invalid lengths of arguments. Generally occurs when the arguments for a command are multiple arrays that must all have the same length.

### **Code:** 10

**Description:**
> Incompatible server and client versions.

### **Code:** 11

**Description:**
> User already exists.

### **Code:** 12

**Description:**
> Session does not exist.
