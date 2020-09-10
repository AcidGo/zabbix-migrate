package main

import (
    "log"
    "testing"
    // "fmt"
)

func GetDBConnectA() (*ZabbixDB, error) {
    return NewZabbixDB("mysql", "192.168.52.61", 3306, "zbxtest", "abcd1234", "zabbix")
}

func GetAPIA() (*ZabbixAPI, error) {
    api, _ := NewZabbixAPI(
        "http://192.168.66.50/api_jsonrpc.php",
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

// func TestCreateNewTemplate(t *testing.T) {
//     zapiA, _ := GetAPIA()
//     zapiB, _ := GetAPIB()

//     err := CreateNewTemplate(zapiA, zapiB)
//     if err != nil {
//         log.Println(err)
//     }
// }

// func TestCreateNewHost(t *testing.T) {
//     zapiA, _ := GetAPIA()
//     zapiB, _ := GetAPIB()
//     zdbA, _ := GetDBConnectA()

//     err := CreateNewHost(zapiA, zdbA, zapiB, "", 0)
//     if err != nil {
//         log.Println(err)
//     }
// }

// func TestGetHostMapList(t *testing.T) {
//     zdbA, _ := GetDBConnectA()

//     res, err := zdbA.GetHostMapList("Linux servers", 0)
//     log.Println(res)
//     log.Println(err)
// }

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

    err := zdbA.SyncHistoryToOne(zbxB, "history_text", 10266, "192.168.52.61_midware")
    log.Println(err)
}

func TestSyncTrendsToOne(t *testing.T) {
    zdbA, _ := GetDBConnectA()
    zbxB, _ := GetDBConnectB()

    err := zdbA.SyncTrendsToOne(zbxB, "trends", 10266, "192.168.52.61_midware")
    log.Println(err)
}

// func TestSyncHistory(t *testing.T) {
//     zdbA, _ := GetDBConnectA()
//     zbxB, _ := GetDBConnectB()

//     SyncHistory(zdbA, zbxB, "Linux servers", 0)
// }

// func TestSyncTrends(t *testing.T) {
//     zdbA, _ := GetDBConnectA()
//     zbxB, _ := GetDBConnectB()

//     err := SyncTrends(zdbA, zbxB, "Linux servers", 0)
//     log.Println(err)
// }

func TestDiffUnitList(t *testing.T) {
    m := []ZUnitMap {
        map[string]interface{} {
            "a": 1,
            "b": 2,
        },
        map[string]interface{} {
            "c": 1,
            "d": 2,
        },
    }
    n := []ZUnitMap {
        map[string]interface{} {
            "a": 1,
            "b": 2,
        },
        map[string]interface{} {
            "c": 1,
            "d": 3,
        },
    }
    res, err := DiffUnitList(m, n, true)
    log.Println(res)
    log.Println(err)
}

func TestSortTemplateDepend(t *testing.T) {
    zapiA, err := GetAPIA()
    res, err := SortTemplateDepend(zapiA)
    log.Println(err)
    log.Println(res)
}

func TestCheckHost(t *testing.T) {
    zapiA, err := GetAPIA()
    _, err = CheckHost(zapiA, zapiA, "Linux servers")
    log.Println(err)
}