package main

import (
    "log"
    "testing"
)

func GetDBConnectA() (*ZabbixDB, error) {
    return NewZabbixDB("mysql", "192.168.52.61", 3306, "zbxtest", "abcd1234", "zabbix")
}

func GetAPIA() (*ZabbixAPI, error) {
    api, _ := NewZabbixAPI(
        "http://192.168.52.61/zabbix/api_jsonrpc.php",
        "Admin",
        "zabbix",
    )
    _, err := api.Login()
    return api, err
}

func GetAPIB() (*ZabbixAPI, error) {
    api, _ := NewZabbixAPI(
        "http://192.168.52.62/zabbix/api_jsonrpc.php",
        "Admin",
        "zabbix",
    )
    _, err := api.Login()
    return api, err
}

func GetDBConnectB() (*ZabbixDB, error) {
    return NewZabbixDB("mysql", "192.168.52.62", 3306, "zbxtest", "abcd1234", "zabbix")
}

func TestDBConnect(t *testing.T) {
    _, err := NewZabbixDB("mysql", "192.168.52.61", 3306, "zbxtest", "abcd1234", "zabbix")
    if err != nil {
        log.Println(err)
    }
}

func TestGetTemplateList(t *testing.T) {
    zdb, _ := GetDBConnectA()
    res, err := zdb.GetTemplateList()
    if err != nil {
        log.Println(err)
    }
    log.Println(res)
}

func TestCleanNewTemplate(t *testing.T) {
    zdb, _ := GetDBConnectB()
    zapi, _ := GetAPIB()
    err := CleanNewTemplate(zapi, zdb)
    if err != nil {
        log.Println(err)
    }
}

func TestCreateNewTemplate(t *testing.T) {
    zapiA, _ := GetAPIA()
    zapiB, _ := GetAPIB()
    zdbA, _ := GetDBConnectA()

    err := CreateNewTemplate(zapiA, zdbA, zapiB)
    if err != nil {
        log.Println(err)
    }
}

func TestCreateNewHost(t *testing.T) {
    zapiA, _ := GetAPIA()
    zapiB, _ := GetAPIB()
    zdbA, _ := GetDBConnectA()

    err := CreateNewHost(zapiA, zdbA, zapiB, "", 0)
    if err != nil {
        log.Println(err)
    }
}

func TestGetHostMapList(t *testing.T) {
    zdbA, _ := GetDBConnectA()

    res, err := zdbA.GetHostMapList("Linux servers", 0)
    log.Println(res)
    log.Println(err)
}

func TestGetItemList(t *testing.T) {
    zdbA, _ := GetDBConnectA()

    res, err := zdbA.GetItemList(10263)
    log.Println(res)
    log.Println(err)
}

func TestGetItemMap(t *testing.T) {
    zdbA, _ := GetDBConnectA()

    res, err := zdbA.GetItemMap(10263)
    log.Println(res)
    log.Println(err)
}

func TestSyncHistoryToOne(t *testing.T) {
    zdbA, _ := GetDBConnectA()
    zbxB, _ := GetDBConnectB()

    err := zdbA.SyncHistoryToOne(zbxB, "history", 10263, "192.168.52.31")
    log.Println(err)
}