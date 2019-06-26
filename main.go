package main

// DB-APEX-Import

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/juju/loggo"
	"github.com/spf13/viper"
	_ "gopkg.in/goracle.v2"
)

// TConfig - parameters in config file
type TConfig struct {
	configFile       string
	debugMode        bool
	script           string
	connectionConfig string
	connectionsDir   string
	appID            string
	alias            string
	schema           string
	workspace        string
}

// TConnection - parameters passed by the user
type TConnection struct {
	dbConnectionString string
	username           string
	password           string
	hostname           string
	port               int
	service            string
}

var config = new(TConfig)
var connection TConnection
var programName = os.Args[0]

var logger = loggo.GetLogger(programName)

/********************************************************************************/
func setDebug(debugMode bool) {
	if debugMode == true {

		loggo.ConfigureLoggers(fmt.Sprintf("%s=DEBUG", programName))
		//loggo.ConfigureLoggers("dbcsvdump=DEBUG")
		logger.Debugf("Debug log enabled")
	}
}

/********************************************************************************/
func parseFlags() {

	flag.StringVar(&config.configFile, "configFile", "config", "Configuration file for general parameters")
	flag.StringVar(&config.script, "i", "", "Script File to Import")

	flag.BoolVar(&config.debugMode, "debug", false, "Debug mode (default=false)")
	flag.StringVar(&config.connectionConfig, "connection", "", "Configuration file for connection")

	flag.StringVar(&config.appID, "appID", "", "application ID to import into")
	flag.StringVar(&config.alias, "alias", "", "application alias (override)")
	flag.StringVar(&config.schema, "schema", "", "schema to import into (override)")
	flag.StringVar(&config.workspace, "workspace", "", "workspace to import into (override)")

	flag.StringVar(&connection.dbConnectionString, "db", "", "Database Connection, e.g. user/password@host:port/sid")

	flag.Parse()

	// At a minimum we either need a dbConnection or a configFile
	if (config.configFile == "") && (connection.dbConnectionString == "") {
		flag.PrintDefaults()
		os.Exit(1)
	}

}

/********************************************************************************/
// To execute, at a minimum we need (connection && (object || sql))
func checkMinFlags() {
	// connection is required
	bHaveConnection := (getConnectionString(connection) != "")

	// check if we have either an object to export or a SQL statement
	bHaveScript := (config.script != "")

	if !bHaveConnection || !bHaveScript {
		fmt.Printf("%s:\n", os.Args[0])
	}

	if !bHaveConnection {
		fmt.Printf("  requires a DB connection to be specified\n")
	}

	if !bHaveScript {
		fmt.Printf("requires a script to be specified\n")
	}

	if !bHaveConnection || !bHaveScript {
		flag.PrintDefaults()
		os.Exit(1)
	}
}

/********************************************************************************/
func getConnectionString(connection TConnection) string {

	if connection.dbConnectionString != "" {
		return connection.dbConnectionString
	}

	var str = fmt.Sprintf("%s/%s@%s:%d/%s", connection.username,
		connection.password,
		connection.hostname,
		connection.port,
		connection.service)

	return str
}

/********************************************************************************/
func loadConfig(configFile string) {
	if config.configFile == "" {
		return
	}

	logger.Debugf("reading configFile: %s", configFile)
	viper.SetConfigType("yaml")
	viper.SetConfigName(configFile)
	viper.AddConfigPath(".")

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	// need to set debug mode if it's not already set
	setDebug(viper.GetBool("debugMode"))

	config.connectionsDir = viper.GetString("connectionsDir")
	config.connectionConfig = viper.GetString("connectionConfig")

	config.debugMode = viper.GetBool("debugMode")

	config.configFile = configFile
}

/********************************************************************************/
func loadConnection(connectionFile string) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigName(config.connectionConfig)
	v.AddConfigPath(config.connectionsDir)

	err := v.ReadInConfig() // Find and read the config file
	if err != nil {         // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	if v.GetString("dbConnectionString") != "" {
		connection.dbConnectionString = v.GetString("dbConnectionString")
	}
	connection.username = v.GetString("username")
	connection.password = v.GetString("password")
	connection.hostname = v.GetString("hostname")
	connection.port = v.GetInt("port")
	connection.service = v.GetString("service")

	/*if (viper.GetString("export") != "") && (config.export == "") {
		logger.Debugf("loadConnection: export loaded: %s\n", v.GetString("export"))
		config.export = v.GetString("export")
	}*/

}

/********************************************************************************/
func execScript(tx *sql.Tx, db *sql.DB, scriptPath string) {
	file, err := ioutil.ReadFile(scriptPath)

	if err != nil {
		// handle error
	}

	lImportSQL := `declare
	                 v_workspace_id number;
					 l_workspace_id number;
					 l_app_id       number;
					 l_alias        varchar2(100);
					 l_schema       varchar2(30);
					 l_workspace    varchar2(100);
				   begin
					 l_app_id := :1;
					 l_alias  := :2;
					 l_schema := :3;
					 l_workspace := :4;
					 					 
                     select workspace_id into l_workspace_id
	                   from apex_workspaces
					  where workspace = upper(nvl(l_workspace, USER));
					 						
					apex_application_install.set_workspace_id( l_workspace_id );  					 

					 if (l_app_id is not null) then
					   apex_application_install.set_application_id( l_app_id );
					 else
					   apex_application_install.generate_application_id;
					 end if;
                     
					 apex_application_install.generate_offset;  
					 
					 if (l_schema is not null) then
					   apex_application_install.set_schema( l_schema );
					 end if;
					 
					 if (l_alias is not null) then
					   apex_application_install.set_application_alias( l_alias );					 
					 end if;
					end;`

	if _, err := tx.Exec(lImportSQL, config.appID, config.alias, config.schema, config.workspace); err != nil {
		fmt.Printf("Failed to execute: %s\n", err)
		os.Exit(1)
	}

	requests := strings.Split(string(file), "/\n")

	for _, request := range requests {
		//logger.Debugf("%s", request)

		str := ""

		for _, line := range strings.Split(strings.TrimSuffix(request, "\n"), "\n") {
			if strings.HasPrefix(line, "prompt") {
				logger.Debugf("%s", line)
			}

			// omit any lines that begin with prompt / set / whenever etc,
			// since we don't support those
			if !strings.HasPrefix(line, "prompt") &&
				!strings.HasPrefix(line, "set define") &&
				!strings.HasPrefix(line, "set verify on") &&
				!strings.HasPrefix(line, "whenever sqlerror") {
				str = str + line + "\n"
			}
		}

		if len(str) == 0 {
			continue
		}

		if _, err := tx.Exec(str); err != nil {
			fmt.Printf("Failed to execute: %s (err: %s\n", str, err)
			f, err := os.Create("./output.txt")
			if err != nil {
				fmt.Println(err)
				return
			}
			_, err = f.WriteString(str)
			if err != nil {
				fmt.Println(err)
				f.Close()
				return
			}
			os.Exit(1)
		}

	}

}

/********************************************************************************/
func main() {

	setDebug(true)
	parseFlags()
	setDebug(config.debugMode)
	loadConfig(config.configFile)
	loadConnection(config.connectionConfig)

	checkMinFlags()

	db, err := sql.Open("goracle", getConnectionString(connection))

	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	execScript(tx, db, config.script)
	tx.Rollback()
	if err != nil {
		panic(err)
	}

}
