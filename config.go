// Copyright 2016 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

import (
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Config for database connection.
// Drivers may have additional properties held in KV.
// If a driver is file based, the file name should be in the "Instance" field.
type Config struct {
	// Registered driver name for the database.
	DriverName string

	// Raw is the pre-parsed connection string, possibly directly
	// used by the driver opener.
	Raw string

	Username string
	Password string
	Hostname string
	Port     int
	Instance string
	Database string // Initial database to connect to.

	// Time for an idle connection to be closed.
	// Zero if there should be no timeout.
	PoolIdleTimeout time.Duration

	// How many connection should be created at startup.
	// Valid range is (0 < init, init <= max).
	PoolInitCapacity int

	// Max number of connections to create.
	// Valid range is (0 < max).
	PoolMaxCapacity int

	// Require the driver to establish a secure connection.
	Secure bool

	// Do not require the secure connection to verify the remote host name.
	// Ignored if Secure is false.
	InsecureSkipVerify bool

	KV map[string]interface{}
}

// ParseConfigURL is a standard method to parse configuration options from a text.
// The instance field can also hold the filename in case of a file based connection.
//   driver://[username:password@][url[:port]]/[Instance]?db=mydatabase&opt1=valA&opt2=valB
//   sqlite:///C:/folder/file.sqlite3?opt1=valA&opt2=valB
//   sqlite:///srv/folder/file.sqlite3?opt1=valA&opt2=valB
//   ms://TESTU@localhost/SqlExpress?db=master
// This will attempt to find the driver to load additional parameters.
//   Additional field options:
//      db=<string>:                  Database
//      init_cap=<int>:               PoolInitCapacity
//      max_cap=<int>:                PoolMaxCapacity
//      idle_timeout=<time.Duration>: PoolIdleTimeout
func ParseConfigURL(connectionString string) (*Config, error) {
	u, err := url.Parse(connectionString)
	if err != nil {
		return nil, err
	}
	var user, pass string
	if u.User != nil {
		user = u.User.Username()
		pass, _ = u.User.Password()
	}
	port := 0
	host := ""

	if len(u.Host) > 0 {
		hostPort := strings.Split(u.Host, ":")
		host = hostPort[0]
		if len(hostPort) > 1 {
			parsedPort, err := strconv.ParseUint(hostPort[1], 10, 16)
			if err != nil {
				return nil, err
			}
			port = int(parsedPort)
		}
	}

	conf := &Config{
		DriverName: u.Scheme,
		Raw:        connectionString,
		Username:   user,
		Password:   pass,
		Hostname:   host,
		Port:       port,
	}

	val := u.Query()

	conf.Database = val.Get("db")
	val.Del("db")

	if st := val.Get("idle_timeout"); len(st) != 0 {
		conf.PoolIdleTimeout, err = time.ParseDuration(st)
		if err != nil {
			return nil, err
		}
	}
	val.Del("idle_timeout")

	if st := val.Get("init_cap"); len(st) != 0 {
		conf.PoolInitCapacity, err = strconv.Atoi(st)
		if err != nil {
			return nil, err
		}
	}
	val.Del("init_cap")

	if st := val.Get("max_cap"); len(st) != 0 {
		conf.PoolMaxCapacity, err = strconv.Atoi(st)
		if err != nil {
			return nil, err
		}
	}
	val.Del("max_cap")

	if len(u.Path) > 0 {
		conf.Instance = u.Path[1:]
	}

	for key, value := range val {
		conf.KV[key] = value
	}
	return conf, nil
}
