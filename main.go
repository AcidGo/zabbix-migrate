package main

import (
    "flag"
    "fmt"
    "log"
    "os"

    "gopkg.in/ini.v1"
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
    helpFlag        bool
    migrateType     string
    checkType       string
    syncType        string

    fHostGroup      string
    fHostIdBegin    int
)

// runtime
var (
    aZAPI       *ZabbixAPI
    aZDB        *ZabbixDB
    bZAPI       *ZabbixAPI
    bZDB        *ZabbixDB
)

func initConfig() error {
    cfg, err := ini.Load("zabbix_migrate.ini")
    if err != nil {
        return err
    }

    sOLD, err := cfg.GetSection("old")
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

    sNEW, err := cfg.GetSection("old")
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
    flag.BoolVar(&helpFlag, "h", false, "show for help")
    flag.StringVar(&migrateType, "m", "", "select the type of migrate")
    flag.StringVar(&checkType, "c", "", "select the type of check")
    flag.StringVar(&syncType, "s", "", "select the type of sync")

    flag.Usage = flagUsage
    flag.Parse()

    return nil
}

func init() {
    var err error

    err = initConfig()
    if err != nil {
        log.Fatal(err)
    }

    err = initFlag()
    if err != nil {
        log.Fatal(err)
    }
}

func flagUsage() {
    fmt.Fprintf(os.Stderr, `zabbix_migrate:
    version: %s/%s
    author: %s
    gitCommit: %s
    buildTime: %s
Usage: %s [-h] [-m migrateType] [-c checkType] [-s syncType]
Options:`, appName, appVersion, appAuthor, appGitCommitHash, appBuildTime, appName)
    fmt.Fprintf(os.Stderr, "\n")
    flag.PrintDefaults()
}

func main() {
    var err error
    if helpFlag {
        flag.Usage()
        os.Exit(1)
    }

    if aZAPI == nil || aZDB == nil {
        fmt.Fprintln(os.Stderr, "the old zabbix api or db object is empty")
        os.Exit(1)
    }
    if bZAPI == nil || bZDB == nil {
        fmt.Fprintln(os.Stderr, "the new zabbix api or db object is empty")
        os.Exit(1)
    }

    if migrateType != "" {
        switch migrateType {
        case "hostgroup":
            err = CreateNewHostGroup(aZAPI, bZAPI)
        case "template":
            err = CleanNewTemplate(bZAPI, bZDB)
            if err != nil {
                fmt.Fprintln(os.Stderr, "clean template on new zabbix failed")
                os.Exit(1)
            }
            err = CreateNewTemplate(aZAPI, aZDB, bZAPI)
        case "host":
            err = CreateNewHost(aZAPI, aZDB, bZAPI, fHostGroup, fHostIdBegin)
        }
        if err != nil {
            fmt.Fprintln(os.Stderr, err)
            os.Exit(1)
        }
    }

    if checkType != "" {
        var isSame bool
        if checkType == "hostgroup" || checkType == "all" {
            isSame, err = CheckHostGroup(aZAPI, bZAPI)
            if err != nil {
                fmt.Fprintf(os.Stderr, "check for hostgroup is error: %s\n", err)
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
                fmt.Fprintf(os.Stderr, "check for host is error: %s\n", err)
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
                fmt.Fprintf(os.Stderr, "check for item is error: %s\n", err)
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
                fmt.Fprintf(os.Stderr, "check for trigger is error: %s\n", err)
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
                fmt.Fprintf(os.Stderr, "check for valuemap is error: %s\n", err)
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
                fmt.Fprintf(os.Stderr, "check for map is error: %s\n", err)
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
                fmt.Fprintf(os.Stderr, "sync for trneds is error: %s\n", err)
            }
        case "history":
            err = SyncHistory(aZDB, bZDB, fHostGroup, fHostIdBegin)
            if err != nil {
                fmt.Fprintf(os.Stderr, "sync for history is error: %s\n", err)
            }
        default:
            fmt.Fprintf(os.Stderr, "sync not support for %s\n", syncType)
        }
    }

}