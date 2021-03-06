package main

import (
    "database/sql"
    "errors"
    "fmt"
    "time"

    log "github.com/sirupsen/logrus"
    _ "github.com/go-sql-driver/mysql"
    _ "github.com/lib/pq"
)

type ZabbixDB struct {
    host        string
    port        int
    user        string
    password    string
    DBDriver    string
    Database    string
    DBVersion   int
    DB          *sql.DB
}

type HostMap map[int]string
type ItemMap map[int]string

func NewZabbixDB(dbDriver string, host string, port int, user string, password string, database string) (*ZabbixDB, error) {
    var dsn string
    switch dbDriver {
    case "mysql":
        dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, database)
    case "postgres":
        dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, database)
    default:
        return &ZabbixDB{}, errors.New("cannot support for the db driver")
    }

    db, err := sql.Open(dbDriver, dsn)
    if err != nil {
        return &ZabbixDB{}, err
    }
    if err := db.Ping(); err != nil {
        return &ZabbixDB{}, errors.New("ping testing fail, open database fail")
    }

    var dbVersion int
    row := db.QueryRow("select floor(mandatory/1000000) from dbversion limit 1;")
    err = row.Scan(&dbVersion)
    if err != nil {
        return &ZabbixDB{}, errors.New("get dbversion failed")
    }



    return &ZabbixDB{
        host: host,
        port: port,
        user: user,
        password: password,
        DBDriver: dbDriver,
        DBVersion: dbVersion,
        Database: database,
        DB: db,
    }, nil
}

func (db *ZabbixDB) Close() error {
    err := db.DB.Close()
    if err != nil {
        return err
    }
    return nil
}

func (db *ZabbixDB) GetTemplateList() ([]int, error) {
    rows, err := db.DB.Query("select hostid from hosts where status = 3 order by hostid")
    if err != nil {
        return []int{}, err
    }
    defer rows.Close()

    res := make([]int, 0)
    for rows.Next() {
        var hostid int
        rows.Scan(&hostid)
        res = append(res, hostid)
    }
    return res, nil
}

func (db *ZabbixDB) GetHostList(hostgroup string, hostIdBegin int) ([]int, error) {
    var rows *sql.Rows
    var err error
    if hostgroup == "" {
        rows, err = db.DB.Query("select hostid from hosts where status != 3 and hostid >= ? order by hostid", hostIdBegin)
    } else {
        rows, err = db.DB.Query(
            `select hg.hostid from hosts_groups hg left join hosts h on hg.hostid = h.hostid where h.status = 0 and h.hostid >= ? and hg.groupid in (select groupid from hstgrp g where g.name = ?) order by hg.hostid`, 
            hostIdBegin,
            hostgroup,
        )
    }
    if err != nil {
        return []int{}, err
    }
    defer rows.Close()

    res := make([]int, 0)
    for rows.Next() {
        var hostid int
        rows.Scan(&hostid)
        res = append(res, hostid)
    }
    return res, nil
}

func (db *ZabbixDB) GetHostMapList(hostgroup string, hostIdBegin int, offset uint) ([]HostMap, error) {
    var rows *sql.Rows
    var err error
    if offset == 0 {
        offset = 999999999
    }
    if hostgroup == "" {
        rows, err = db.DB.Query("select hostid, host from hosts where status = 0 and hostid >= ? and h.status != 3 order by hostid limit ?", hostIdBegin, offset)
    } else {
        rows, err = db.DB.Query(`select hg.hostid, h.host from hosts_groups hg left join hosts h on hg.hostid = h.hostid where h.hostid >= ? and h.status != 3 and hg.groupid in (select groupid  from hstgrp g where g.name = ?) order by hg.hostid limit ?`, 
            hostIdBegin,
            hostgroup,
            offset,
        )
    }
    if err != nil {
        return []HostMap{}, err
    }
    defer rows.Close()

    res := make([]HostMap, 0)
    for rows.Next() {
        var hostid int
        var host string
        rows.Scan(&hostid, &host)
        res = append(res, HostMap{hostid: host})
    }
    return res, nil
}

func (db *ZabbixDB) GetItemList(hostid int) ([]int, error) {
    rows, err := db.DB.Query("select itemid from items where flags not in (1,2) and hostid = ? order by itemid", hostid)
    if err != nil {
        return []int{}, err
    }
    defer rows.Close()

    res := make([]int, 0)
    for rows.Next() {
        var itemid int
        rows.Scan(&itemid)
        res = append(res, itemid)
    }
    return res, nil
}

func (db *ZabbixDB) GetItemMap(hostid int) (ItemMap, error) {
    rows, err := db.DB.Query("select itemid, key_ from items where flags not in (1,2) and hostid = ? order by itemid", hostid)
    if err != nil {
        return ItemMap{}, err
    }
    defer rows.Close()

    res := make(ItemMap, 0)
    for rows.Next() {
        var itemid int
        var key_ string
        rows.Scan(&itemid, &key_)
        res[itemid] = key_
    }
    return res, nil
}

func (db *ZabbixDB) MappingItemId(host string, iMap ItemMap) (map[int]int, error) {
    res := make(map[int]int)
    for itemid, key_ := range iMap {
        var row *sql.Row
        sql_ := fmt.Sprintf("select i.itemid from items i left join hosts h on i.hostid = h.hostid where i.flags not in (1,2) and h.host = '%s' and i.key_ = '%s'", host, key_)
        switch db.DBDriver {
        case "mysql":
            row = db.DB.QueryRow(sql_)
        case "postgres":
            row = db.DB.QueryRow(sql_)
        }

        var _itemid int
        err := row.Scan(&_itemid)
        if err != nil {
            switch {
            case err == sql.ErrNoRows:
            case err != nil:
                return map[int]int{}, err
            }
        }
        res[itemid] = _itemid
    }
    return res, nil
}

func (db *ZabbixDB) SyncHistoryToOne(bZDB *ZabbixDB, hTable string, hostid int, host string, offsetDay uint, ignoreErr bool) error {
    var err error
    var value string

    aHostid := hostid
    aHost := host
    aItemList, err := db.GetItemList(aHostid)
    if err != nil {
        return err
    }
    aItemMap, err := db.GetItemMap(aHostid)
    if err != nil {
        return err
    }
    mappingI, err := bZDB.MappingItemId(aHost, aItemMap)
    if err != nil {
        return err
    }

    endClock := time.Now().Unix() - 3600*24*int64(offsetDay)
    limitOffset := 1000

    if hTable != "history_log" {
        sql1 := fmt.Sprintf("select * from %s where itemid = ? and clock < ? limit ? offset ?", hTable)
        var sql2 string
        switch bZDB.DBDriver {
        case "mysql":
            sql2 = fmt.Sprintf("insert into %s values(?, ?, ?, ?)", hTable)
        case "postgres":
            sql2 = fmt.Sprintf("insert into %s values($1, $2, $3, $4)", hTable)
        }

        for _, itemid := range aItemList {
            log.WithFields(log.Fields{
                "func": "ZabbixDB.SyncHistoryToOne",
                "step": "insert",
            }).Tracef("prepare sql hostid [%d] itemid [%d] mapItemid [%d]", aHostid, itemid, mappingI[itemid])

            if val, ok := mappingI[itemid]; !ok || val == 0 {
                log.WithFields(log.Fields{
                    "func": "ZabbixDB.SyncHistoryToOne",
                    "step": "insert",
                }).Errorf("not found itemid mapping for itemid [%d]", itemid)
                continue
            }

            limitStart := 0
            for {
                log.WithFields(log.Fields{
                    "func": "ZabbixDB.SyncHistoryToOne",
                    "step": "select.sql",
                }).Tracef(
                    "prepare sql: itemid [%d] endClock [%d] limitStart [%d] limitOffset [%d]", 
                    itemid,
                    endClock,
                    limitStart,
                    limitOffset,
                )

                aRows, err := db.DB.Query(sql1, itemid, endClock, limitOffset, limitStart)
                if err != nil {
                    return err
                }

                isEmpty := true
                iCount := 0
                for aRows.Next() {
                    isEmpty = false
                    var _itemid int
                    var _clock int
                    var _ns int
                    aRows.Scan(&_itemid, &_clock, &value, &_ns)
                    _, err := bZDB.DB.Exec(
                        sql2, 
                        mappingI[itemid], _clock, value, _ns,
                    )
                    if err != nil {
                        log.WithFields(log.Fields{
                            "func": "ZabbixDB.SyncHistoryToOne",
                            "step": "insert",
                        }).Errorf("try to sync hostid [%d] itemid [%d] is failed", aHostid, itemid)
                        break
                    }
                    iCount++
                }

                log.WithFields(log.Fields{
                    "func": "ZabbixDB.SyncHistoryToOne",
                    "step": "insert",
                }).Tracef("done sync %s hostid [%d] itemid [%d] mapItemid [%d], insert count is %d", hTable, aHostid, itemid, mappingI[itemid], iCount)

                aRows.Close()
                if err != nil {
                    return err
                }
                if isEmpty {
                    break
                }
                limitStart += limitOffset
            }
        }
    } else {
        sql1 := fmt.Sprintf("select * from %s where itemid = ? and clock < ? limit ? offset ?", hTable)
        var sql2 string
        switch bZDB.DBDriver {
        case "mysql":
            sql2 = fmt.Sprintf("insert into %s values(?, ?, ?, ?, ?, ?, ?, ?)", hTable)
        case "postgres":
            sql2 = fmt.Sprintf("insert into %s values($1, $2, $3, $4, $5, $6, $7, $8)", hTable)
        }

        for _, itemid := range aItemList {
            log.WithFields(log.Fields{
                "func": "ZabbixDB.SyncHistoryToOne",
                "step": "insert",
            }).Tracef("prepare sql hostid [%d] itemid [%d] mapItemid [%d]", aHostid, itemid, mappingI[itemid])

            if val, ok := mappingI[itemid]; !ok || val == 0 {
                log.WithFields(log.Fields{
                    "func": "ZabbixDB.SyncHistoryToOne",
                    "step": "insert",
                }).Errorf("not found itemid mapping for itemid [%d]", itemid)
                continue
            }

            limitStart := 0
            for {
                log.WithFields(log.Fields{
                    "func": "ZabbixDB.SyncHistoryToOne",
                    "step": "select.sql",
                }).Tracef(
                    "prepare sql: itemid [%d] endClock [%d] limitStart [%d] limitOffset [%d]", 
                    itemid,
                    endClock,
                    limitStart,
                    limitOffset,
                )

                aRows, err := db.DB.Query(sql1, itemid, endClock, limitOffset, limitStart)
                if err != nil {
                    return err
                }

                isEmpty := true
                iCount := 0
                for aRows.Next() {
                    isEmpty = false
                    var _itemid int
                    var _clock int
                    var _timestamp int
                    var _source string
                    var _severity int
                    var _logeventid int
                    var _ns int
                    aRows.Scan(&_itemid, &_clock, &_timestamp, &_source, &_severity, &value, &_logeventid, &_ns)
                    _, err =bZDB.DB.Exec(
                        sql2, 
                        mappingI[itemid], _clock, _timestamp, _source, _severity, value, _logeventid, _ns,
                    )
                    if err != nil {
                        log.WithFields(log.Fields{
                            "func": "ZabbixDB.SyncHistoryToOne",
                            "step": "insert",
                        }).Errorf("try to sync %s hostid [%d] itemid [%d] is failed", hTable, aHostid, itemid)
                        break
                    }
                    iCount++
                }

                log.WithFields(log.Fields{
                    "func": "ZabbixDB.SyncHistoryToOne",
                    "step": "insert",
                }).Tracef("done sync %s hostid [%d] itemid [%d] mapItemid [%d], insert count is %d", hTable, aHostid, itemid, mappingI[itemid], iCount)

                aRows.Close()
                if err != nil {
                    return err
                }

                if isEmpty {
                    break
                }

                limitStart += limitOffset
            }
        }
    }
    return nil
}

func (db *ZabbixDB) SyncTrendsToOne(bZDB *ZabbixDB, tTable string, hostid int, host string, ignoreErr bool) error {
    var err error

    aHostid := hostid
    aHost := host
    aItemList, err := db.GetItemList(aHostid)
    if err != nil {
        return err
    }
    aItemMap, err := db.GetItemMap(aHostid)
    if err != nil {
        return err
    }

    mappingI, err := bZDB.MappingItemId(aHost, aItemMap)
    if err != nil {
        return err
    }

    sql1 := fmt.Sprintf("select * from %s where itemid = ?", tTable)
    var sql2 string
    switch bZDB.DBDriver {
    case "mysql":
        sql2 = fmt.Sprintf(
            "insert into %s values(?, ?, ?, ?, ?, ?) on duplicate key update num=?, value_min=?, value_avg=?, value_max=?", 
            tTable,
        )
    case "postgres":
        sql2 = fmt.Sprintf(`insert into %s values($1, $2, $3, $4, $5, $6) 
            on conflict(itemid, clock) do update
            set
              num = $7,
              value_min = $8,
              value_avg = $9,
              value_max = $10`, 
            tTable,
        )
    }

    for _, itemid := range aItemList {
        log.WithFields(log.Fields{
            "func": "ZabbixDB.SyncTrendsToOne",
            "step": "insert",
        }).Tracef("prepare sql [%s] hostid [%d] itemid [%d] mapItemid [%d]", sql1, aHostid, itemid, mappingI[itemid])

        if val, ok := mappingI[itemid]; !ok || val == 0 {
            log.WithFields(log.Fields{
                "func": "ZabbixDB.SyncTrendsToOne",
                "step": "insert",
            }).Errorf("not found itemid mapping for itemid [%d]", itemid)
            continue
        }

        aRows, err := db.DB.Query(sql1, itemid)
        if err != nil {
            if ignoreErr {
                log.WithFields(log.Fields{
                    "func": "ZabbixDB.SyncTrendsToOne",
                    "step": "insert",
                }).Error(err)
                log.WithFields(log.Fields{
                    "func": "ZabbixDB.SyncTrendsToOne",
                    "step": "insert",
                }).Info("ignore error ...")
                continue
            }
            return err
        }
        defer aRows.Close()

        iCount := 0
        for aRows.Next() {
            var _itemid int
            var _clock int
            var _num int
            var _value_min string
            var _value_avg string
            var _value_max string
            aRows.Scan(&_itemid, &_clock, &_num, &_value_min, &_value_avg, &_value_max)
            _, err := bZDB.DB.Exec(
                sql2, 
                mappingI[itemid], _clock, _num, _value_min, _value_avg, _value_max,
                _num, _value_min, _value_avg, _value_max,
            )
            if err != nil {
                log.WithFields(log.Fields{
                    "func": "ZabbixDB.SyncTrendsToOne",
                    "step": "insert",
                }).Errorf("try to sync %s hostid [%d] itemid [%d] mapItemid [%d] is failed", tTable, aHostid, itemid, mappingI[itemid])
                if ignoreErr {
                    log.WithFields(log.Fields{
                        "func": "ZabbixDB.SyncTrendsToOne",
                        "step": "insert",
                    }).Info("ignore error ...")
                    break
                }
                return err
            }
            iCount++
        }

        log.WithFields(log.Fields{
            "func": "ZabbixDB.SyncTrendsToOne",
            "step": "insert",
        }).Tracef("done sync %s hostid [%d] itemid [%d] mapItemid [%d], insert count is %d", tTable, aHostid, itemid, mappingI[itemid], iCount)

    }
    return nil
}