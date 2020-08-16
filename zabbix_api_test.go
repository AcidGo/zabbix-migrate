package main

import (
    "testing"
    "log"
)

func TestZabbixAPI(t *testing.T) {
    api, _ := NewZabbixAPI(
        "http://192.168.52.61/zabbix/api_jsonrpc.php",
        "Admin",
        "zabbix",
    )
    _, err := api.Login()
    if err != nil {
        log.Println(err)
    }
    log.Println(api.auth)
    params := make(map[string]interface{}, 0)
    filter := make(map[string][]string, 0)
    params["output"] = "extend"
    filter["name"] = []string{"Linux servers", "Zabbix servers"}
    params["filter"] = filter
    res, err := api.HostGroup("get", params)
    if err != nil {
        log.Println(err)
    }
    log.Println(res)
}

func TestHGCreate(t *testing.T) {
    aAPI, _ := NewZabbixAPI(
        "http://192.168.52.61/zabbix/api_jsonrpc.php",
        "Admin",
        "zabbix",
    )
    bAPI, _ := NewZabbixAPI(
        "http://192.168.52.61/zabbix/api_jsonrpc.php",
        "Admin",
        "zabbix",
    )

    aAPI.Login()
    bAPI.Login()

    err := CreateNewHostGroup(aAPI, bAPI)
    if err != nil {
        log.Println(err)
    }
}

func TestGetHost(t *testing.T) {
    api, _ := NewZabbixAPI(
        "http://192.168.52.61/zabbix/api_jsonrpc.php",
        "Admin",
        "zabbix",
    )
    _, err := api.Login()
    if err != nil {
        log.Println(err)
    }
    params := make(map[string]interface{}, 0)
    filter := make(map[string][]string, 0)
    filter["status"] = []string{"0"}
    params["filter"] = filter
    res, err := api.Host("get", params)
    if err != nil {
        log.Println(err)
    }
    log.Println(res)
}


func TestTemplate(t *testing.T) {
    api, _ := NewZabbixAPI(
        "http://192.168.52.61/zabbix/api_jsonrpc.php",
        "Admin",
        "zabbix",
    )
    _, err := api.Login()
    if err != nil {
        log.Println(err)
    }
    params := make(map[string]interface{}, 0)
    filter := make(map[string][]string, 0)
    filter["host"] = []string{"Template OS Linux"}
    params["filter"] = filter
    params["output"] = "extend"
    res, err := api.Template("get", params)
    if err != nil {
        log.Println(err)
    }
    log.Println(res)
}

func TestConfiguration(t *testing.T) {
    api, _ := NewZabbixAPI(
        "http://192.168.52.61/zabbix/api_jsonrpc.php",
        "Admin",
        "zabbix",
    )
    _, err := api.Login()
    if err != nil {
        log.Println(err)
    }
    params := make(map[string]interface{}, 0)
    options := make(map[string]interface{}, 0)
    options["templates"] = []string{"10225", "10226"}
    params["options"] = options
    params["format"] = "xml"
    res, err := api.Configuration("export", params)
    if err != nil {
        log.Println(err)
    }
    log.Println(res)
}