package sql

import (
	"encoding/json"
	"os/exec"
)

type SQLServer struct {
	Name          string `json:"name"`
	Location      string `json:"location"`
	ResourceGroup string `json:"resourceGroup"`
}

type SQLDatabase struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

func ListSQLServers() ([]SQLServer, error) {
	cmd := exec.Command("az", "sql", "server", "list", "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var servers []SQLServer
	if err := json.Unmarshal(out, &servers); err != nil {
		return nil, err
	}
	return servers, nil
}

func ListSQLDatabases(server, group string) ([]SQLDatabase, error) {
	cmd := exec.Command("az", "sql", "db", "list", "--server", server, "--resource-group", group, "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var dbs []SQLDatabase
	if err := json.Unmarshal(out, &dbs); err != nil {
		return nil, err
	}
	return dbs, nil
}

func CreateSQLServer(name, group, location, adminUser, adminPass string) error {
	return exec.Command("az", "sql", "server", "create", "--name", name, "--resource-group", group, "--location", location, "--admin-user", adminUser, "--admin-password", adminPass).Run()
}

func DeleteSQLServer(name, group string) error {
	return exec.Command("az", "sql", "server", "delete", "--name", name, "--resource-group", group, "--yes").Run()
}
