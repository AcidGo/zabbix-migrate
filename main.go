package main

import (
    "flag"
    "fmt"
    "log"
    "os"

    // "gopkg.in/ini.v1"
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
    return nil
}

func initFlag() error {
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
    fmt.Fprintf(os.Stderr, ``)
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
        var isDiff bool
        switch checkType {
        case "hostgroup":
            isDiff, err = CheckHostGroup(aZAPI, bZAPI)
            if err != nil {
                fmt.Fprintf(os.Stderr, "check for hostgroup is error: %s\n", err)
                os.Exit(1)
            }
            if isDiff {
                fmt.Println("check for hostgroup is different !!!")
            } else {
                fmt.Println("check for hostgroup is same !!!")
            }
        case "host":
            isDiff, err = CheckHost(aZAPI, bZAPI, fHostGroup)
            if err != nil {
                fmt.Fprintf(os.Stderr, "check for host is error: %s\n", err)
                os.Exit(1)
            }
            if isDiff {
                fmt.Println("check for host is different !!!")
            } else {
                fmt.Println("check for host is same !!!")
            }
        }
    }
}