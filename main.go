package main

import (
    "flag"
    "fmt"
    "os"
    "time"

    "gopkg.in/ini.v1"
    log "github.com/sirupsen/logrus"
    nested "github.com/antonfisher/nested-logrus-formatter"
)

const (
    CFG_S_OLD = "old"
    CFG_S_OLD_K_DBDRIVER = "db_driver"
    CFG_S_OLD_K_DBHOST = "db_host"
    CFG_S_OLD_K_DBPORT = "db_port"
    CFG_S_OLD_K_DBUSER = "db_user"
    CFG_S_OLD_K_DBPASSWD = "db_passwd"
    CFG_S_OLD_K_DBSCHEMA = "db_schema"
    CFG_S_OLD_K_APIURL = "api_url"
    CFG_S_OLD_K_APIUSER = "api_user"
    CFG_S_OLD_K_APIPASSWD = "api_passwd"

    CFG_S_NEW = "new"
    CFG_S_NEW_K_DBDRIVER = "db_driver"
    CFG_S_NEW_K_DBHOST = "db_host"
    CFG_S_NEW_K_DBPORT = "db_port"
    CFG_S_NEW_K_DBUSER = "db_user"
    CFG_S_NEW_K_DBPASSWD = "db_passwd"
    CFG_S_NEW_K_DBSCHEMA = "db_schema"
    CFG_S_NEW_K_APIURL = "api_url"
    CFG_S_NEW_K_APIUSER = "api_user"
    CFG_S_NEW_K_APIPASSWD = "api_passwd"
)

// app info
var (
    appName             string
    appAuthor           string
    appVersion          string
    appGitCommitHash    string
    appBuildTime        string
    appGoVersion        string
)

// config
var (
    // old zabbix config
    aZDBDriver      string
    aZDBHost        string
    aZDBPort        int
    aZDBUser        string
    aZDBPasswd      string
    aZDBDatabase    string
    aZAPIUrl        string
    aZAPIUser       string
    aZAPIPasswd     string

    // new zabbix config
    bZDBDriver      string
    bZDBHost        string
    bZDBPort        int
    bZDBUser        string
    bZDBPasswd      string
    bZDBDatabase    string
    bZAPIUrl        string
    bZAPIUser       string
    bZAPIPasswd     string
)

// flag
var (
    confPath        string

    helpFlag        bool
    migrateType     string
    checkType       string
    syncType        string

    fHostGroup      string
    fHostIdBegin    int
    fIdOffset         uint

    fLogLevel       uint
)

// runtime
var (
    aZAPI       *ZabbixAPI
    aZDB        *ZabbixDB
    bZAPI       *ZabbixAPI
    bZDB        *ZabbixDB
)

func initLogger() error {
    log.SetFormatter(&nested.Formatter{
        HideKeys:        true,
        TimestampFormat: time.RFC3339,
        FieldsOrder:     []string{"func", "step"},
    })
    return nil
}

func initConfig() error {
    cfg, err := ini.Load(confPath)
    if err != nil {
        return err
    }

    sOLD, err := cfg.GetSection(CFG_S_OLD)
    if err != nil {
        return err
    }
    aZDBDriver      = sOLD.Key(CFG_S_OLD_K_DBDRIVER).Value()
    aZDBHost        = sOLD.Key(CFG_S_OLD_K_DBHOST).Value()
    aZDBPort, _     = sOLD.Key(CFG_S_OLD_K_DBPORT).Int()
    aZDBUser        = sOLD.Key(CFG_S_OLD_K_DBUSER).Value()
    aZDBPasswd      = sOLD.Key(CFG_S_OLD_K_DBPASSWD).Value()
    aZDBDatabase    = sOLD.Key(CFG_S_OLD_K_DBSCHEMA).Value()
    aZAPIUrl        = sOLD.Key(CFG_S_OLD_K_APIURL).Value()
    aZAPIUser       = sOLD.Key(CFG_S_OLD_K_APIUSER).Value()
    aZAPIPasswd     = sOLD.Key(CFG_S_OLD_K_APIPASSWD).Value()

    sNEW, err := cfg.GetSection(CFG_S_NEW)
    if err != nil {
        return err
    }
    bZDBDriver      = sNEW.Key(CFG_S_NEW_K_DBDRIVER).Value()
    bZDBHost        = sNEW.Key(CFG_S_NEW_K_DBHOST).Value()
    bZDBPort, _     = sNEW.Key(CFG_S_NEW_K_DBPORT).Int()
    bZDBUser        = sNEW.Key(CFG_S_NEW_K_DBUSER).Value()
    bZDBPasswd      = sNEW.Key(CFG_S_NEW_K_DBPASSWD).Value()
    bZDBDatabase    = sNEW.Key(CFG_S_NEW_K_DBSCHEMA).Value()
    bZAPIUrl        = sNEW.Key(CFG_S_NEW_K_APIURL).Value()
    bZAPIUser       = sNEW.Key(CFG_S_NEW_K_APIUSER).Value()
    bZAPIPasswd     = sNEW.Key(CFG_S_NEW_K_APIPASSWD).Value()

    return nil
}

func initFlag() error {
    flag.StringVar(&confPath, "f", "zabbix_migrate.ini", "set path of config file than ini format")

    flag.BoolVar(&helpFlag, "h", false, "show for help")
    flag.StringVar(&migrateType, "m", "", "select the type of migrate, support for hostgroup|valuemap|template|host")
    flag.StringVar(&checkType, "c", "", "select the type of check, support for hostgroup|host|item|trigger|valuemap|map|all")
    flag.StringVar(&syncType, "s", "", "select the type of sync, support for trends|history")

    flag.StringVar(&fHostGroup, "g", "", "input params about hostgroup")
    flag.IntVar(&fHostIdBegin, "i", 0, "input params about host begin id")
    flag.UintVar(&fIdOffset, "o", 50, "input params about id offset")

    flag.UintVar(&fLogLevel, "l", 4, "set log level number, 0 is panic ... 6 is trace")

    flag.Usage = flagUsage
    flag.Parse()

    return nil
}

func init() {
    var err error

    err = initLogger()
    if err != nil {
        log.WithFields(log.Fields{
            "func": "init",
            "step": "initLogger",
        }).Fatal(err)
    }

    err = initFlag()
    if err != nil {
        log.WithFields(log.Fields{
            "func": "init",
            "step": "initFlag",
        }).Fatal(err)
    }

    log.SetLevel(log.Level(fLogLevel))
}

func flagUsage() {
    fmt.Fprintf(os.Stderr, `%s:
  Version: %s
  Author: %s
  GitCommit: %s
  BuildTime: %s
  GoVersion: %s
Usage:
  %s [-h] [-f <configPath>] [-m <migrateType>] [-c <checkType>] [-s <syncType>] [-g <hostGroup>] [-i <hostIdBegin]>
Options:`, appName, appVersion, appAuthor, appGitCommitHash, appBuildTime, appGoVersion, appName)
    fmt.Fprintf(os.Stderr, "\n")
    flag.PrintDefaults()
}

func main() {
    var err error

    if helpFlag {
        flag.Usage()
        os.Exit(1)
    }

    err = initConfig()
    if err != nil {
        log.WithFields(log.Fields{
            "func": "init",
            "step": "initConfig",
        }).Fatal(err)
    }

    log.WithFields(log.Fields{
        "func": "main",
    }).Debugf("old zabbix url: %s", aZAPIUrl)
    log.WithFields(log.Fields{
        "func": "main",
    }).Debugf("new zabbix url: %s", bZAPIUrl)

    aZAPI, err = NewZabbixAPI(aZAPIUrl, aZAPIUser, aZAPIPasswd)
    aZDB, err = NewZabbixDB(aZDBDriver, aZDBHost, aZDBPort, aZDBUser, aZDBPasswd, aZDBDatabase)
    if err != nil {
        log.WithFields(log.Fields{
            "func": "main",
            "step": "db.connect",
        }).Fatalf("connect for db [%s:%d] get error: %s", aZDBHost, aZDBPort, err)
    }
    bZAPI, err = NewZabbixAPI(bZAPIUrl, bZAPIUser, bZAPIPasswd)
    bZDB, err = NewZabbixDB(bZDBDriver, bZDBHost, bZDBPort, bZDBUser, bZDBPasswd, bZDBDatabase)
    if err != nil {
        log.WithFields(log.Fields{
            "func": "main",
            "step": "db.connect",
        }).Fatalf("connect for db [%s:%d] get error: %s", bZDBHost, bZDBPort, err)
    }

    _, err = aZAPI.Login()
    if err != nil {
        log.WithFields(log.Fields{
            "func": "main",
            "step": "api.login",
        }).Fatalf( "login for api [%s] get error: %s", aZAPI.url, err)
    }
    _, err = bZAPI.Login()
    if err != nil {
        log.WithFields(log.Fields{
            "func": "main",
            "step": "api.login",
        }).Fatalf( "login for api [%s] get error: %s", bZAPI.url, err)
    }

    if aZAPI == nil || aZDB == nil {
        log.WithFields(log.Fields{
            "func": "main",
        }).Fatal("the old zabbix api or db object is empty")
    }
    if bZAPI == nil || bZDB == nil {
        log.WithFields(log.Fields{
            "func": "main",
        }).Fatal("the new zabbix api or db object is empty")
    }

    if migrateType != "" {
        switch migrateType {
        case "hostgroup":
            err = CreateNewHostGroup(aZAPI, bZAPI)
        case "valuemap":
            err = CreateNewValuemap(aZAPI, bZAPI)
        case "template":
            err = CleanNewTemplate(bZAPI, bZDB)
            if err != nil {
                log.WithFields(log.Fields{
                    "func": "main",
                    "step": "migrate.template",
                }).Fatal("clean template on new zabbix failed")
            }
            err = CreateNewTemplate(aZAPI, aZDB, bZAPI)
        case "host":
            err = CreateNewHost(aZAPI, aZDB, bZAPI, fHostGroup, fHostIdBegin, fIdOffset)
        }
        if err != nil {
            log.WithFields(log.Fields{
                "func": "main",
                "step": "migrate",
            }).Fatal(err)
        }
    }

    if checkType != "" {
        var isSame bool
        if checkType == "hostgroup" || checkType == "all" {
            isSame, err = CheckHostGroup(aZAPI, bZAPI)
            if err != nil {
                log.WithFields(log.Fields{
                    "func": "main",
                    "step": "check.hostgroup",
                }).Fatalf("check for hostgroup is error: %s", err)
            }
            if !isSame {
                fmt.Println("check for hostgroup is different !!!")
            } else {
                fmt.Println("check for hostgroup is same !!!")
            }
        }

        if checkType == "host" || checkType == "all" {
            isSame, err = CheckHost(aZAPI, bZAPI, fHostGroup)
            if err != nil {
                log.WithFields(log.Fields{
                    "func": "main",
                    "step": "check.host",
                }).Errorf("check for host is error: %s", err)
            }
            if !isSame {
                fmt.Println("check for host is different !!!")
            } else {
                fmt.Println("check for host is same !!!")
            }
        }

        if checkType == "item" || checkType == "all" {
            isSame, err = CheckItemGroup(aZAPI, bZAPI, fHostGroup)
            if err != nil {
                log.WithFields(log.Fields{
                    "func": "main",
                    "step": "check.item",
                }).Errorf("check for item is error: %s", err)
            }
            if !isSame {
                fmt.Println("check for item is different !!!")
            } else {
                fmt.Println("check for item is same !!!")
            }
        }

        if checkType == "trigger" || checkType == "all" {
            isSame, err = CheckTriggerNumGroup(aZAPI, bZAPI, fHostGroup)
            if err != nil {
                log.WithFields(log.Fields{
                    "func": "main",
                    "step": "check.trigger",
                }).Errorf("check for trigger is error: %s", err)
            }
            if !isSame {
                fmt.Println("check for trigger number is different !!!")
            } else {
                fmt.Println("check for trigger number is same !!!")
            }
        }

        if checkType == "valuemap" || checkType == "all" {
            isSame, err = CheckValuemap(aZAPI, bZAPI)
            if err != nil {
                log.WithFields(log.Fields{
                    "func": "main",
                    "step": "check.valuemap",
                }).Errorf("check for valuemap is error: %s", err)
            }
            if !isSame {
                fmt.Println("check for valuemap is different !!!")
            } else {
                fmt.Println("check for valuemap is same !!!")
            }
        }

        if checkType == "map" || checkType == "all" {
            isSame, err = CheckMap(aZAPI, bZAPI)
            if err != nil {
                log.WithFields(log.Fields{
                    "func": "main",
                    "step": "check.map",
                }).Errorf("check for map is error: %s", err)
            }
            if !isSame {
                fmt.Println("check for map is different !!!")
            } else {
                fmt.Println("check for map is same !!!")
            }
        }
    }

    if syncType != "" {
        switch syncType {
        case "trends":
            err = SyncTrends(aZDB, bZDB, fHostGroup, fHostIdBegin)
            if err != nil {
                log.WithFields(log.Fields{
                    "func": "main",
                    "step": "sync.trends",
                }).Errorf("sync for trneds is error: %s", err)
            }
        case "history":
            err = SyncHistory(aZDB, bZDB, fHostGroup, fHostIdBegin)
            if err != nil {
                log.WithFields(log.Fields{
                    "func": "main",
                    "step": "sync.history",
                }).Errorf("sync for history is error: %s", err)
            }
        default:
            log.WithFields(log.Fields{
                "func": "main",
                "step": "sync.default",
            }).Errorf("sync not support for %s", syncType)
        }
    }

}
