// cmd/config.go
// Configuration commands.

package cmd

import (
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cubeflix/lily/drive"
	"github.com/cubeflix/lily/marshal"
	"github.com/cubeflix/lily/security/access"
	"github.com/cubeflix/lily/server"
	"github.com/cubeflix/lily/server/config"
	"github.com/cubeflix/lily/user"
	ulist "github.com/cubeflix/lily/user/list"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

// Config init command.
func ConfigInit(cmd *cobra.Command, args []string) {
	// Load the config file.
	cfg, err := ini.Load(args[0])
	if err != nil {
		fmt.Println("config:", err)
		return
	}

	// Get drives.
	driveFiles := map[string]string{}
	drivesSec := cfg.Section("drives")
	if len(drivesSec.Keys()) == 0 {
		fmt.Println("config: no drives")
	}
	for i := range drivesSec.Keys() {
		name := drivesSec.Keys()[i].Name()
		value := drivesSec.Keys()[i].String()
		if !filepath.IsAbs(value) {
			fmt.Println("config: drive path must be absolute")
			return
		}
		driveFiles[name] = value
	}

	// Get certificates.
	certFiles := strings.Split(cfg.Section("certs").Key("certFiles").String(), ",")
	keyFiles := strings.Split(cfg.Section("certs").Key("keyFiles").String(), ",")
	if len(certFiles) != len(keyFiles) {
		fmt.Println("config: cert files and key files must be the same length")
		return
	}
	if len(certFiles) == 0 {
		fmt.Println("config: no certificates")
		return
	}
	certs := make([]config.CertFilePair, len(certFiles))
	for i := range certFiles {
		if !filepath.IsAbs(certFiles[i]) {
			fmt.Println("config: cert file path must be absolute")
			return
		}
		if !filepath.IsAbs(keyFiles[i]) {
			fmt.Println("config: key file path must be absolute")
			return
		}
		certs[i] = config.CertFilePair{Cert: certFiles[i], Key: keyFiles[i]}
	}

	// Get TLS config.
	configSec := cfg.Section("config")
	tlsConfig := &tls.Config{}
	tlsConfig.InsecureSkipVerify = configSec.Key("insecureSkipVerify").MustBool(false)
	if configSec.Key("tlsServerName").String() != "" {
		tlsConfig.ServerName = configSec.Key("tlsServerName").String()
	}

	// Get admin user.
	adminSec := cfg.Section("admin")
	username := adminSec.Key("username").MustString("admin")
	password := adminSec.Key("password").MustString("admin")

	// Create the new config.
	serverFile, err = filepath.Abs(serverFile)
	if err != nil {
		fmt.Println("config:", err.Error())
		return
	}
	c, err := config.NewConfig(serverFile,
		cfg.Section("config").Key("name").MustString("lily"),
		cfg.Section("config").Key("host").MustString("127.0.0.1"),
		cfg.Section("config").Key("port").MustInt(42069),
		driveFiles,
		cfg.Section("config").Key("workers").MustInt(5),
		cfg.Section("config").Key("backlog").MustInt(5),
		cfg.Section("config").Key("cronInterval").MustDuration(time.Second*60),
		cfg.Section("config").Key("sessionInterval").MustDuration(time.Second*5),
		cfg.Section("config").Key("timeout").MustDuration(time.Second*10),
		cfg.Section("config").Key("verbose").MustBool(true),
		cfg.Section("config").Key("logToFile").MustBool(false),
		cfg.Section("config").Key("logJSON").MustBool(false),
		cfg.Section("config").Key("logLevel").MustString("info"),
		cfg.Section("config").Key("logPath").String(),
		cfg.Section("config").Key("defaultSessionExpiration").MustDuration(time.Hour),
		cfg.Section("config").Key("allowChangeSessionExpiration").MustBool(true),
		cfg.Section("config").Key("allowNonExpiringSessions").MustBool(true),
		cfg.Section("config").Key("perUserSessionLimit").MustInt(10),
		cfg.Section("config").Key("rateLimitInterval").MustDuration(time.Second),
		cfg.Section("config").Key("maxLimitEvents").MustInt(10),
		certs,
		tlsConfig)
	if err != nil {
		fmt.Println("config:", err.Error())
		return
	}

	uobj, err := user.NewUser(username, password, access.ClearanceLevelFive)
	if err != nil {
		fmt.Println("config:", err.Error())
		return
	}
	userlist := ulist.NewUserList()
	userlist.SetUsersByName(map[string]*user.User{"admin": uobj})

	// Save the file.
	o, err := os.Create(serverFile)
	if err != nil {
		fmt.Println("config:", err.Error())
		return
	}
	err = marshal.MarshalConfig(c, o)
	if err != nil {
		o.Close()
		fmt.Println("config:", err.Error())
		return
	}
	err = marshal.MarshalUserList(userlist, o)
	if err != nil {
		o.Close()
		fmt.Println("config:", err.Error())
		return
	}
	o.Close()
}

// Config set command.
func ConfigSet(cmd *cobra.Command, args []string) {
	// Load the server file.
	s, err := server.LoadServerFromFile(serverFile)
	if err != nil {
		fmt.Println("config:", err.Error())
		return
	}
	name := args[0]
	if name == "file" {
		if !filepath.IsAbs(args[1]) {
			fmt.Println("config: file path must be absolute")
			return
		}
		s.Config().SetServerFile(args[1])
	} else if name == "name" {
		s.Config().SetName(args[1])
	} else if name == "host" {
		_, port := s.Config().GetHostAndPort()
		s.Config().SetHostAndPort(args[1], port)
	} else if name == "port" {
		host, _ := s.Config().GetHostAndPort()
		port, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("config:", err.Error())
			return
		}
		s.Config().SetHostAndPort(host, port)
	} else if name == "workers" {
		workers, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("config:", err.Error())
			return
		}
		s.Config().SetNumWorkers(workers)
	} else if name == "backlog" {
		backlog, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("config:", err.Error())
			return
		}
		s.Config().SetBacklog(backlog)
	} else if name == "cronInterval" {
		_, sessionInterval := s.Config().GetCronIntervals()
		cronInterval, err := time.ParseDuration(args[1])
		if err != nil {
			fmt.Println("config:", err.Error())
			return
		}
		s.Config().SetCronIntervals(cronInterval, sessionInterval)
	} else if name == "sessionInterval" {
		cronInterval, _ := s.Config().GetCronIntervals()
		sessionInterval, err := time.ParseDuration(args[1])
		if err != nil {
			fmt.Println("config:", err.Error())
			return
		}
		s.Config().SetCronIntervals(cronInterval, sessionInterval)
	} else if name == "timeout" {
		timeout, err := time.ParseDuration(args[1])
		if err != nil {
			fmt.Println("config:", err.Error())
			return
		}
		s.Config().SetTimeout(timeout)
	} else if name == "verbose" {
		_, logToFile, logJSON, logLevel, logPath := s.Config().GetLogging()
		verbose, err := strconv.ParseBool(args[1])
		if err != nil {
			fmt.Println("config:", err.Error())
			return
		}
		s.Config().SetLogging(verbose, logToFile, logJSON, logLevel, logPath)
	} else if name == "logToFile" {
		verbose, _, logJSON, logLevel, logPath := s.Config().GetLogging()
		logToFile, err := strconv.ParseBool(args[1])
		if err != nil {
			fmt.Println("config:", err.Error())
			return
		}
		s.Config().SetLogging(verbose, logToFile, logJSON, logLevel, logPath)
	} else if name == "logJSON" {
		verbose, logToFile, _, logLevel, logPath := s.Config().GetLogging()
		logJSON, err := strconv.ParseBool(args[1])
		if err != nil {
			fmt.Println("config:", err.Error())
			return
		}
		s.Config().SetLogging(verbose, logToFile, logJSON, logLevel, logPath)
	} else if name == "logLevel" {
		verbose, logToFile, logJSON, _, logPath := s.Config().GetLogging()
		logLevel := args[1]
		if err := s.Config().SetLogging(verbose, logToFile, logJSON, logLevel, logPath); err != nil {
			fmt.Println("config:", err.Error())
			return
		}
	} else if name == "logPath" {
		verbose, logToFile, logJSON, logLevel, _ := s.Config().GetLogging()
		logPath := args[1]
		s.Config().SetLogging(verbose, logToFile, logJSON, logLevel, logPath)
	} else if name == "defaultSessionExpiration" {
		_, allowChangeSessionExpiration, allowNonExpiringSessions := s.Config().GetSessionExpirationSettings()
		defaultSessionExpiration, err := time.ParseDuration(args[1])
		if err != nil {
			fmt.Println("config:", err.Error())
			return
		}
		s.Config().SetSessionExpirationSettings(defaultSessionExpiration, allowChangeSessionExpiration, allowNonExpiringSessions)
	} else if name == "allowChangeSessionExpiration" {
		defaultSessionExpiration, _, allowNonExpiringSessions := s.Config().GetSessionExpirationSettings()
		allowChangeSessionExpiration, err := strconv.ParseBool(args[1])
		if err != nil {
			fmt.Println("config:", err.Error())
			return
		}
		s.Config().SetSessionExpirationSettings(defaultSessionExpiration, allowChangeSessionExpiration, allowNonExpiringSessions)
	} else if name == "allowNonExpiringSessions" {
		defaultSessionExpiration, allowChangeSessionExpiration, _ := s.Config().GetSessionExpirationSettings()
		allowNonExpiringSessions, err := strconv.ParseBool(args[1])
		if err != nil {
			fmt.Println("config:", err.Error())
			return
		}
		s.Config().SetSessionExpirationSettings(defaultSessionExpiration, allowChangeSessionExpiration, allowNonExpiringSessions)
	} else if name == "perUserSessionLimit" {
		perUserSessionLimit, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("config:", err.Error())
			return
		}
		s.Config().SetUserSessionLimit(perUserSessionLimit)
	} else if name == "rateLimitInterval" {
		_, maxLimitEvents := s.Config().GetRateLimit()
		rateLimitInterval, err := time.ParseDuration(args[1])
		if err != nil {
			fmt.Println("config:", err.Error())
			return
		}
		s.Config().SetRateLimit(rateLimitInterval, maxLimitEvents)
	} else if name == "maxLimitEvents" {
		rateLimitInterval, _ := s.Config().GetRateLimit()
		maxLimitEvents, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("config:", err.Error())
			return
		}
		s.Config().SetRateLimit(rateLimitInterval, maxLimitEvents)
	} else if name == "insecureSkipVerify" {
		insecureSkipVerify, err := strconv.ParseBool(args[1])
		if err != nil {
			fmt.Println("config:", err.Error())
			return
		}
		tlsConfig := s.Config().GetTLSConfig()
		tlsConfig.InsecureSkipVerify = insecureSkipVerify
		s.Config().SetTLSConfig(tlsConfig)
	} else if name == "tlsServerName" {
		tlsServerName := args[1]
		tlsConfig := s.Config().GetTLSConfig()
		tlsConfig.ServerName = tlsServerName
		s.Config().SetTLSConfig(tlsConfig)
	} else {
		fmt.Println("config: invalid setting name")
		return
	}
	file, err := os.OpenFile(serverFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("config:", err.Error())
		return
	}
	err = marshal.MarshalConfig(s.Config(), file)
	if err != nil {
		file.Close()
		fmt.Println("config:", err.Error())
		return
	}
	err = marshal.MarshalUserList(s.Users(), file)
	if err != nil {
		file.Close()
		fmt.Println("config:", err.Error())
		return
	}
	file.Close()
}

// Config get command.
func ConfigGet(cmd *cobra.Command, args []string) {
	// Load the server file.
	s, err := server.LoadServerFromFile(serverFile)
	if err != nil {
		fmt.Println("config:", err.Error())
		return
	}
	name := args[0]
	if name == "file" {
		fmt.Println(s.Config().GetServerFile())
	} else if name == "name" {
		fmt.Println(s.Config().GetName())
	} else if name == "host" {
		host, _ := s.Config().GetHostAndPort()
		fmt.Println(host)
	} else if name == "port" {
		_, port := s.Config().GetHostAndPort()
		fmt.Println(port)
	} else if name == "workers" {
		fmt.Println(s.Config().GetNumWorkers())
	} else if name == "backlog" {
		fmt.Println(s.Config().GetBacklog())
	} else if name == "cronInterval" {
		cronInterval, _ := s.Config().GetCronIntervals()
		fmt.Println(cronInterval)
	} else if name == "sessionInterval" {
		_, sessionInterval := s.Config().GetCronIntervals()
		fmt.Println(sessionInterval)
	} else if name == "timeout" {
		fmt.Println(s.Config().GetTimeout())
	} else if name == "verbose" {
		verbose, _, _, _, _ := s.Config().GetLogging()
		fmt.Println(verbose)
	} else if name == "logToFile" {
		_, logToFile, _, _, _ := s.Config().GetLogging()
		fmt.Println(logToFile)
	} else if name == "logJSON" {
		_, _, logJSON, _, _ := s.Config().GetLogging()
		fmt.Println(logJSON)
	} else if name == "logLevel" {
		_, _, _, logLevel, _ := s.Config().GetLogging()
		fmt.Println(logLevel)
	} else if name == "logPath" {
		_, _, _, _, logPath := s.Config().GetLogging()
		fmt.Println(logPath)
	} else if name == "defaultSessionExpiration" {
		defaultSessionExpiration, _, _ := s.Config().GetSessionExpirationSettings()
		fmt.Println(defaultSessionExpiration)
	} else if name == "allowChangeSessionExpiration" {
		_, allowChangeSessionExpiration, _ := s.Config().GetSessionExpirationSettings()
		fmt.Println(allowChangeSessionExpiration)
	} else if name == "allowNonExpiringSessions" {
		_, _, allowNonExpiringSessions := s.Config().GetSessionExpirationSettings()
		fmt.Println(allowNonExpiringSessions)
	} else if name == "perUserSessionLimit" {
		fmt.Println(s.Config().GetUserSessionLimit())
	} else if name == "rateLimitInterval" {
		rateLimitInterval, _ := s.Config().GetRateLimit()
		fmt.Println(rateLimitInterval)
	} else if name == "maxLimitEvents" {
		_, maxLimitEvents := s.Config().GetRateLimit()
		fmt.Println(maxLimitEvents)
	} else if name == "insecureSkipVerify" {
		fmt.Println(s.Config().GetTLSConfig().InsecureSkipVerify)
	} else if name == "tlsServerName" {
		fmt.Println(s.Config().GetTLSConfig().ServerName)
	} else {
		fmt.Println("config: invalid setting name")
		return
	}
}

// Config list command.
func ConfigList(cmd *cobra.Command, args []string) {
	// Load the server file.
	s, err := server.LoadServerFromFile(serverFile)
	if err != nil {
		fmt.Println("config:", err.Error())
		return
	}
	fmt.Println("server file:", s.Config().GetServerFile())
	fmt.Println("name:", s.Config().GetName())
	host, port := s.Config().GetHostAndPort()
	fmt.Printf("addr: %s:%d\n", host, port)
	fmt.Println("workers:", s.Config().GetNumWorkers())
	fmt.Println("backlog:", s.Config().GetBacklog())
	ci, si := s.Config().GetCronIntervals()
	fmt.Println("cron interval:", ci)
	fmt.Println("session interval:", si)
	fmt.Println("timeout:", s.Config().GetTimeout())
	verbose, logToFile, logJSON, logLevel, logFile := s.Config().GetLogging()
	fmt.Println("verbose:", verbose)
	fmt.Println("log to file:", logToFile)
	fmt.Println("log JSON:", logJSON)
	fmt.Println("log level:", logLevel)
	fmt.Println("log file:", logFile)
	defaultSessionExpiration, allowChangeSessionExpiration, allowNonExpiringSessions := s.Config().GetSessionExpirationSettings()
	fmt.Println("default session expiration:", defaultSessionExpiration)
	fmt.Println("allow change session expiraiton:", allowChangeSessionExpiration)
	fmt.Println("allow non expiring sessions:", allowNonExpiringSessions)
	fmt.Println("per user session limit:", s.Config().GetUserSessionLimit())
	limit, maxLimitEvents := s.Config().GetRateLimit()
	fmt.Println("rate limit interval:", limit)
	fmt.Println("max rate limit events:", maxLimitEvents)
	fmt.Println("certificates:")
	certs := s.Config().GetCertFilePairs()
	for i := range certs {
		fmt.Println("	cert:", certs[i].Cert, "key:", certs[i].Key)
	}
}

// Add a drive file.
func ConfigAddDrive(cmd *cobra.Command, args []string) {
	// Load the server file.
	s, err := server.LoadServerFromFile(serverFile)
	if err != nil {
		fmt.Println("config:", err.Error())
		return
	}

	if !filepath.IsAbs(args[1]) {
		fmt.Println("config: drive path must be absolute")
		return
	}
	err = s.Config().AddDriveFiles(map[string]string{args[0]: args[1]})
	if err != nil {
		fmt.Println("config:", err.Error())
		return
	}
	err = s.LoadDrives()
	if err != nil {
		fmt.Println("config:", err.Error())
		return
	}

	// Save the server file.
	file, err := os.OpenFile(serverFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("config:", err.Error())
		return
	}
	err = marshal.MarshalConfig(s.Config(), file)
	if err != nil {
		file.Close()
		fmt.Println("config:", err.Error())
		return
	}
	err = marshal.MarshalUserList(s.Users(), file)
	if err != nil {
		file.Close()
		fmt.Println("config:", err.Error())
		return
	}
	file.Close()
}

// Rename a drive.
func ConfigRenameDrive(cmd *cobra.Command, args []string) {
	// Load the server file.
	s, err := server.LoadServerFromFile(serverFile)
	if err != nil {
		fmt.Println("config:", err.Error())
		return
	}

	driveFiles := s.Config().GetDriveFiles()
	if err != nil {
		fmt.Println("config:", err.Error())
		return
	}
	driveFile, ok := driveFiles[args[0]]
	if !ok {
		fmt.Println("config: drive does not exist")
		return
	}

	f, err := os.Open(driveFile)
	if err != nil {
		fmt.Println("drive:", err.Error())
		return
	}
	d, err := drive.Unmarshal(f)
	if err != nil {
		f.Close()
		fmt.Println("drive:", err.Error())
		return
	}
	f.Close()

	d.SetName(args[1])

	f, err = os.Create(driveFile)
	if err != nil {
		fmt.Println("drive:", err.Error())
		return
	}
	err = d.Marshal(f)
	if err != nil {
		fmt.Println("drive:", err.Error())
		f.Close()
		return
	}
	f.Close()

	err = s.Config().RemoveDriveFiles([]string{args[0]})
	if err != nil {
		fmt.Println("drive:", err.Error())
		return
	}

	err = s.Config().AddDriveFiles(map[string]string{args[1]: driveFile})
	if err != nil {
		fmt.Println("config:", err.Error())
		return
	}

	// Save the server file.
	file, err := os.OpenFile(serverFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("config:", err.Error())
		return
	}
	err = marshal.MarshalConfig(s.Config(), file)
	if err != nil {
		file.Close()
		fmt.Println("config:", err.Error())
		return
	}
	err = marshal.MarshalUserList(s.Users(), file)
	if err != nil {
		file.Close()
		fmt.Println("config:", err.Error())
		return
	}
	file.Close()
}

// Remove a drive file.
func ConfigRemoveDrive(cmd *cobra.Command, args []string) {
	// Load the server file.
	s, err := server.LoadServerFromFile(serverFile)
	if err != nil {
		fmt.Println("config:", err.Error())
		return
	}

	err = s.Config().RemoveDriveFiles([]string{args[0]})
	if err != nil {
		fmt.Println("config:", err.Error())
		return
	}

	// Save the server file.
	file, err := os.OpenFile(serverFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("config:", err.Error())
		return
	}
	err = marshal.MarshalConfig(s.Config(), file)
	if err != nil {
		file.Close()
		fmt.Println("config:", err.Error())
		return
	}
	err = marshal.MarshalUserList(s.Users(), file)
	if err != nil {
		file.Close()
		fmt.Println("config:", err.Error())
		return
	}
	file.Close()
}

// Add a user.
func ConfigAddUser(cmd *cobra.Command, args []string) {
	// Load the server file.
	s, err := server.LoadServerFromFile(serverFile)
	if err != nil {
		fmt.Println("config:", err.Error())
		return
	}

	c := access.Clearance(clearance)
	err = c.Validate()
	if err != nil {
		fmt.Println("config:", err.Error())
		return
	}
	uobj, err := user.NewUser(args[0], args[1], c)
	if err != nil {
		fmt.Println("config:", err.Error())
		return
	}
	s.Users().SetUsersByName(map[string]*user.User{args[0]: uobj})

	// Save the server file.
	file, err := os.OpenFile(serverFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("config:", err.Error())
		return
	}
	err = marshal.MarshalConfig(s.Config(), file)
	if err != nil {
		file.Close()
		fmt.Println("config:", err.Error())
		return
	}
	err = marshal.MarshalUserList(s.Users(), file)
	if err != nil {
		file.Close()
		fmt.Println("config:", err.Error())
		return
	}
	file.Close()
}

// Remove a user.
func ConfigRemoveUser(cmd *cobra.Command, args []string) {
	// Load the server file.
	s, err := server.LoadServerFromFile(serverFile)
	if err != nil {
		fmt.Println("config:", err.Error())
		return
	}

	err = s.Users().RemoveUsersByName([]string{args[0]})
	if err != nil {
		fmt.Println("config:", err.Error())
		return
	}

	// Save the server file.
	file, err := os.OpenFile(serverFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("config:", err.Error())
		return
	}
	err = marshal.MarshalConfig(s.Config(), file)
	if err != nil {
		file.Close()
		fmt.Println("config:", err.Error())
		return
	}
	err = marshal.MarshalUserList(s.Users(), file)
	if err != nil {
		file.Close()
		fmt.Println("config:", err.Error())
		return
	}
	file.Close()
}

// Config drive list command.
func ConfigListDrive(cmd *cobra.Command, args []string) {
	// Load the server file.
	s, err := server.LoadServerFromFile(serverFile)
	if err != nil {
		fmt.Println("config:", err.Error())
		return
	}

	fmt.Println("drives:")
	drives := s.Config().GetDriveFiles()
	for name := range drives {
		fmt.Println("	"+name+":", drives[name])
	}
}

// Config set certs command.
func ConfigSetCerts(cmd *cobra.Command, args []string) {
	// Load the server file.
	s, err := server.LoadServerFromFile(serverFile)
	if err != nil {
		fmt.Println("config:", err.Error())
		return
	}

	certFiles := strings.Split(args[0], ",")
	keyFiles := strings.Split(args[1], ",")
	if len(certFiles) != len(keyFiles) {
		fmt.Println("config: cert files and key files must be the same length")
		return
	}
	if len(certFiles) == 0 {
		fmt.Println("config: no certificates")
		return
	}
	certs := make([]config.CertFilePair, len(certFiles))
	for i := range certFiles {
		if !filepath.IsAbs(certFiles[i]) {
			fmt.Println("config: cert file must be absolute")
			return
		}
		if !filepath.IsAbs(keyFiles[i]) {
			fmt.Println("config: key file must be absolute")
			return
		}
		certs[i] = config.CertFilePair{Cert: certFiles[i], Key: keyFiles[i]}
	}
	s.Config().SetCertFilePairs(certs)

	// Save the server file.
	file, err := os.OpenFile(serverFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("config:", err.Error())
		return
	}
	err = marshal.MarshalConfig(s.Config(), file)
	if err != nil {
		file.Close()
		fmt.Println("config:", err.Error())
		return
	}
	err = marshal.MarshalUserList(s.Users(), file)
	if err != nil {
		file.Close()
		fmt.Println("config:", err.Error())
		return
	}
	file.Close()
}
