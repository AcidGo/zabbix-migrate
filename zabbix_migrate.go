package main

import (
    "errors"
    "fmt"
    "time"
    log "github.com/sirupsen/logrus"
)

func CreateNewHostGroup(aZAPI, bZAPI *ZabbixAPI) error {
    log.Debug("start create new host group on new zabbix")

    aParams := make(map[string]interface{}, 0)
    aFilter := make(map[string]interface{}, 0)
    aParams["output"] = "extend"
    aParams["filter"] = aFilter
    aZHostGroupLst, err := aZAPI.HostGroup("get", aParams)
    if err != nil {
        return err
    }

    bParams := make(map[string]interface{}, 0)
    bFilter := make(map[string]interface{}, 0)
    bParams["output"] = "extend"
    bParams["filter"] = bFilter
    bZHostGroupLst, err := bZAPI.HostGroup("get", bParams)
    if err != nil {
        return err
    }

    tParams := make(map[string]interface{}, 0)
    for _, aZHostGroup := range aZHostGroupLst {
        hasExists := false
        for _, bZHostGroup := range bZHostGroupLst {
            if bZHostGroup["name"] == aZHostGroup["name"] {
                hasExists = true
                break
            }
        }
        if hasExists {
            log.Info(fmt.Sprintf("host group [%s] alread exists", aZHostGroup["name"]))
            continue
        }
        tParams["name"] = aZHostGroup["name"]
        _, err := bZAPI.HostGroup("create", tParams)
        if err != nil {
            return err
        }
    }

    log.Debug("finish create new host group on new zabbix")
    return nil
}

func CleanNewTemplate(bZAPI *ZabbixAPI, bZDB *ZabbixDB) error {
    log.Debug("start clean all template on new zabbix")

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
            log.Error(fmt.Sprintf("try to delete first template [%d] is failed", tTemplateList[0]))
            return err
        }
        time.Sleep(2*time.Second)
    }

    log.Debug("finish clean all template on new zabbix")
    return nil
}

func CreateNewTemplate(aZAPI *ZabbixAPI, aZDB *ZabbixDB, bZAPI *ZabbixAPI) error {
    log.Debug("start create new template on new zabbix")

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
            log.Error(fmt.Sprintf("try to export first template [%d] is failed", tTemplateList[0]))
            return err
        }

        bParams := make(map[string]interface{}, 0)
        bRules := make(map[string]interface{}, 0)
        bRules["groups"] = map[string]bool{
            "createMissing": true,
        }
        bRules["hosts"] = map[string]bool{
            "createMissing": false,
            "updateExisting": false,
        }
        bRules["templates"] = map[string]bool{
            "createMissing": true,
            "updateExisting": true,
        }
        bRules["templateScreens"] = map[string]bool{
            "createMissing": true,
            "updateExisting": true,
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
            "createMissing": true,
            "updateExisting": true,
            "deleteMissing": false,
        }
        bRules["discoveryRules"] = map[string]bool{
            "createMissing": true,
            "updateExisting": true,
            "deleteMissing": false,
        }
        bRules["triggers"] = map[string]bool{
            "createMissing": true,
            "updateExisting": true,
            "deleteMissing": false,
        }
        bRules["graphs"] = map[string]bool{
            "createMissing": true,
            "updateExisting": true,
            "deleteMissing": false,
        }
        bRules["httptests"] = map[string]bool{
            "createMissing": true,
            "updateExisting": true,
            "deleteMissing": false,
        }
        bRules["screens"] = map[string]bool{
            "createMissing": false,
            "updateExisting": false,
        }
        bRules["maps"] = map[string]bool{
            "createMissing": false,
            "updateExisting": false,
        }
        bRules["images"] = map[string]bool{
            "createMissing": false,
            "updateExisting": false,
        }
        bRules["valueMaps"] = map[string]bool{
            "createMissing": false,
            "updateExisting": true,
        }
        bParams["rules"] = bRules
        bParams["format"] = "xml"
        bParams["source"] = aTemplateExport
        res, err := bZAPI.Configuration("import", bParams)
        if err != nil {
            log.Error(fmt.Sprintf("try to import first template [%d] is failed", tTemplateList[0]))
            return err
        }
        if res.(bool) != true {
            log.Error(fmt.Sprintf("try to import first template [%d] is failed", tTemplateList[0]))
            return errors.New("result of import template task is false")
        }
        time.Sleep(2*time.Second)
    }

    log.Debug("finish creat new template on new zabbix")
    return nil
}

func CreateNewHost(aZAPI *ZabbixAPI, aZDB *ZabbixDB, bZAPI *ZabbixAPI, hostgroup string, hostIdBegin int) error {
    log.Debug("start create new host on new zabbix")

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
                log.Error(fmt.Sprintf("try to export first host [%d] is failed", tHostList[0]))
            } else {
                log.Error(fmt.Sprintf("UNKNOWN"))
            }
            return err
        }

        bParams := make(map[string]interface{}, 0)
        bRules := make(map[string]interface{}, 0)
        bRules["groups"] = map[string]bool{
            "createMissing": true,
        }
        bRules["hosts"] = map[string]bool{
            "createMissing": true,
            "updateExisting": true,
        }
        bRules["templates"] = map[string]bool{
            "createMissing": false,
            "updateExisting": false,
        }
        bRules["templateScreens"] = map[string]bool{
            "createMissing": false,
            "updateExisting": false,
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
            "createMissing": true,
            "updateExisting": true,
            "deleteMissing": false,
        }
        bRules["discoveryRules"] = map[string]bool{
            "createMissing": true,
            "updateExisting": true,
            "deleteMissing": false,
        }
        bRules["triggers"] = map[string]bool{
            "createMissing": true,
            "updateExisting": true,
            "deleteMissing": false,
        }
        bRules["graphs"] = map[string]bool{
            "createMissing": true,
            "updateExisting": true,
            "deleteMissing": false,
        }
        bRules["httptests"] = map[string]bool{
            "createMissing": true,
            "updateExisting": true,
            "deleteMissing": false,
        }
        bRules["screens"] = map[string]bool{
            "createMissing": false,
            "updateExisting": false,
        }
        bRules["maps"] = map[string]bool{
            "createMissing": false,
            "updateExisting": false,
        }
        bRules["images"] = map[string]bool{
            "createMissing": false,
            "updateExisting": false,
        }
        bRules["valueMaps"] = map[string]bool{
            "createMissing": false,
            "updateExisting": true,
        }
        bParams["rules"] = bRules
        bParams["format"] = "xml"
        bParams["source"] = aHostExport
        res, err := bZAPI.Configuration("import", bParams)
        if err != nil {
            log.Error(fmt.Sprintf("try to import first host [%d] is failed", tHostList[0]))
            return err
        }
        if res.(bool) != true {
            log.Error(fmt.Sprintf("try to import first host [%d] is failed", tHostList[0]))
            return errors.New("result of import host task is false")
        }
        time.Sleep(2*time.Second)
    }

    log.Debug("finish creat new host on new zabbix")
    return nil
}

func SyncHistory(aZDB *ZabbixDB, bZDB *ZabbixDB) error {
    return nil
}