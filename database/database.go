package database

import (
	"bufio"
	"encoding/gob"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

type DB struct {
	tableMutex *sync.RWMutex
	cnstMutex  *sync.RWMutex
	Tables     map[string]Table
	DSs        map[string]reflect.Type
}

// type Datastructure struct {
// 	ConstFuncs map[string]bool
// 	Initializer
// }

func (db DB) save(dirname string) error {
	db.tableMutex.Lock()
	defer db.tableMutex.Unlock()

	err := os.RemoveAll(dirname)
	if err != nil {
		return err
	}
	err = os.Mkdir(dirname, 0777)
	if err != nil {
		return err
	}
	for key, table := range db.Tables {
		f, err := os.Create(dirname + "/" + key)
		if err != nil {
			return err
		}
		enc := gob.NewEncoder(f)
		err = enc.Encode(table)
		f.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) load(dirname string) error {
	db.tableMutex.Lock()
	defer db.tableMutex.Unlock()

	db.Tables = make(map[string]Table)
	err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			fmt.Println(path)
			defer f.Close()
			q := Table{}
			dec := gob.NewDecoder(f)
			if err := dec.Decode(&q); err != nil {
				return err
			}
			q.mutex = &sync.Mutex{}
			fmt.Println(q)
			fmt.Println(info.Name())
			db.Tables[info.Name()] = q
		}
		return nil
	})
	return err
}

func (db DB) Register(ds interface{}) {
	db.RegisterByName(reflect.TypeOf(ds).Name(), ds)
}

func (db DB) RegisterByName(dsname string, ds interface{}) {
	gob.Register(ds)

	// if _, ok := dst.MethodByName("Init"); !ok {
	// 	return errors.New("not fount Init function")
	// }

	db.cnstMutex.Lock()
	defer db.cnstMutex.Unlock()
	db.DSs[dsname] = reflect.TypeOf(ds)
}

func (db DB) addTable(name string, table Table) error {
	db.tableMutex.Lock()
	defer db.tableMutex.Unlock()
	_, ok := db.Tables[name]
	if ok {
		return fmt.Errorf("table named %s already exists", name)
	}
	db.Tables[name] = table
	return nil
}

func getFuncMap(v reflect.Value, t reflect.Type) map[string]interface{} {
	funcMap := map[string]interface{}{}
	for i := 0; i < v.NumMethod(); i++ {
		m := t.Method(i)
		ok := true
		fmt.Println(m.Name)
		for j := 1; j < m.Type.NumIn(); j++ {
			k := m.Type.In(j).Kind()
			//fmt.Println(k.String())
			if k != reflect.Int && k != reflect.String {
				ok = false
			}
		}
		//fmt.Println(ok)
		if ok {
			funcMap[m.Name] = v.Method(i).Interface()
		}
	}
	return funcMap
}

func (db DB) mkTable(dsname string, tbname string, args []string) error {
	dst, ok := db.DSs[dsname]
	if !ok {
		return fmt.Errorf("there doesnt exist datastructure:%s", dsname)
	}

	ds := reflect.New(dst)

	table := Table{
		Ds: ds.Interface(),
		//Funcs: getFuncMap(ds, reflect.PtrTo(dst)),
		//PtrFuncs: getFuncMap(outs[0].Addr(), reflect.PtrTo(reflect.TypeOf(cns).Out(0))),
		mutex: &sync.Mutex{},
	}

	_, err := callInterfaceFunc(ds.MethodByName("Init").Interface(), args) // return value must be (obj, error)
	if err != nil {
		return err
	}

	db.tableMutex.Lock()
	db.Tables[tbname] = table
	db.tableMutex.Unlock()
	return nil
}

func (db DB) delTable(tbname string) {
	db.tableMutex.Lock()
	delete(db.Tables, tbname)
	db.tableMutex.Unlock()
}

func MakeDB() DB {
	return DB{
		tableMutex: &sync.RWMutex{},
		cnstMutex:  &sync.RWMutex{},
		Tables:     make(map[string]Table),
		DSs:        make(map[string]reflect.Type),
	}
}

//Run Server
func (db DB) Run(listener net.Listener) error {
	fmt.Println("Server running")
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go db.handleConnection(conn)
	}
}

func (db DB) handleConnection(conn net.Conn) {
	const idleTime = time.Second * 100

	defer conn.Close()
	conn.SetDeadline(time.Now().Add(idleTime))
	fmt.Println("connected")

	buf := bufio.NewReader(conn)

	for {
		str, err := buf.ReadString('\n')
		conn.SetDeadline(time.Now().Add(idleTime))
		if err != nil {
			fmt.Println(err)
			break
		}

		ret, err := db.parseStr(strings.ReplaceAll(str, "\n", ""))
		msg := "ok:" + ret + "\n"
		if err != nil {
			fmt.Println(err)
			msg = "ng:" + err.Error() + "\n"
		}

		_, err = conn.Write([]byte(msg))
		if err != nil {
			fmt.Println(err)
			break
		}
	}
	fmt.Println("disconnected")
}

func callInterfaceFunc(fun interface{}, args []string) ([]reflect.Value, error) {

	fnt := reflect.TypeOf(fun)

	if len(args) != fnt.NumIn() {
		return nil, errors.New("mismatch num of argments")
	}

	//ins := make([]reflect.Type, fnt.NumIn())
	ins := make([]reflect.Value, fnt.NumIn())
	//ins[0] = reflect.ValueOf(ptr)
	for i := 0; i < fnt.NumIn(); i++ {
		tp := fnt.In(i).Kind()
		switch tp {
		case reflect.Int:
			val, err := strconv.Atoi(args[i])
			if err != nil {
				return nil, fmt.Errorf("arg[%d] must be %s", i, tp.String())
			}
			ins[i] = reflect.ValueOf(val)
		case reflect.String:
			ins[i] = reflect.ValueOf(args[i])
		default:
			return nil, errors.New("invalid function")
		}
	}

	return reflect.ValueOf(fun).Call(ins), nil
}

var errNotEnoughArg = fmt.Errorf("not enough error")
var errCommandNotFound = fmt.Errorf("unknown command")
var errDBNotFound = fmt.Errorf("unknown database")
var errInvalidArguments = fmt.Errorf("invalid arguments")
var successMsg = "success"
var cmds = []string{
	":make", ":save", ":load", ":help", ":tbls", ":dsls", ":delete",
}

func (db *DB) showHelp(args []string) (string, error) {
	if len(args) == 0 {
		return strings.Join(cmds, " "), nil
	} else if len(args) == 1 { // :help Trie
		ds, ok := db.DSs[args[0]]
		if !ok {
			return "", errDBNotFound
		}
		methods := []string{}
		for i := 0; i < ds.NumMethod(); i++ {
			d := ds.Method(i).Name + strings.TrimPrefix(fmt.Sprint(ds.Method(i).Type), "func")
			methods = append(methods, d)
		}
		ret := ";\n" + strings.Join(methods, ";\n")
		return ret, nil
	}

	return "", errInvalidArguments
}

func (db *DB) parseSysCmd(str string) (string, error) {
	args := strings.Split(str, " ")

	switch args[0] {
	case "make": // :make dsname tablename arg1 arg2...
		if len(args) < 3 {
			return "", errNotEnoughArg
		}
		err := db.mkTable(args[1], args[2], args[3:])
		if err != nil {
			return "", err
		}
	case "tbls": // :tbls
		return strings.Join(getKeys(db.Tables), " "), nil
	case "dsls": // :dsls
		return strings.Join(getKeys(db.DSs), " "), nil
	case "delete":
		if len(args) < 2 {
			return "", errNotEnoughArg
		}
		db.delTable(args[1])
	case "help":
		return db.showHelp(args[1:])
	case "save":
		if len(args) < 2 {
			return "", errNotEnoughArg
		}
		err := db.save(args[1])
		if err != nil {
			return "", err
		}
	case "load":
		if len(args) < 2 {
			return "", errNotEnoughArg
		}
		err := db.load(args[1])
		if err != nil {
			return "", err
		}
	default:
		return "", errCommandNotFound
	}
	return successMsg, nil
}

func isError(t reflect.Type) bool {
	errorInterface := reflect.TypeOf((*error)(nil)).Elem()

	return t.Implements(errorInterface)
}

func (db *DB) parseStr(str string) (string, error) {
	fmt.Println(str)
	if str == "" {
		return "", nil
	}
	if str[0] == ':' {
		return db.parseSysCmd(str[1:])
	}

	args := strings.Split(str, " ")

	if len(args) < 2 {
		return "", errNotEnoughArg
	}

	db.tableMutex.RLock()
	t, ok := db.Tables[args[0]]
	db.tableMutex.RUnlock()

	if !ok {
		return "", fmt.Errorf("table:%s not found", args[0])
	}
	query := args[1]

	_, ok = reflect.TypeOf(t.Ds).MethodByName(query)
	if !ok {
		return "", errCommandNotFound
	}
	fun := reflect.ValueOf(t.Ds).MethodByName(query).Interface()

	var outs []reflect.Value
	var err error
	t.mutex.Lock()
	outs, err = callInterfaceFunc(fun, args[2:])
	t.mutex.Unlock()
	if err != nil {
		return "", err
	}

	if len(outs) > 0 {
		if isError(outs[len(outs)-1].Type()) {
			if !outs[len(outs)-1].IsNil() {
				return "", outs[len(outs)-1].Interface().(error)
			}
			outs = outs[:len(outs)-1]
		}
	}
	if len(outs) > 0 {
		switch outs[0].Kind() {
		case reflect.Bool:
			return strconv.FormatBool(outs[0].Interface().(bool)), nil
		case reflect.Int:
			return strconv.Itoa(outs[0].Interface().(int)), nil
		case reflect.String:
			return outs[0].Interface().(string), nil
		default:
			return "", errors.New("invalid return value")
		}
	}
	return successMsg, nil
}
