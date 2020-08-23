package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "time"
    "reflect"
    log "github.com/sirupsen/logrus"
)

var (
    HistoryTables = []string{
        "history",
        "history_log",
        "history_str",
        "history_text",
        "history_uint",
    }
    TrendsTables = []string{
        "trends",
        "trends_uint",
    }
)


func DiffUnitList(aUnitList, bUnitList []ZUnitMap, hasEcho bool) (bool, error) {
    var isSame bool
    aDiffUnit := make([]map[string]interface{}, 0)
    bDiffUnit := make([]map[string]interface{}, 0)
    _bDiffUnit := make([]map[string]interface{}, 0)
    for _, v := range bUnitList {
        _bDiffUnit = append(_bDiffUnit, v)
    }

    for _, aV := range aUnitList {
        var isExist = false
        for bI, bV := range bUnitList {
            if reflect.DeepEqual(aV, bV) {
                _bDiffUnit[bI] = nil
                isExist = true
            }
        }
        if !isExist {
            aDiffUnit = append(aDiffUnit, aV)
        }
    }

    for _, v := range _bDiffUnit {
        if v == nil {
            continue
        }
        bDiffUnit = append(bDiffUnit, v)
    }

    if len(aDiffUnit) == 0 && len(bDiffUnit) == 0 {
        isSame = true
    } else {
        isSame = false
    }

    if !hasEcho {
        return isSame, nil
    }

    fmt.Println("===[start: detail for a]")
    for _, v := range aDiffUnit {
        jsonStr, err := json.Marshal(v)
        if err != nil {
            return false, err
        }
        fmt.Printf("- %s\n", jsonStr)
    }
    fmt.Println("===[end:   detail for a]")

    fmt.Println("===[start: detail for b]")
    for _, v := range bDiffUnit {
        jsonStr, err := json.Marshal(v)
        if err != nil {
            return false, err
        }
        fmt.Printf("+ %s\n", jsonStr)
    }
    fmt.Println("===[end:   detail for b]")

    return isSame, nil
}

func CreateNewHostGroup(aZAPI, bZAPI *ZabbixAPI) error {
    log.WithFields(log.Fields{
        "func": "CreateNewHostGroup",
        "step": "start",
    }).Debug("start create new host group on new zabbix")

    aParams := make(map[string]interface{}, 0)
    aFilter := make(map[string]interface{}, 0)
    aParams["output"] = "extend"
    aParams["filter"] = aFilter
    aZHostGroupList, err := aZAPI.HostGroup("get", aParams)
    if err != nil {
        return err
    }

    bParams := make(map[string]interface{}, 0)
    bFilter := make(map[string]interface{}, 0)
    bParams["output"] = "extend"
    bParams["filter"] = bFilter
    bZHostGroupList, err := bZAPI.HostGroup("get", bParams)
    if err != nil {
        return err
    }

    tParams := make(map[string]interface{}, 0)
    for _, aZHostGroup := range aZHostGroupList {
        hasExists := false
        for _, bZHostGroup := range bZHostGroupList {
            if bZHostGroup["name"] == aZHostGroup["name"] {
                hasExists = true
                break
            }
        }
        if hasExists {
            log.WithFields(log.Fields{
                "func": "CreateNewHostGroup",
                "step": "check.isExist",
            }).Infof("host group [%s] alread exists", aZHostGroup["name"])
            continue
        }
        tParams["name"] = aZHostGroup["name"]
        _, err := bZAPI.HostGroup("create", tParams)
        if err != nil {
            return err
        }
    }

    log.WithFields(log.Fields{
        "func": "CreateNewHostGroup",
        "step": "finish",
    }).Debug("finish create new host group on new zabbix")
    return nil
}

func CreateNewValuemap(aZAPI ,bZAPI *ZabbixAPI) error {
    log.WithFields(log.Fields{
        "func": "CreateNewValuemap",
        "step": "start",
    }).Debug("start create new valuemap on new zabbix")

    aParams := make(map[string]interface{}, 0)
    aParams["output"] = "extend"
    aZValuemapList, err := aZAPI.Valuemap("get", aParams)
    if err != nil {
        return err
    }
    aValuemapList := make([]string, 0)
    for _, zUM := range aZValuemapList {
        if valI, ok := zUM["valuemapid"]; ok {
            if valS, _ok := valI.(string); _ok {
                aValuemapList = append(aValuemapList, valS)
            }
        }
    }

    step := 10
    stepN := int(len(aValuemapList)/step)
    for i:=0; i<=stepN; i++ {
        var tValuemapList []string
        if i == stepN {
            tValuemapList = aValuemapList[step*i:len(aValuemapList)]
        } else {
            tValuemapList = aValuemapList[step*i:step*(i+1)]
        }
        aParams := make(map[string]interface{}, 0)
        aOptions := make(map[string]interface{}, 0)
        aOptions["valueMaps"] = tValuemapList
        aParams["options"] = aOptions
        aParams["format"] = "xml"
        aTemplateExport, err := aZAPI.Configuration("export", aParams)
        if err != nil {
            log.WithFields(log.Fields{
                "func": "CreateNewValuemap",
                "step": "export",
            }).Errorf("try to export first valuemap [%s] is failed", tValuemapList[0])
            return err
        }

        bParams := make(map[string]interface{}, 0)
        bRules := make(map[string]interface{}, 0)
        bRules["groups"] = map[string]bool{
            "createMissing": false,
        }
        bRules["hosts"] = map[string]bool{
            "updateExisting": false,
            "createMissing": false,
        }
        bRules["templates"] = map[string]bool{
            "updateExisting": false,
            "createMissing": false,
        }
        bRules["templateScreens"] = map[string]bool{
            "updateExisting": false,
            "createMissing": false,
            "deleteMissing": false,
        }
        bRules["templateLinkage"] = map[string]bool{
            "createMissing": false,
        }
        bRules["applications"] = map[string]bool{
            "createMissing": false,
            "deleteMissing": false,
        }
        bRules["items"] = map[string]bool{
            "updateExisting": false,
            "createMissing": false,
            "deleteMissing": false,
        }
        bRules["discoveryRules"] = map[string]bool{
            "updateExisting": false,
            "createMissing": false,
            "deleteMissing": false,
        }
        bRules["triggers"] = map[string]bool{
            "updateExisting": false,
            "createMissing": false,
            "deleteMissing": false,
        }
        bRules["graphs"] = map[string]bool{
            "updateExisting": false,
            "createMissing": false,
            "deleteMissing": false,
        }
        bRules["httptests"] = map[string]bool{
            "updateExisting": false,
            "createMissing": false,
            "deleteMissing": false,
        }
        bRules["screens"] = map[string]bool{
            "updateExisting": false,
            "createMissing": false,
        }
        bRules["maps"] = map[string]bool{
            "updateExisting": false,
            "createMissing": false,
        }
        bRules["images"] = map[string]bool{
            "updateExisting": false,
            "createMissing": false,
        }
        bRules["valueMaps"] = map[string]bool{
            "updateExisting": false,
            "createMissing": true,
        }
        bParams["rules"] = bRules
        bParams["format"] = "xml"
        bParams["source"] = aTemplateExport
        res, err := bZAPI.Configuration("import", bParams)
        if err != nil {
            log.WithFields(log.Fields{
                "func": "CreateNewValuemap",
                "step": "import",
            }).Errorf("try to import first valuemap [%s] is failed", tValuemapList[0])
            return err
        }
        if res.(bool) != true {
            log.WithFields(log.Fields{
                "func": "CreateNewValuemap",
                "step": "import",
            }).Errorf("try to import first valuemap [%s] is failed", tValuemapList[0])
            return errors.New("result of import valuemap task is false")
        }
        time.Sleep(2*time.Second)
    }

    log.WithFields(log.Fields{
        "func": "CreateNewValuemap",
        "step": "start",
    }).Debug("start create new valuemap on new zabbix")
    return nil
}

func SortTemplateDepend(aZAPI *ZabbixAPI) ([]string, error) {
    // res := make([]string, 0)
    aParams := make(map[string]interface{}, 0)
    aFilter := make(map[string]interface{}, 0)
    aParams["output"] = "templateid"
    aParams["filter"] = aFilter
    aParams["selectParentTemplates"] = []string{"templateid"}
    aZMulTemplateList, err := aZAPI.Template("get", aParams)
    if err != nil {
        return []string{}, err
    }
    fmt.Println(aZMulTemplateList)

    tDependMap := make(map[string][]string, 0)
    for _, zUM := range aZMulTemplateList {
        if templateidI, ok := zUM["templateid"]; ok {
            if templateidS, ok := templateidI.(string); ok {
                if parentTemplatesI, ok := zUM["parentTemplates"]; ok {
                    if parentTemplatesM, ok := parentTemplatesI.([]interface{}); ok {
                        fmt.Println("ffffffffff")
                        fmt.Println(templateidS)
                        fmt.Println(parentTemplatesM)
                        // for _, valM := range parentTemplatesM {
                        //     for _, val := range valM {
                        //         tDependMap[templateidS] = append(tDependMap[templateidS], val)
                        //     }
                        // }
                    }
                }
            }
        }
    }

    fmt.Println(tDependMap)

    return []string{"1"}, nil
}

func CleanNewTemplate(bZAPI *ZabbixAPI, bZDB *ZabbixDB) error {
    log.WithFields(log.Fields{
        "func": "CleanNewTemplate",
        "step": "start",
    }).Debug("start clean all template on new zabbix")

    bTemplateList, err := bZDB.GetTemplateList()
    if err != nil {
        return err
    }
    if len(bTemplateList) == 0 {
        return nil
    }

    step := 10
    stepN := int(len(bTemplateList)/step)
    for i:=0;i<=stepN;i++ {
        var tTemplateList []int
        if i == stepN {
            tTemplateList = bTemplateList[step*i:len(bTemplateList)]
        } else {
            tTemplateList = bTemplateList[step*i:step*(i+1)]
        }
        bParams := tTemplateList
        _, err := bZAPI.Template("delete", bParams)
        if err != nil {
            if len(tTemplateList) > 0 {
                log.Errorf("try to delete first template [%d] is failed", tTemplateList[0])
            }
            return err
        }
        time.Sleep(2*time.Second)
    }

    log.WithFields(log.Fields{
        "func": "CleanNewTemplate",
        "step": "finish",
    }).Debug("finish clean all template on new zabbix")
    return nil
}

func CreateNewTemplate(aZAPI *ZabbixAPI, aZDB *ZabbixDB, bZAPI *ZabbixAPI) error {
    log.WithFields(log.Fields{
        "func": "CreateNewTemplate",
        "step": "start",
    }).Debug("start create new template on new zabbix")

    aTemplateList, err := aZDB.GetTemplateList()
    if err != nil {
        return err
    }
    if len(aTemplateList) == 0 {
        return nil
    }

    step := 10
    stepN := int(len(aTemplateList)/step)
    for i:=0; i<=stepN; i++ {
        var tTemplateList []int
        if i == stepN {
            tTemplateList = aTemplateList[step*i:len(aTemplateList)]
        } else {
            tTemplateList = aTemplateList[step*i:step*(i+1)]
        }
        aParams := make(map[string]interface{}, 0)
        aOptions := make(map[string]interface{}, 0)
        aOptions["templates"] = tTemplateList
        aParams["options"] = aOptions
        aParams["format"] = "xml"
        aTemplateExport, err := aZAPI.Configuration("export", aParams)
        if err != nil {
            log.WithFields(log.Fields{
                "func": "CreateNewTemplate",
                "step": "export",
            }).Errorf("try to export first template [%d] is failed", tTemplateList[0])
            return err
        }

        bParams := make(map[string]interface{}, 0)
        bRules := make(map[string]interface{}, 0)
        bRules["groups"] = map[string]bool{
            "createMissing": true,
        }
        bRules["hosts"] = map[string]bool{
            "updateExisting": false,
            "createMissing": false,
        }
        bRules["templates"] = map[string]bool{
            "updateExisting": true,
            "createMissing": true,
        }
        bRules["templateScreens"] = map[string]bool{
            "updateExisting": true,
            "createMissing": true,
            "deleteMissing": false,
        }
        bRules["templateLinkage"] = map[string]bool{
            "createMissing": false,
        }
        bRules["applications"] = map[string]bool{
            "createMissing": true,
            "deleteMissing": false,
        }
        bRules["items"] = map[string]bool{
            "updateExisting": true,
            "createMissing": true,
            "deleteMissing": false,
        }
        bRules["discoveryRules"] = map[string]bool{
            "updateExisting": true,
            "createMissing": true,
            "deleteMissing": false,
        }
        bRules["triggers"] = map[string]bool{
            "updateExisting": true,
            "createMissing": true,
            "deleteMissing": false,
        }
        bRules["graphs"] = map[string]bool{
            "updateExisting": true,
            "createMissing": true,
            "deleteMissing": false,
        }
        bRules["httptests"] = map[string]bool{
            "updateExisting": true,
            "createMissing": true,
            "deleteMissing": false,
        }
        bRules["screens"] = map[string]bool{
            "updateExisting": false,
            "createMissing": false,
        }
        bRules["maps"] = map[string]bool{
            "updateExisting": false,
            "createMissing": false,
        }
        bRules["images"] = map[string]bool{
            "updateExisting": false,
            "createMissing": false,
        }
        bRules["valueMaps"] = map[string]bool{
            "updateExisting": false,
            "createMissing": true,
        }
        bParams["rules"] = bRules
        bParams["format"] = "xml"
        bParams["source"] = aTemplateExport
        res, err := bZAPI.Configuration("import", bParams)
        if err != nil {
            log.WithFields(log.Fields{
                "func": "CreateNewTemplate",
                "step": "import",
            }).Errorf("try to import first template [%d] is failed", tTemplateList[0])
            return err
        }
        if res.(bool) != true {
            log.WithFields(log.Fields{
                "func": "CreateNewTemplate",
                "step": "import",
            }).Errorf("try to import first template [%d] is failed", tTemplateList[0])
            return errors.New("result of import template task is false")
        }
        time.Sleep(2*time.Second)
    }

    log.WithFields(log.Fields{
        "func": "CreateNewTemplate",
        "step": "finish",
    }).Debug("finish creat new template on new zabbix")
    return nil
}

func CreateNewHost(aZAPI *ZabbixAPI, aZDB *ZabbixDB, bZAPI *ZabbixAPI, hostgroup string, hostIdBegin int) error {
    log.WithFields(log.Fields{
        "func": "CreateNewHost",
        "step": "start",
    }).Debug("start create new host on new zabbix")

    aHostList, err := aZDB.GetHostList(hostgroup, hostIdBegin)
    if err != nil {
        return err
    }
    if len(aHostList) == 0 {
        return nil
    }

    step := 100
    stepN := int(len(aHostList)/step)
    for i:=0; i<=stepN; i++ {
        var tHostList []int
        if i == stepN {
            tHostList = aHostList[step*i:len(aHostList)]
        } else {
            tHostList = aHostList[step*i:step*(i+1)]
        }
        aParams := make(map[string]interface{}, 0)
        aOptions := make(map[string]interface{}, 0)
        aOptions["hosts"] = tHostList
        aParams["options"] = aOptions
        aParams["format"] = "xml"
        aHostExport, err := aZAPI.Configuration("export", aParams)
        if err != nil {
            if len(tHostList) > 0{
                log.WithFields(log.Fields{
                    "func": "CreateNewHost",
                    "step": "export",
                }).Errorf("try to export first host [%d] is failed", tHostList[0])
            } else {
                log.WithFields(log.Fields{
                    "func": "CreateNewHost",
                    "step": "export",
                }).Error("UNKNOWN")
            }
            return err
        }

        bParams := make(map[string]interface{}, 0)
        bRules := make(map[string]interface{}, 0)
        bRules["groups"] = map[string]bool{
            "createMissing": true,
        }
        bRules["hosts"] = map[string]bool{
            "updateExisting": true,
            "createMissing": true,
        }
        bRules["templates"] = map[string]bool{
            "updateExisting": false,
            "createMissing": false,
        }
        bRules["templateScreens"] = map[string]bool{
            "updateExisting": false,
            "createMissing": false,
            "deleteMissing": false,
        }
        bRules["templateLinkage"] = map[string]bool{
            "createMissing": true,
        }
        bRules["applications"] = map[string]bool{
            "createMissing": true,
            "deleteMissing": false,
        }
        bRules["items"] = map[string]bool{
            "updateExisting": true,
            "createMissing": true,
            "deleteMissing": false,
        }
        bRules["discoveryRules"] = map[string]bool{
            "updateExisting": true,
            "createMissing": true,
            "deleteMissing": false,
        }
        bRules["triggers"] = map[string]bool{
            "updateExisting": true,
            "createMissing": true,
            "deleteMissing": false,
        }
        bRules["graphs"] = map[string]bool{
            "updateExisting": true,
            "createMissing": true,
            "deleteMissing": false,
        }
        bRules["httptests"] = map[string]bool{
            "updateExisting": true,
            "createMissing": true,
            "deleteMissing": false,
        }
        bRules["screens"] = map[string]bool{
            "updateExisting": false,
            "createMissing": false,
        }
        bRules["maps"] = map[string]bool{
            "updateExisting": false,
            "createMissing": false,
        }
        bRules["images"] = map[string]bool{
            "updateExisting": false,
            "createMissing": false,
        }
        bRules["valueMaps"] = map[string]bool{
            "updateExisting": false,
            "createMissing": true,
        }
        bParams["rules"] = bRules
        bParams["format"] = "xml"
        bParams["source"] = aHostExport
        res, err := bZAPI.Configuration("import", bParams)
        if err != nil {
            log.WithFields(log.Fields{
                "func": "CreateNewHost",
                "step": "import",
            }).Errorf("try to import first host [%d] is failed", tHostList[0])
            return err
        }
        if res.(bool) != true {
            log.WithFields(log.Fields{
                "func": "CreateNewHost",
                "step": "import",
            }).Errorf("try to import first host [%d] is failed", tHostList[0])
            return errors.New("result of import host task is false")
        }
        time.Sleep(2*time.Second)
    }

    log.WithFields(log.Fields{
        "func": "CreateNewHost",
        "step": "finish",
    }).Debug("start create new host on new zabbix")
    return nil
}

func SyncHistory(aZDB *ZabbixDB, bZDB *ZabbixDB, hostgroup string, hostIdBegin int) error {
    log.WithFields(log.Fields{
        "func": "SyncHistory",
        "step": "start",
    }).Debug("start sync old history to new zabbix")

    hMapList, err := aZDB.GetHostMapList(hostgroup, hostIdBegin)
    if err != nil {
        return err
    }
    for _, hMap := range hMapList {
        for _, hTable := range HistoryTables {
            for hostid, host := range hMap {
                err := aZDB.SyncHistoryToOne(bZDB, hTable, hostid, host)
                if err != nil {
                    return err
                }
                log.WithFields(log.Fields{
                    "func": "CreateNewHost",
                    "step": "sync.done",
                }).Debugf("done sync history table [%s]: host [%s] hostid [%d]", hTable, host, hostid)
            }
        }
    }

    log.WithFields(log.Fields{
        "func": "SyncHistory",
        "step": "finish",
    }).Debug("start sync old history to new zabbix")
    return nil
}

func SyncTrends(aZDB *ZabbixDB, bZDB *ZabbixDB, hostgroup string, hostIdBegin int) error {
    log.WithFields(log.Fields{
        "func": "SyncTrends",
        "step": "start",
    }).Debug("start sync old trends to new zabbix")

    hMapList, err := aZDB.GetHostMapList(hostgroup, hostIdBegin)
    if err != nil {
        return err
    }
    for _, hMap := range hMapList {
        for _, sTable := range TrendsTables {
            for hostid, host := range hMap {
                err := aZDB.SyncTrendsToOne(bZDB, sTable, hostid, host)
                if err != nil {
                    return err
                }
                log.WithFields(log.Fields{
                    "func": "SyncTrends",
                    "step": "sync.done",
                }).Debugf("done sync trends table [%s]: host [%s] hostid [%d]", sTable, host, hostid)
            }
        }
    }

    log.WithFields(log.Fields{
        "func": "SyncTrends",
        "step": "finish",
    }).Debug("finish sync old trends to new zabbix")
    return nil
}

func CheckHostGroup(aZAPI, bZAPI *ZabbixAPI) (bool, error) {
    aParams := make(map[string]interface{}, 0)
    aFilter := make(map[string]interface{}, 0)
    aParams["output"] = "extend"
    aParams["filter"] = aFilter
    aZHostGroupList, err := aZAPI.HostGroup("get", aParams)
    if err != nil {
        return false, err
    }

    bParams := make(map[string]interface{}, 0)
    bFilter := make(map[string]interface{}, 0)
    bParams["output"] = "extend"
    bParams["filter"] = bFilter
    bZHostGroupList, err := bZAPI.HostGroup("get", bParams)
    if err != nil {
        return false, err
    }

    mFilter := []string {"groupid", "internal"}
    FilterZUM(aZHostGroupList, mFilter)
    FilterZUM(bZHostGroupList, mFilter)
    isSame, err := DiffUnitList(aZHostGroupList, bZHostGroupList, true)
    if err != nil {
        return false, err
    }

    return isSame, nil
}

func CheckHost(aZAPI, bZAPI *ZabbixAPI, hostgroup string) (bool, error) {
    aParams := make(map[string]interface{}, 0)
    aFilter := make(map[string]interface{}, 0)
    aParams["output"] = []string{"host"}
    aParams["selectGroups"] = hostgroup
    aParams["filter"] = aFilter
    aZHostList, err := aZAPI.Host("get", aParams)
    if err != nil {
        return false, err
    }

    bParams := make(map[string]interface{}, 0)
    bFilter := make(map[string]interface{}, 0)
    bParams["output"] = []string{"host"}
    bParams["selectGroups"] = hostgroup
    bParams["filter"] = bFilter
    bZHostList, err := bZAPI.Host("get", bParams)
    if err != nil {
        return false, err
    }

    mFilter := []string {"groups", "hostid"}
    FilterZUM(aZHostList, mFilter)
    FilterZUM(bZHostList, mFilter)
    isSame, err := DiffUnitList(aZHostList, bZHostList, true)
    if err != nil {
        return false, err
    }

    return isSame, nil
}

func CheckItem(aZAPI, bZAPI *ZabbixAPI, host string) (bool, error) {
    aParams := make(map[string]interface{}, 0)
    aParams["output"] = "key_"
    aParams["host"] = host
    aParams["sortfield"] = "key_"
    aZItemList, err := aZAPI.Item("get", aParams)
    if err != nil {
        return false, err
    }

    bParams := make(map[string]interface{}, 0)
    bParams["output"] = "key_"
    bParams["host"] = host
    bParams["sortfield"] = "key_"
    bZItemList, err := bZAPI.Item("get", bParams)
    if err != nil {
        return false, err
    }

    mFilter := []string {"itemid"}
    FilterZUM(aZItemList, mFilter)
    FilterZUM(bZItemList, mFilter)
    isSame, err := DiffUnitList(aZItemList, bZItemList, true)
    if err != nil {
        return false, err
    }

    return isSame, nil
}

func CheckItemGroup(aZAPI, bZAPI *ZabbixAPI, hostgroup string) (bool, error) {
    aParams := make(map[string]interface{}, 0)
    aParams["output"] = []string{"host"}
    aParams["selectGroups"] = hostgroup
    aZHostList, err := aZAPI.Host("get", aParams)
    if err != nil {
        return false, err
    }
    aHostList := make([]string, len(aZHostList))
    for inx, val := range aZHostList {
        for mKey, mVal := range val {
            if mKey != "host" {
                continue
            }
            _mVal, ok := mVal.(string)
            if !ok {
                return false, errors.New("convert ZUnitMap is failed")
            }
            aHostList[inx] = _mVal
        }
    }

    isSame := true
    for _, host := range aHostList {
        var innerIsSame bool
        fmt.Printf("check for host [%s] ...\n", host)
        innerIsSame, err = CheckItem(aZAPI, bZAPI, host)
        if err != nil {
            return false, err
        }
        if !innerIsSame {
            isSame = false
        }
    }

    return isSame, nil
}

func CheckTriggerNum(aZAPI, bZAPI *ZabbixAPI, host string) (bool, error) {
    aParams := make(map[string]interface{}, 0)
    aParams["output"] = "triggerid"
    aParams["host"] = host
    aParams["sortfield"] = "triggerid"
    aZTriggerList, err := aZAPI.Trigger("get", aParams)
    if err != nil {
        return false, err
    }

    bParams := make(map[string]interface{}, 0)
    bParams["output"] = "triggerid"
    bParams["host"] = host
    bParams["sortfield"] = "triggerid"
    bZTriggerList, err := bZAPI.Trigger("get", bParams)
    if err != nil {
        return false, err
    }

    return len(aZTriggerList) == len(bZTriggerList), nil
}

func CheckTriggerNumGroup(aZAPI, bZAPI *ZabbixAPI, hostgroup string) (bool, error) {
    aParams := make(map[string]interface{}, 0)
    aParams["output"] = []string{"host"}
    aParams["selectGroups"] = hostgroup
    aZHostList, err := aZAPI.Host("get", aParams)
    if err != nil {
        return false, err
    }
    aHostList := make([]string, len(aZHostList))
    for inx, val := range aZHostList {
        for mKey, mVal := range val {
            if mKey != "host" {
                continue
            }
            _mVal, ok := mVal.(string)
            if !ok {
                return false, errors.New("convert ZUnitMap is failed")
            }
            aHostList[inx] = _mVal
        }
    }

    isSame := true
    for _, host := range aHostList {
        var innerIsSame bool
        innerIsSame, err = CheckTriggerNum(aZAPI, bZAPI, host)
        if err != nil {
            return false, err
        }
        if !innerIsSame {
            isSame = false
        }
    }

    return isSame, nil
}

func CheckValuemap(aZAPI, bZAPI *ZabbixAPI) (bool, error) {
    aParams := make(map[string]interface{}, 0)
    aParams["output"] = "extend"
    aValuemapList, err := aZAPI.Valuemap("get", aParams)
    if err != nil {
        return false, err
    }

    bParams := make(map[string]interface{}, 0)
    bParams["output"] = "extend"
    bValuemapList, err := bZAPI.Valuemap("get", bParams)
    if err != nil {
        return false, err
    }

    mFilter := []string {"valuemapid"}
    FilterZUM(aValuemapList, mFilter)
    FilterZUM(bValuemapList, mFilter)
    isSame, err := DiffUnitList(aValuemapList, bValuemapList, true)
    if err != nil {
        return false, err
    }

    return isSame, nil
}

func CheckMap(aZAPI, bZAPI *ZabbixAPI) (bool, error) {
    return false, errors.New("not support for map check, please manually")
}
