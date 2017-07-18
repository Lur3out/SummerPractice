package main

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	_ "github.com/lib/pq"
)

// Router :: class for Router type -- Изменить!
type Router struct {
	num   int
	name  string
	login string
	pass  string
	ip    string
	port  int
}

func main() {
	helpArg := flag.Bool("help", false, "a boolean")      // Отображает доступные команды
	ipArg := flag.String("ip", "0.0.0.0", "a string")     // Задает Ip роутера
	allArg := flag.Bool("all", false, "a boolean")        // Включает снятия бэкапа со всех роутеров
	newArg := flag.Bool("new", false, "a boolean")        // Создает новое подключение к роутеру
	loginArg := flag.String("login", "admin", "a string") // Задает логин для подключения к роутеру
	passArg := flag.String("pass", "", "a string")        // Задает пароль для подключения к роутеру
	nameArg := flag.String("name", "Unknown", "a string") // Задает псеводним роутера/список бэкапов/экспорт бэкапа
	hostArg := flag.String("host", "Default", "a string") // Задает хостнейм роутера
	portArg := flag.Int("port", 22, "an int")             // Задает порт SSH соединения
	bkpArg := flag.Bool("bkp", false, "a boolean")        // Включает снятие бэкапов
	makeArg := flag.Bool("make", false, "a boolean")      // Снятие бэкапа
	pathArg := flag.String("path", "./", "a string")      // Указывает путь на data.json
	lsArg := flag.Bool("ls", false, "a boolean")          // Выводит список подключенных роутеров
	timeArg := flag.String("time", "20:24", "a string")   // Задает время для экспорта
	dateArg := flag.String("date", "18.07.2017", "a str") // Задает дату для экспорта
	getArg := flag.Bool("get", false, "a boolean")        // Экспортирует бэкап
	lsroutArg := flag.Bool("lsrout", false, "a boolean")  // Выводит список бэкапов по имени
	flag.Args()                                           // Имена роутеров, работает только после флага.

	flag.Parse()

	//[0] - sql.BD address; [1] - path to BAckUps
	params := getData(*pathArg)
	//Массив входщящих параметров(имен роутреов)
	names := flag.Args()

	if *helpArg == true {
		helpPrint()
	}

	if *newArg != false {
		var r Router
		newConnection(r, *nameArg, *hostArg, *ipArg, *loginArg, *portArg, *passArg, params)
		fmt.Println()
	}

	if *makeArg != false {
		if *bkpArg != false {
			if *allArg != false {
				makeAllBackUp(params, *ipArg, *portArg, *loginArg, *passArg, *bkpArg, names)
			} else {
				makeBackUp(params, *ipArg, *portArg, *loginArg, *passArg, *bkpArg, names)
			}
		}
		if *allArg != false && *bkpArg == false {
			makeAllConfig(params, *ipArg, *portArg, *loginArg, *passArg, *bkpArg, names)
		}
		if *bkpArg == false && *allArg == false {
			makeConfig(params, *ipArg, *portArg, *loginArg, *passArg, *bkpArg, names)
		}
	}

	if *lsArg != false {
		current := time.Now()
		fmt.Println("Date: ", current.Format("02.01.2006"))
		fmt.Println("Time: ", current.Format("15:04"))
		printAllConnected(connectDB(params[0]))
	}

	if *lsroutArg != false {
		listRout(*nameArg, params)
	}

	if *getArg != false {
		getBack(*nameArg, *dateArg, *timeArg, params, *bkpArg)
	}

}

//*************************************************************
//---------------------- -HELP---------------------------------
// helpPrint : Функция-принтер
func helpPrint() {
	fmt.Print("----------------------------------------------------------------------------------------------")
	fmt.Println("\nAllowed commands:")
	fmt.Println("\npath:  <Path to data.json>\t\t\t\t\t\t\t\t\t\t\t\t\t// Укажите путь к файлу data.json в формате <Disk>:/~/")
	fmt.Println("\nget: [-bkp] -name <Router's name> -date <dd.mm.yyyy> -time <hh:mm>\t\t\t\t\t\t\t\t//Выбор конфига/бэкапа для экспорта")
	fmt.Println("\nls: \t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t//Выводит список подключенных роутеров.")
	fmt.Println("\nlsrout: -name <Router's name>\t\t\t\t\t\t\t\t\t\t\t\t\t//Выводит список бэкапов для роутера с именем <Name>")
	fmt.Println("\nnew: -name <Router's name> -host <Hostname> -login <Username> -pass <Password> -ip <Router's IP> -port <Port>\t\t//Создать новое подключение")
	fmt.Println("\nmake: [-bkp ] [-all] <Routers' names>\t\t\t\t\t\t\t\t\t\t\t\t//Запуск процедуры снятия бэкапов.\n\n\nАтрибуты:\n\n-bkp\t\t\tПри указании атрибута произведется снятие полного бэкапа. По умолчанию снимается конфигурационный файл.\n\n-all\t\t\tПри указании снимаются бэкапы/конфиги со всех роутеров, подключенных к прогамме\n\n\nВходящие параметры:\n\n<Routers' names>\tЧерез пробел перечилите названия роутеров, с которых произвести снятие бэкапа. При наличии атрибута -all снятие будет произведено со всех.")
	fmt.Println("\n----------------------------------------------------------------------------------------------")
}

//*************************************************************
//-----------------------------SSH-----------------------------

// sshRouter : Функция, возвращающая ssh-клиент для роутера
func sshRouter(login string, pass string, ip string, port int) *ssh.Client {
	config := &ssh.ClientConfig{
		User: login,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	addr := fmt.Sprintf("%s:%d", ip, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		fmt.Printf("Failed to dial: %s", err)
	}
	fmt.Println("Successfully connected to ", ip, ":", port)

	session, err := client.NewSession()
	if err != nil {
		fmt.Printf("Failed to create a new session: %s", err)
	}
	defer session.Close()

	b, err := session.CombinedOutput("/system backup save name=BackUp dont-encrypt=yes") // /system backup save name=BackUp dont-encrypt=yes
	if err != nil {
		fmt.Printf("Failed to send output command: %s", err)
	}
	fmt.Print(string(b))

	session2, err := client.NewSession()
	if err != nil {
		fmt.Printf("Failed to create a new session: %s", err)
	}
	defer session2.Close()

	c, err := session2.CombinedOutput("/export file=config.rsc") // /export file=config.rsc
	if err != nil {
		fmt.Printf("Failed to send output command: %s", err)
	}
	fmt.Print(string(c))

	return client
}

//*************************************************************
//---------------------------SFTP------------------------------

// sftpRouter : Функция, создающая sftp соединение и экспортирующая backup\rsc.
func sftpRouter(client *ssh.Client, bkp bool, path string) {
	sftp, err := sftp.NewClient(client)
	if err != nil {
		fmt.Printf("#172 Failed to create new sftp-client: %s", err)
	}
	defer sftp.Close()

	srcPath := "/"
	filename := "BackUp.backup"
	config := "config.rsc"

	if bkp == true {

		// Open the source file
		srcFile, err := sftp.Open(srcPath + filename)
		if err != nil {
			fmt.Printf("#185 Failed to open backup file on router: %s", err)
		}
		defer srcFile.Close()

		// Create the destination file
		dstFile, err := os.Create(path + filename)
		if err != nil {
			fmt.Printf("#192 Failed to create destination file: %s", err)
		}
		defer dstFile.Close()

		// Copy the file
		srcFile.WriteTo(dstFile)
	} else {

		// Open the source file
		srcFile2, err := sftp.Open(srcPath + config)
		if err != nil {
			fmt.Printf("#203 Failed to open config file on router: %s", err)
		}
		defer srcFile2.Close()

		dstFile2, err := os.Create(path + config)
		if err != nil {
			fmt.Printf("#209 Failed to create destination file: %s", err)
		}
		defer dstFile2.Close()
		// Copy the file
		srcFile2.WriteTo(dstFile2)
	}

}

//*************************************************************
//-------------------------SQL---------------------------------

// connectDB : Функция, создающая подключение к PostgreSql по настройкам из файла data.json
func connectDB(settings string) *sql.DB {

	// Подключение к БД
	db, err := sql.Open("postgres", settings)
	if err != nil {
		fmt.Println("#227 sql.DB.Open()") // log.Fatal(err)
	}
	return db
}

// addNewRouter : Функция, добавляющая данные нового подключения в PostgreSql
func addNewRouter(r Router, settings string) {
	db, err := sql.Open("postgres", settings)

	var lastInsertID int
	err = db.QueryRow("INSERT INTO routers(name, login, pass, ip, port) VALUES($1,$2,$3,$4,$5) returning routerid;", r.name, r.login, r.pass, r.ip, r.port).Scan(&lastInsertID)
	if err != nil {
		fmt.Println("QueryRow err #370 string")
	}

	defer db.Close()
}

// addNewHash : Функция, добавляющая хэши нового backup/rsc в БД
func addNewHash(md5 string, sha1 string, name string, bkp bool, params [2]string) {
	current := time.Now()
	time := current.Format("15:04")
	date := current.Format("02.01.2006")

	db, err := sql.Open("postgres", params[0])
	if err != nil {
		fmt.Println("#253 sql.DB.Open()") // log.Fatal(err)
	}

	defer db.Close()

	if bkp == true {
		var lastInsertID int
		err = db.QueryRow("INSERT INTO hashs(name, date, time, md5bkp, sha1bkp, md5cfg, sha1cfg) VALUES($1,$2,$3,$4,$5,$6,$7) returning hashid;", name, date, time, md5, sha1, "", "").Scan(&lastInsertID)
		if err != nil {
			fmt.Println("QueryRow err # 262 string")
		}
	} else {
		var lastInsertID int
		err = db.QueryRow("INSERT INTO hashs(name, date, time, md5bkp, sha1bkp, md5cfg, sha1cfg) VALUES($1,$2,$3,$4,$5,$6,$7) returning hashid;", name, date, time, "", "", md5, sha1).Scan(&lastInsertID)
		if err != nil {
			fmt.Println("QueryRow err # 268 string")
		}
	}

}

// addNewFile : Функция, добавляющая новые файлы в БД
func addNewFile(params [2]string, r Router, bkp bool) {
	current := time.Now()
	time := current.Format("15:04")
	date := current.Format("02.01.2006")

	db, err := sql.Open("postgres", params[0])
	if err != nil {
		fmt.Println("#282 sql.DB.Open()")
	}

	defer db.Close()

	if bkp == true {
		file, err := os.Open(params[1] + "BackUp.backup")
		if err != nil {
			fmt.Println("#290 string")
		}
		fileInfo, _ := file.Stat()
		fileSize := fileInfo.Size()
		bytes := make([]byte, fileSize)

		var lastInsertID int
		err = db.QueryRow("INSERT INTO backups(name, date, time, bkp, cfg) VALUES($1,$2,$3,$4,$5) returning backupid;", r.name, date, time, bytes, "").Scan(&lastInsertID)
		if err != nil {
			fmt.Println("QueryRow err # 299 string")
		}

	} else {
		file, err := os.Open(params[1] + "config.rsc")
		if err != nil {
			fmt.Println("#305 string")
		}
		fileInfo, _ := file.Stat()
		fileSize := fileInfo.Size()
		bytes := make([]byte, fileSize)

		var lastInsertID int
		err = db.QueryRow("INSERT INTO backups(name, date, time, bkp, cfg) VALUES($1,$2,$3,$4,$5) returning backupid;", r.name, date, time, "", bytes).Scan(&lastInsertID)
		if err != nil {
			fmt.Println("QueryRow err # 314 string")
		}

	}
}

// printAllConnected : Функция, выводящая все подключенные роутеры
func printAllConnected(db *sql.DB) {
	rows, err := db.Query("SELECT * FROM routers")
	if err != nil {
		fmt.Println("Query error #452 string")
	}

	for rows.Next() {
		var testid int
		name := ""
		ip := ""
		port := 0
		login := ""
		pass := ""
		err = rows.Scan(&routerid, &name, &login, &pass, &ip, &port)
		if err != nil {
			fmt.Println("Scan error #336 string")
		}
		fmt.Println("routerid | name | login | pass | ip | port")
		fmt.Printf("%3v | %8v | %8v | %6v | %8v | %3v\n", routerid, name, login, pass, ip, port)

		//*************************************
	}
	defer db.Close()
}

/*// deleteRow : Delete row
func deleteRow(index int, settings string) {
	fmt.Println("# Deleting")

	db := connectDB(settings)
	stmt, _ := db.Prepare("delete from router where test_id=$1")

	res, _ := stmt.Exec(index)

	affect, _ := res.RowsAffected()

	fmt.Println(affect, "rows changed")
}*/

// routerData : Функция, создающая бэкапы для всех роутеров
func routerData(db *sql.DB, bkp bool, params [2]string) {
	var r Router
	rows, err := db.Query("SELECT * FROM routers")
	if err != nil {
		fmt.Println("Query error #365 string")
	}

	for rows.Next() {
		var routerid int
		name := ""
		ip := ""
		port := 0
		login := ""
		pass := ""
		err = rows.Scan(&routerid, &name, &login, &pass, &ip, &port)
		if err != nil {
			fmt.Println("Scan error #464 string")
		}

		fmt.Println("routerid | name | login | pass | ip | port")
		fmt.Printf("%3v | %8v | %6v | %8v | %6v | %3v\n", routerid, name, login, pass, ip, port)

		r.num = routerid
		r.name = name
		r.ip = ip
		r.port = port
		r.login = login
		r.pass = pass

		sftpRouter(sshRouter(r.login, r.pass, r.ip, r.port), bkp, params[1])
		sqlRouter(params, r, bkp)

	}
	defer db.Close()
}

// sqlRouter : Функция добавляющая хэш и файлы в БД
func sqlRouter(params [2]string, r Router, bkp bool) {

	if bkp == true {
		addNewHash(hashMD5Bkp(params[1]), hashSHA1Bkp(params[1]), r.name, bkp, params)
		addNewFile(params, r, bkp)
	} else {
		addNewHash(hashMD5Cfg(params[1]), hashSHA1Cfg(params[1]), r.name, bkp, params)
		addNewFile(params, r, bkp)
	}
}

// routerDataNm : Функция, создающая бэкапы по именам роутеров
func routerDataNm(db *sql.DB, bkp bool, params [2]string, names []string) {
	var r Router
	rows, err := db.Query("SELECT * FROM routers")
	if err != nil {
		fmt.Println("Query error #414 string")
	}

	for rows.Next() {
		var router int
		name := ""
		ip := ""
		port := 0
		login := ""
		pass := ""
		err = rows.Scan(&routerid, &name, &login, &pass, &ip, &port)
		if err != nil {
			fmt.Println("Scan error #426 string")
		}
		fmt.Println("routerid | name | login | pass | ip | port")
		fmt.Printf("%3v | %8v | %6v | %8v | %6v | %3v\n", routerid, name, login, pass, ip, port)

		r.num = routerid
		r.name = name
		r.ip = ip
		r.port = port
		r.login = login
		r.pass = pass

		for i := 0; i < len(names); i++ {
			if names[i] == name {
				sftpRouter(sshRouter(r.login, r.pass, r.ip, r.port), bkp, params[1])
				sqlRouter(params, r, bkp)
			}
		}
	}
	defer db.Close()
}

// listRout : Функция, выводящая все сделанные бэкапы(по хэшам) по имени роутера
func listRout(name string, params [2]string) {
	db, err := sql.Open("postgres", params[0])
	if err != nil {
		fmt.Println("#452")
	}

	cmd := "SELECT * FROM hashs WHERE name LIKE '" + name + "'"
	rows, err := db.Query(cmd)
	if err != nil {
		fmt.Println("Query error #458 string")
	}

	for rows.Next() {
		var hashid int
		name := ""
		date := ""
		time := ""
		md5bkp := ""
		sha1bkp := ""
		md5cfg := ""
		sha1cfg := ""
		err = rows.Scan(&hashid, &name, &date, &time, &md5bkp, &sha1bkp, &md5cfg, &sha1cfg)
		if err != nil {
			fmt.Println("Scan error #472 string")
		}

		fmt.Println("hashid | name | date | time | MD5 BackUp| SHA-1 BackUp | MD5 config | SHA-1 config")
		fmt.Printf("%3v | %8v | %8v | %4v | %8v | %8v | %8v | %8v\n", hashid, name, date, time, md5bkp, sha1bkp, md5cfg, sha1cfg)
	}
	defer db.Close()
}

//**************************************************************
//----------------------MD5 u SHA-1-----------------------------

// hashMD5Bkp : Функция, создающая hash MD5 BackUp.backup
func hashMD5Bkp(path string) string {
	backup := "BackUp.backup"

	backupFile, err := os.Open(path + backup)
	if err != nil {
		fmt.Print("Не удалось открыть BackUp.backup", err)
	}
	defer backupFile.Close()

	backupHash := md5.New()
	if _, err := io.Copy(backupHash, backupFile); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", backupHash.Sum(nil))
}

// hashMD5Cfg : Функция, создающая hash MD5 config.rsc
func hashMD5Cfg(path string) string {
	config := "config.rsc"

	configFile, err := os.Open(path + config)
	if err != nil {
		fmt.Print("Не удалось открыть config.rsc", err)
	}
	defer configFile.Close()

	configHash := md5.New()
	if _, err := io.Copy(configHash, configFile); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", configHash.Sum(nil))
}

// hashSHA1Bkp : Функция, создающая hash SHA1 BackUp.backup
func hashSHA1Bkp(path string) string {
	backup := "BackUp.backup"

	backupFile, err := os.Open(path + backup)
	if err != nil {
		fmt.Print("Не удалось открыть BackUp.backup", err)
	}
	defer backupFile.Close()

	backupHash := sha1.New()
	if _, err := io.Copy(backupHash, backupFile); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", backupHash.Sum(nil))
}

// hashSHA1Cfg : Функция, создающая hash SHA1 config.rsc
func hashSHA1Cfg(path string) string {
	config := "config.rsc"

	configFile, err := os.Open(path + config)
	if err != nil {
		fmt.Print("Не удалось открыть config.rsc", err)
	}
	defer configFile.Close()

	configHash := sha1.New()
	if _, err := io.Copy(configHash, configFile); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", configHash.Sum(nil))
}

//**************************************************************
//----------------------DATA.JSON-------------------------------
// getData : Функция, загружающая информацию с файла path.Json && connection.Json
func getData(path string) [2]string {
	//	str[0] - info for connection to sql.BD
	//	str[1] - path to backup's save zone

	data := "data.json"

	var str [2]string
	file, err := os.Open(path + data)
	if err != nil {
		fmt.Println("Не удалось открыть файл data.json")
	}
	defer file.Close()
	f := bufio.NewReader(file)
	for i := 0; i < len(str); i++ {
		str[i], _ = f.ReadString('\n')
	}

	return str
}

//**************************************************************
//-------------------NEW CONNECTION(ROUTER)---------------------
// newConnection : Функция, создающая новое подключение к роутеру -- Изменить! Добавить!
func newConnection(r Router, name string, hostname string, ip string, login string, port int, pass string, params [2]string) {
	r.name = name
	r.ip = ip
	r.login = login
	r.port = port
	r.pass = pass
	fmt.Println("----------------------------------------------------------------------------------------------")
	fmt.Println("\nРоутер\t", r.name, "\t", r.ip, "\t", r.login, "\t", r.pass, "\t", r.port, "\t\tбыл добавлен")
	fmt.Println("\n----------------------------------------------------------------------------------------------")

	addNewRouter(r, params[0])
}

// routerPrint : Функция-принтер для Router -- Изменить! Добавить!
func routerPrint(r Router, i int) {
	fmt.Println("#", i, "\tName: ", r.name, "\tLogin: ", r.login, "\tPassword: ", r.pass, "\tIp: ", r.ip, "\tPort: ", r.port)
}

//***************************************************************
//----------------------Type Of BackUps--------------------------
// makeAllBackUp : -make -all -bkp
func makeAllBackUp(params [2]string, ip string, port int, login string, pass string, bkp bool, names []string) {
	//Снятие полных бэкапов со всех роутеров
	routerData(connectDB(params[0]), bkp, params)
}

// makeBackUp : -make -bkp <names>
func makeBackUp(params [2]string, ip string, port int, login string, pass string, bkp bool, names []string) {
	//Снятие бэкапов с перечисленных роутеров

	routerDataNm(connectDB(params[0]), bkp, params, names)
}

// makeAllConfig : -make -all
func makeAllConfig(params [2]string, ip string, port int, login string, pass string, bkp bool, names []string) {
	//Снятие конфигов со всех роутеров
	routerData(connectDB(params[0]), bkp, params)
}

// makeConfig : -make <names>
func makeConfig(params [2]string, ip string, port int, login string, pass string, bkp bool, names []string) {
	//Снятие конфигов с перечисленных роутеров

	routerDataNm(connectDB(params[0]), bkp, params, names)
}

//***************************************************************
//--------------------------GET BACK-----------------------------
// getBack : Функция, возвращающая бэкап
func getBack(_name string, _date string, _time string, params [2]string, bkp bool) {

	cmd := "SELECT * FROM backups WHERE name LIKE '" + _name + "'"
	db := connectDB(params[0])
	rows, err := db.Query(cmd)
	if err != nil {
		fmt.Println("Query error #635 string")
	}

	for rows.Next() {

		var backupid int
		var time string
		var bkpfile []byte
		var cfgfile []byte
		var name string
		var date string
		err = rows.Scan(&backupid, &name, &date, &time, &bkpfile, &cfgfile)
		if err != nil {
			fmt.Println("Scan error #650 string")
		}
		fmt.Println("backupid | name | date | time | bkp | cfg")
		fmt.Printf("%3v | %8v | %6v | %3v | %1v | %1v\n", backupid, name, date, time, bkpfile, cfgfile)
		if _date == date {
			if _time == time {
				if bkp == true {
					convertToFile(bkpfile, bkp)
				} else {
					convertToFile(cfgfile, bkp)
				}
			}
		}

	}
	defer db.Close()

}

// convertToFile : Функция, конвертуирющая byteA to *File
func convertToFile(bytes []byte, bkp bool) {
	permissions := os.FileMode(0644)
	if bkp == true {
		bytes := []byte("to be written to a file\n")
		bkperr := ioutil.WriteFile("./Export/sourceBackUp.backup", bytes, permissions)
		if bkperr != nil {
			fmt.Println("Не удалось конвертировать обратно в файл")
		}
	} else {
		bytes := []byte("to be written to a file\n")
		cfgerr := ioutil.WriteFile("./Export/sourceconfig.rsc", bytes, permissions)
		if cfgerr != nil {
			fmt.Println("Не удалось конвертировать обратно в файл")
		}
	}
}
