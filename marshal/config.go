// marshal/config.go
// Marshaling for config objects.

package marshal

import (
	"crypto/tls"
	"encoding/binary"
	"io"
	"time"

	"github.com/cubeflix/lily/server/config"
)

// Marshal a config object.
func MarshalConfig(c *config.Config, w io.Writer) error {
	// Write the config.
	err := MarshalString(c.GetServerFile(), w)
	if err != nil {
		return err
	}
	err = MarshalString(c.GetName(), w)
	if err != nil {
		return err
	}
	host, port := c.GetHostAndPort()
	err = MarshalString(host, w)
	if err != nil {
		return err
	}
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(port))
	_, err = w.Write(data)
	if err != nil {
		return err
	}

	// Write the drive files.
	err = MarshalMapStringString(c.GetDriveFiles(), w)
	if err != nil {
		return err
	}

	// Write more config.
	binary.LittleEndian.PutUint32(data, uint32(c.GetNumWorkers()))
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	binary.LittleEndian.PutUint32(data, uint32(c.GetBacklog()))
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	optDaemons, optArgs := c.GetOptionalDaemons()
	err = MarshalStringSlice(optDaemons, w)
	if err != nil {
		return err
	}
	err = MarshalMapStringString(optArgs, w)
	if err != nil {
		return err
	}

	// Marshal the durations.
	data = make([]byte, 8)
	mint, sint := c.GetCronIntervals()
	binary.LittleEndian.PutUint64(data, uint64(mint))
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	binary.LittleEndian.PutUint64(data, uint64(sint))
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	binary.LittleEndian.PutUint64(data, uint64(c.GetTimeout()))
	_, err = w.Write(data)
	if err != nil {
		return err
	}

	// Marshal the logging settings.
	verbose, logToFile, logJSON, logLevel, logPath := c.GetLogging()
	err = MarshalBool(verbose, w)
	if err != nil {
		return err
	}
	err = MarshalBool(logToFile, w)
	if err != nil {
		return err
	}
	err = MarshalBool(logJSON, w)
	if err != nil {
		return err
	}
	err = MarshalString(logLevel, w)
	if err != nil {
		return err
	}
	err = MarshalString(logPath, w)
	if err != nil {
		return err
	}

	// Marshal the session expiration settings.
	defaultSessionExpiration, allowChangeSessionExpiration, allowNonExpiringSessions := c.GetSessionExpirationSettings()
	binary.LittleEndian.PutUint64(data, uint64(defaultSessionExpiration))
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	err = MarshalBool(allowChangeSessionExpiration, w)
	if err != nil {
		return err
	}
	err = MarshalBool(allowNonExpiringSessions, w)
	if err != nil {
		return err
	}

	// Marshal the config.
	data = make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(c.GetUserSessionLimit()))
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	data = make([]byte, 8)
	limit, maxLimitEvents := c.GetRateLimit()
	binary.LittleEndian.PutUint64(data, uint64(limit))
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	data = make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(maxLimitEvents))
	_, err = w.Write(data)
	if err != nil {
		return err
	}

	// Marshal the cert files.
	err = MarshalCertFilePair(c.GetCertFilePairs(), w)
	if err != nil {
		return err
	}

	// Marshal the server name and InsecureSkipVerify.
	tlsConfig := c.GetTLSConfig()
	err = MarshalString(tlsConfig.ServerName, w)
	if err != nil {
		return err
	}
	err = MarshalBool(tlsConfig.InsecureSkipVerify, w)
	if err != nil {
		return err
	}

	// Return.
	return nil
}

// Unmarshal access settings.
func UnmarshalConfig(r io.Reader) (*config.Config, error) {
	// Receive the file.
	file, err := UnmarshalString(r)
	if err != nil {
		return nil, err
	}
	name, err := UnmarshalString(r)
	if err != nil {
		return nil, err
	}
	host, err := UnmarshalString(r)
	if err != nil {
		return nil, err
	}
	data := make([]byte, 4)
	_, err = r.Read(data)
	if err != nil {
		return nil, err
	}
	port := binary.LittleEndian.Uint32(data)
	driveFiles, err := UnmarshalMapStringString(r)
	if err != nil {
		return nil, err
	}
	_, err = r.Read(data)
	if err != nil {
		return nil, err
	}
	numWorkers := binary.LittleEndian.Uint32(data)
	_, err = r.Read(data)
	if err != nil {
		return nil, err
	}
	backlog := binary.LittleEndian.Uint32(data)
	optionalDaemons, err := UnmarshalStringSlice(r)
	if err != nil {
		return nil, err
	}
	optionalArgs, err := UnmarshalMapStringString(r)
	if err != nil {
		return nil, err
	}
	data = make([]byte, 8)
	_, err = r.Read(data)
	if err != nil {
		return nil, err
	}
	mainCronInterval := time.Duration(binary.LittleEndian.Uint64(data))
	_, err = r.Read(data)
	if err != nil {
		return nil, err
	}
	sessionCronInterval := time.Duration(binary.LittleEndian.Uint64(data))
	_, err = r.Read(data)
	if err != nil {
		return nil, err
	}
	netTimeout := time.Duration(binary.LittleEndian.Uint64(data))
	verbose, err := UnmarshalBool(r)
	if err != nil {
		return nil, err
	}
	logToFile, err := UnmarshalBool(r)
	if err != nil {
		return nil, err
	}
	logJSON, err := UnmarshalBool(r)
	if err != nil {
		return nil, err
	}
	logLevel, err := UnmarshalString(r)
	if err != nil {
		return nil, err
	}
	logPath, err := UnmarshalString(r)
	if err != nil {
		return nil, err
	}
	_, err = r.Read(data)
	if err != nil {
		return nil, err
	}
	defaultSessionExpiration := time.Duration(binary.LittleEndian.Uint64(data))
	allowChangeSessionExpiration, err := UnmarshalBool(r)
	if err != nil {
		return nil, err
	}
	allowNonExpiringSessions, err := UnmarshalBool(r)
	if err != nil {
		return nil, err
	}
	data = make([]byte, 4)
	_, err = r.Read(data)
	if err != nil {
		return nil, err
	}
	perUserSessionLimit := binary.LittleEndian.Uint32(data)
	data = make([]byte, 8)
	_, err = r.Read(data)
	if err != nil {
		return nil, err
	}
	limit := time.Duration(binary.LittleEndian.Uint64(data))
	data = make([]byte, 4)
	_, err = r.Read(data)
	if err != nil {
		return nil, err
	}
	maxLimitEvents := binary.LittleEndian.Uint32(data)
	certFiles, err := UnmarshalCertFilePair(r)
	if err != nil {
		return nil, err
	}
	serverName, err := UnmarshalString(r)
	if err != nil {
		return nil, err
	}
	InsecureSkipVerify, err := UnmarshalBool(r)
	if err != nil {
		return nil, err
	}

	// Create the new config object.
	return config.NewConfig(file, name, host, int(port), driveFiles, int(numWorkers),
		int(backlog), optionalDaemons, optionalArgs, mainCronInterval, sessionCronInterval,
		netTimeout, verbose, logToFile, logJSON, logLevel, logPath,
		defaultSessionExpiration, allowChangeSessionExpiration,
		allowNonExpiringSessions, int(perUserSessionLimit), limit, int(maxLimitEvents),
		certFiles, &tls.Config{ServerName: serverName,
			InsecureSkipVerify: InsecureSkipVerify})
}
