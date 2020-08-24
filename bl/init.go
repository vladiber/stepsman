package bl

import (
	"flag"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path"
)

var DB *sqlx.DB
var StoreDir string
var cfgFile string

func InitBL(cfgFile string) error {
	flag.Parse()
	dir, err := homedir.Dir()
	if err != nil {
		return fmt.Errorf("failed to detect home directory: %w", err)
	}
	StoreDir = path.Join(dir, ".stepsman")
	_, err = os.Stat(StoreDir)

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(StoreDir)
		viper.SetConfigName(".stepsman")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Info("Using config file:", viper.ConfigFileUsed())
	}

	// use this later on
	//log.SetOutput(&lumberjack.Logger{
	//	Filename:   path.Join(StoreDir, "stepsman.log"),
	//	MaxSize:    100, // megabytes
	//	MaxBackups: 2,
	//	MaxAge:     1, // days
	//	Compress:   true,
	//})

	mw := io.MultiWriter(os.Stdout, &lumberjack.Logger{
		Filename:   path.Join(StoreDir, "stepsman.log"),
		MaxSize:    10, // megabytes
		MaxBackups: 2,
		MaxAge:     1, // days
		Compress:   true,
	})
	log.SetOutput(mw)
	log.SetLevel(log.TraceLevel)

	if os.IsNotExist(err) {
		err = os.MkdirAll(StoreDir, 0700)
		if err != nil {
			return fmt.Errorf("failed to create the .stepsman diretory: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to determine existance of .stepsman directory: %w", err)
	}

	DB, err = sqlx.Open("sqlite3", path.Join(StoreDir, "stepsman.DB"))
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("failed to open a database connection: %w", err)
	}
	err = upgrade()
	if err != nil {
		return err
	}
	return nil
}

func upgrade() error {
	var version = -1
	_, err := DB.Exec(`CREATE TABLE IF NOT EXISTS migration (
    id INTEGER PRIMARY KEY,
	version INTEGER
    );`)
	if err != nil {
		return fmt.Errorf("failed to verify database migration table creation: %w", err)
	}
	tx, err := DB.Beginx()
	if err != nil {
		return fmt.Errorf("failed to start a database transaction: %w", err)
	}
	{
		var count = -1
		err = tx.Get(&count, "select count(*) from migration")
		if err != nil {
			err = Rollback(tx, err)
			return fmt.Errorf("failed to get database table migration count: %w", err)
		}
		if count == 0 {
			_, err = tx.Exec("insert into migration (id, version) values(1,0)")
			if err != nil {
				err = Rollback(tx, err)
				return fmt.Errorf("failed to add database migration row: %w", err)
			}
		}
	}
	err = tx.Get(&version, "select version from migration where id=1")
	if err != nil {
		err = Rollback(tx, err)
		return fmt.Errorf("failed to get database version: %w", err)
	}
	switch version {
	case 0:
		_, err := tx.Exec(`CREATE TABLE runs (
                                     id INTEGER PRIMARY KEY AUTOINCREMENT,
                                     uuid TEXT,
	                                 name TEXT,
	                                 status INTEGER,
	                                 checked INTEGER,
	                                 script TEXT
                                     )`)
		if err != nil {
			err = Rollback(tx, err)
			return fmt.Errorf("failed to create database runs table: %w", err)
		}
		_, err = tx.Exec(`CREATE UNIQUE INDEX idx_runs_uuid ON runs (uuid)`)
		if err != nil {
			err = Rollback(tx, err)
			return fmt.Errorf("failed to create index idx_runs_name_status: %w", err)
		}
		_, err = tx.Exec(`CREATE INDEX idx_runs_name_status ON runs (name, status)`)
		if err != nil {
			err = Rollback(tx, err)
			return fmt.Errorf("failed to create index idx_runs_name_status: %w", err)
		}

		_, err = tx.Exec(`CREATE INDEX idx_runs_status ON runs (status)`)
		if err != nil {
			err = Rollback(tx, err)
			return fmt.Errorf("failed to create index idx_runs_status: %w", err)
		}
		//CREATE INDEX idx_contacts_name
		//ON contacts (first_name, last_name);
		_, err = tx.Exec("update migration set version=1 where id=1")
		if err != nil {
			err = Rollback(tx, err)
			return fmt.Errorf("failed to update database migration row to version 1: %w", err)
		}
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit migration transaction: %w", err)
	}
	return nil
}
