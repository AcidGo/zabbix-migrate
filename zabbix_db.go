package main

import (
    "database/sql"
    "errors"
    "fmt"

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
    DB          *sql.DB
}

type HostMap map[int]string
type ItemMap map[int]string

func NewZabbixDB(dbDriver string, host string, port int, user string, password string, database string) (*ZabbixDB, error) {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, database)
    db, err := sql.Open(dbDriver, dsn)
    if err != nil {
        return &ZabbixDB{}, err
    }
    if err := db.Ping(); err != nil {
        return &ZabbixDB{}, errors.New("ping testing fail, open database fail")
    }

    return &ZabbixDB{
        host: host,
        port: port,
        user: user,
        password: password,
        DBDriver: dbDriver,
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
        rows, err = db.DB.Query("select hostid from hosts where status != 3 and hostid > ? order by hostid", hostIdBegin)
    } else {
        rows, err = db.DB.Query(
            `select hg.hostid from hosts_groups hg left join hosts h on hg.hostid = h.hostid where h.status = 0 and hg.groupid in (select groupid  from hstgrp g where g.name = ?) order by hg.hostid`, 
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

func (db *ZabbixDB) GetHostMapList(hostgroup string, hostIdBegin int) ([]HostMap, error) {
    var rows *sql.Rows
    var err error
    if hostgroup == "" {
        rows, err = db.DB.Query("select hostid, host from hosts where status = 0 and hostid > ? order by hostid", hostIdBegin)
    } else {
        rows, err = db.DB.Query(`select hg.hostid, h.host from hosts_groups hg left join hosts h on hg.hostid = h.hostid where h.hostid > ? and h.status != 3 and hg.groupid in (select groupid  from hstgrp g where g.name = ?) order by hg.hostid`, 
            hostIdBegin,
            hostgroup,
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
    rows, err := db.DB.Query("select itemid from items where hostid = ? order by itemid", hostid)
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
    rows, err := db.DB.Query("select itemid, key_ from items where flags = 0 and hostid = ? order by itemid", hostid)
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
        row := db.DB.QueryRow("select i.itemid from items i left join hosts h on i.hostid = h.hostid where h.host = ? and i.key_ = ?", host, key_)
        var _itemid int
        err := row.Scan(&_itemid)
        if err != nil {
            return map[int]int{}, err
        }
        res[itemid] = _itemid
    }
    return res, nil
}

func (db *ZabbixDB) SyncHistoryToOne(bZDB *ZabbixDB, hTable string, hostid int, host string) error {
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

    if hTable != "history_log" {
        sql1 := fmt.Sprintf("select * from %s where itemid = ?", hTable)
        sql2 := fmt.Sprintf("insert into %s values(?, ?, ?, ?) on duplicate key update value=?,ns=?", hTable)
        for _, itemid := range aItemList {
            aRows, err := db.DB.Query(sql1, itemid)
            if err != nil {
                return err
            }
            defer aRows.Close()

            for aRows.Next() {
                var _itemid int
                var _clock int
                var _ns int
                aRows.Scan(&_itemid, &_clock, &value, &_ns)
                res, err := bZDB.DB.Exec(
                    sql2, 
                    mappingI[itemid], _clock, value, _ns,
                    value, _ns,
                )
                if err != nil {
                    log.Error(fmt.Sprintf("try to sync hostid [%d] itemid [%d] is failed", aHostid, itemid))
                    return err
                }
                log.Debug(res.RowsAffected())
            }
        }
    } else {
        sql1 := fmt.Sprintf("select * from %s where itemid = ?", hTable)
        sql2 := fmt.Sprintf("insert into %s values(?, ?, ?, ?, ?, ?, ?, ?) on duplicate key update timestamp=?,source=?,severity=?,logeventid=?,ns=?", hTable)
        for _, itemid := range aItemList {
            aRows, err := db.DB.Query(sql1, itemid)
            if err != nil {
                return err
            }
            defer aRows.Close()

            for aRows.Next() {
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
                    _timestamp, _source, _severity, value, _logeventid, _ns,
                )
                if err != nil {
                    log.Error(fmt.Sprintf("try to sync %s hostid [%d] itemid [%d] is failed", hTable, aHostid, itemid))
                    return err
                }
            }
        }
    }
    return nil
}

func (db *ZabbixDB) SyncTrendsToOne(bZDB *ZabbixDB, tTable string, hostid int, host string) error {
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
    sql2 := fmt.Sprintf(
        "insert into %s values(?, ?, ?, ?, ?, ?) on duplicate key update num=?, value_min=?, value_avg=?, value_max=?", 
        tTable,
    )
    for _, itemid := range aItemList {
        aRows, err := db.DB.Query(sql1, itemid)
        if err != nil {
            return err
        }
        defer aRows.Close()

        for aRows.Next() {
            var _itemid int
            var _clock int
            var _num int
            var _value_min string
            var _value_avg string
            var _value_max string
            aRows.Scan(&_itemid, &_clock, &_num, &_value_min, &_value_avg, &_value_max)
            res, err := bZDB.DB.Exec(
                sql2, 
                mappingI[itemid], _clock, _num, _value_min, _value_avg, _value_max,
                _num, _value_min, _value_avg, _value_max,
            )
            if err != nil {
                log.Error(fmt.Sprintf("try to sync %s hostid [%d] itemid [%d] is failed", tTable, aHostid, itemid))
                return err
            }
            log.Debug(res.RowsAffected())
        }
    }
    return nil
}