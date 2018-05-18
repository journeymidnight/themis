package database

import (
	"fmt"
	"time"

	"github.com/coreos/pkg/capnslog"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"

	"github.com/ljjjustin/themis/config"
)

var plog = capnslog.NewPackageLogger("github.com/ljjjustin/themis", "database")

var (
	engine    *xorm.Engine
	allTables []interface{}
)

func Engine(cfg *config.DatabaseConfig) *xorm.Engine {
	var err error

	if engine == nil {
		url := ""
		switch cfg.Driver {
		case "mysql":
			url = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true",
				cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
		case "sqlite3":
			url = cfg.Path
		default:
			plog.Fatal("unsupported database driver, check your configurations")
		}
		engine, err = xorm.NewEngine(cfg.Driver, url)
		if err != nil {
			plog.Fatal(err)
		}
		engine.DatabaseTZ = time.Local
		engine.TZLocation = time.Local
		// fast fail if if we can not connect to database
		err = engine.Ping()
		if err != nil {
			plog.Fatal(err)
		}
		// register and update database models
		err := engine.Sync2(allTables...)
		if err != nil {
			plog.Fatal(err)
		}
	}
	return engine
}

func HostInsert(host *Host) error {
	_, err := engine.Insert(host)
	return err
}

func HostGetAll() ([]*Host, error) {
	hosts := make([]*Host, 0)

	err := engine.Iterate(new(Host),
		func(i int, bean interface{}) error {
			host := bean.(*Host)
			hosts = append(hosts, host)
			return nil
		})
	return hosts, err
}

func HostGetById(id int) (*Host, error) {
	var host = Host{Id: id}

	exist, err := engine.Get(&host)
	if err != nil {
		return nil, err
	} else if exist {
		return &host, nil
	} else {
		return nil, nil
	}
}

func HostGetByName(hostname string) (*Host, error) {
	var host = Host{Name: hostname}

	exist, err := engine.Get(&host)
	if err != nil {
		return nil, err
	} else if exist {
		return &host, nil
	} else {
		return nil, nil
	}
}

func HostUpdate(id int, host *Host) error {
	_, err := engine.ID(id).Update(host)
	return err
}

/*func HostUpdateFields(host *Host, fields ...string) error {
	_, err := engine.ID(host.Id).Cols(fields...).Update(host)
	return err
}*/

func HostUpdateFields(host *Host, fields ...string) error {

	sql := "update host set updated_at= now()"
	for _, field := range fields{

		var subsql string

		if field == "name" {
			subsql = fmt.Sprint(", name = '", host.Name, "'")
		} else if field == "notified" {
			subsql = fmt.Sprint(", notified = ", host.Notified)
		} else if field == "disabled" {
			subsql = fmt.Sprint(", disabled = ", host.Disabled)
		} else if field == "status" {
			subsql = fmt.Sprint(", status = '", host.Status, "'")
		} else if field == "fenced_times" {
			subsql = fmt.Sprint(", fenced_times = ", host.FencedTimes)
		} else {
			subsql = ""
		}

		sql += subsql
	}

	where := fmt.Sprint(" where id = ", host.Id, ";")

	sql += where

	plog.Debug("sql:", sql)

	res, err := engine.Exec(sql)
	if err != nil {
		return err
	}
	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func HostDelete(id int) error {
	_, err := engine.ID(id).Delete(new(Host))
	return err
}

func StateGetAll(hostId int) ([]*HostState, error) {
	states := make([]*HostState, 0)

	err := engine.Where("host_id=?", hostId).Iterate(new(HostState),
		func(i int, bean interface{}) error {
			state := bean.(*HostState)
			states = append(states, state)
			return nil
		})
	return states, err
}

func StateGetById(id int) (*HostState, error) {
	var state = HostState{Id: id}

	exist, err := engine.Get(&state)
	if err != nil {
		return nil, err
	} else if exist {
		return &state, nil
	} else {
		return nil, nil
	}
}

func StateInsert(state *HostState) error {
	_, err := engine.Insert(state)
	return err
}

func StateUpdate(id int, state *HostState) error {
	_, err := engine.ID(id).Update(state)
	return err
}

func StateUpdateFields(state *HostState, fields ...string) error {
	_, err := engine.ID(state.Id).Cols(fields...).Update(state)
	return err
}

func StateDelete(id int) error {
	_, err := engine.ID(id).Delete(new(HostState))
	return err
}

func FencerGetAll() ([]*HostFencer, error) {
	fencers := make([]*HostFencer, 0)

	err := engine.Iterate(new(HostFencer),
		func(i int, bean interface{}) error {
			fencer := bean.(*HostFencer)
			fencers = append(fencers, fencer)
			return nil
		})
	return fencers, err
}

func FencerGetByHost(hostId int) ([]*HostFencer, error) {
	fencers := make([]*HostFencer, 0)

	err := engine.Where("host_id=?", hostId).Iterate(new(HostFencer),
		func(i int, bean interface{}) error {
			fencer := bean.(*HostFencer)
			fencers = append(fencers, fencer)
			return nil
		})
	return fencers, err
}

func FencerGetById(id int) (*HostFencer, error) {
	var fencer = HostFencer{Id: id}

	exist, err := engine.Get(&fencer)
	if err != nil {
		return nil, err
	} else if exist {
		return &fencer, nil
	} else {
		return nil, nil
	}
}

func FencerInsert(fencer *HostFencer) error {
	_, err := engine.Insert(fencer)
	return err
}

func FencerUpdate(id int, fencer *HostFencer) error {
	_, err := engine.ID(id).Update(fencer)
	return err
}

func FencerDelete(id int) error {
	_, err := engine.ID(id).Delete(new(HostFencer))
	return err
}

func GetLeader(electionName string) (*ElectionRecord, error) {

	var leaders = ElectionRecord{ElectionName: electionName}

	exist, err := engine.Get(&leaders)
	if err != nil {
		return nil, err
	} else if exist {
		return &leaders, nil
	} else {
		return nil, nil
	}
}

func GetFencedTimes() (int64 ,error) {

	var host = Host{}

	fencedTimes, err := engine.SumInt(&host, "fenced_times")

	return fencedTimes, err
}

func SetAllHostDisable() error {

	sql := "update host set disabled = true, updated_at= now()"

	res, err := engine.Exec(sql)
	if err != nil {
		return err
	}
	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func SetAllHostEnable() error {

	sql := "update host set disabled = false, status = 'initializing', notified = false, updated_at= now()"

	res, err := engine.Exec(sql)
	if err != nil {
		return err
	}
	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func CurrentTime() (*time.Time, error) {

	sql := "select now() as time"
	result, err := engine.Query(sql)
	if err != nil {
		return nil, err
	}

	dbtime := string(result[0]["time"][:])

	parseTime, err := time.ParseInLocation("2006-01-02T15:04:05Z", dbtime, time.Local)
	if err != nil {
		return nil, err
	}

	return &parseTime, nil
}
