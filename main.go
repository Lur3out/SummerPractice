<<<<<<< HEAD
//  ►Сделать три таблицы: роутеры+конфиги, хэши бэкапов+хэши конфигов, бэкапы+конфиги
//
//  ►Сделать флаг -bkp :: bool для снятия только бэкапа(если отсутствует - то снимается конфиг)
//  ►Сделать массив входящих параметров=имена роутеров, с которых снять бэкапы, имена сверяются с именами из БД
//  ►Сделать флаг -all снятия бэкапа со всех роутеров из БД
//
//  ►Подключение к базе вынести в Json
//  ►Временную папку для файлов туда же
//  ►Информацию для подключения к роутерам в базу.
//  ►Процесс снятия конфигурации отделить от процесса снятия бекапа
//  ►В таблицах. Хранить бекапы и конфигурации в одной таблице, хеши в другой!
//  ►Добавить в таблицу хешей время создания и имя роутера.
//  ►Так же добавь в таблицу поле-флаг. Конфигурация или бекап.
//	Сделать экспорт бэкапа из БД по имени и дате
//	►Сделать флаг просмотра подключенных роутеров
//
//

=======
>>>>>>> master
package main

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"io/ioutil"

	_ "github.com/lib/pq"
)

// Router :: class for Router type -- Изменить!
type Router struct {
	num  int
	name string
	//host  string
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
	nameArg := flag.String("name", "Unknown", "a string") // Задает псеводним роутера
	hostArg := flag.String("host", "Default", "a string") // Задает хостнейм роутера
	portArg := flag.Int("port", 22, "an int")             // Задает порт SSH соединения
	bkpArg := flag.Bool("bkp", false, "a boolean")        // Включает снятие бэкапов
	makeArg := flag.Bool("make", false, "a boolean")      // Снятие бэкапа
	pathArg := flag.String("path", "./", "a string")      // Указывает путь на data.json
	lsArg := flag.Bool("ls", false, "a boolean")          // Выводит список подключенных роутеров
	flag.Args()                                           // Имена роутеров, работает только после флага.

	flag.Parse()

	//[0] - sql.BD address; [1] - path to BAckUps
	params := getData(*pathArg)
	//Массив входщящих параметров(имен роутреов)
	names := flag.Args()

	if *helpArg == true {
		helpPrint()
	}

	// Изменить!
	if *newArg != false {
		var r Router
		newConnection(r, *nameArg, *hostArg, *ipArg, *loginArg, *portArg, *passArg, params)
		fmt.Println()
	}

	if *makeArg != false {
		if *bkpArg != false {
			if *allArg != false {
				makeAllBackUp(params, *ipArg, *portArg, *loginArg, *passArg, *bkpArg, names)
			}
			makeBackUp(params, *ipArg, *portArg, *loginArg, *passArg, *bkpArg)
		}
		if *allArg != false && *bkpArg == false {
			makeAllConfig(params, *ipArg, *portArg, *loginArg, *passArg, *bkpArg, names)
		}
		if *bkpArg == false && *allArg == false {
			makeConfig(params, *ipArg, *portArg, *loginArg, *passArg, *bkpArg)
		}
	}

	if *lsArg != false {
		//Дата + время
		current := time.Now()
		_date := current.Format("02.01.2006")
		_time := current.Format("15:04")
		fmt.Print(_date)
		fmt.Println(" ", _time)
		fmt.Println()
		printAllConnected(connectDB(params[0]))

	}
}

//*************************************************************
//---------------------- -HELP---------------------------------
// helpPrint : Функция-принтер
func helpPrint() {
	fmt.Print("----------------------------------------------------------------------------------------------")
	fmt.Println("\nAllowed commands:")
	fmt.Println("\npath:  <Path to data.json>\t\t\t\t\t\t\t\t\t\t\t\t\t// Укажите путь к файлу data.json в формате <Disk>:/~/")
	fmt.Println("\nnew: -name <Router's name> -host <Hostname> -login <Username> -pass <Password> -ip <Router's IP> -port <Port>\t\t//Создать новое подключение")
	fmt.Println("\nmake: [-bkp ] [-all] <Routers' names>\t\t\t\t\t\t\t\t\t\t\t\t//Запуск процедуры снятия бэкапов.\n\n\nАтрибуты:\n\n-bkp\t\t\tПри указании атрибута произведется снятие полного бэкапа. По умолчанию снимается конфигурационный файл.\n\n-all\t\t\tПри указании снимаются бэкапы/конфиги со всех роутеров, подключенных к прогамме\n\n\nВходящие параметры:\n\n<Routers' names>\tЧерез пробел перечилите названия роутеров, с которых произвести снятие бэкапа. При наличии атрибута -all снятие будет произведено со всех.")
	fmt.Println("\n----------------------------------------------------------------------------------------------")
}

//*************************************************************
//-----------------------------SSH-----------------------------

// sshClient : Функция, создающая ssh клиент -- Изменить!
func sshClient(loginArg string, passArg string, ipArg string, portArg int) *ssh.Client {
	config := &ssh.ClientConfig{
		User: loginArg,
		Auth: []ssh.AuthMethod{
			ssh.Password(passArg),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	addr := fmt.Sprintf("%s:%d", ipArg, portArg)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		fmt.Printf("Failed to dial: %s", err)
	}
	fmt.Println("Successfully connected to ", ipArg, ":", portArg)

	session, err := client.NewSession()
	if err != nil {
		fmt.Printf("Failed to create a new session: %s", err)
	}
	defer session.Close()

	session2, err := client.NewSession()
	if err != nil {
		fmt.Printf("Failed to create a new session: %s", err)
	}
	defer session2.Close()

	b, err := session.CombinedOutput("/system backup save name=BackUp dont-encrypt=yes") // /system backup save name=BackUp dont-encrypt=yes
	if err != nil {
		fmt.Printf("Failed to send output command: %s", err)
	}
	fmt.Print(string(b))

	c, err := session2.CombinedOutput("/export file=config.rsc") // /export file=config.rsc
	if err != nil {
		fmt.Printf("Failed to send output command: %s", err)
	}
	fmt.Print(string(c))

	return client
}

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

// sftpConnection : Функция, создающая sftp соединение и импортирующая BackUp файл -- Изменить!
func sftpConnection(client *ssh.Client) {
	sftp, err := sftp.NewClient(client)
	if err != nil {
		fmt.Printf("Failed to create new sftp-client: %s", err)
	}
	defer sftp.Close()

	srcPath := "/"
	dstPath := "C:/Go/Projects/Test/BackUp/"
	filename := "BackUp.backup"
	config := "config.rsc"

	// Open the source file
	srcFile, err := sftp.Open(srcPath + filename)
	if err != nil {
		fmt.Printf("Failed to open backup file on router: %s", err)
	}
	defer srcFile.Close()

	// Open the source file
	srcFile2, err := sftp.Open(srcPath + config)
	if err != nil {
		fmt.Printf("Failed to open backup file on router: %s", err)
	}
	defer srcFile2.Close()

	// Create the destination file
	dstFile, err := os.Create(dstPath + filename)
	if err != nil {
		fmt.Printf("Failed to create destination file: %s", err)
	}
	defer dstFile.Close()

	dstFile2, err := os.Create(dstPath + config)
	if err != nil {
		fmt.Printf("Failed to create destination file: %s", err)
	}
	defer dstFile2.Close()

	// Copy the file
	srcFile.WriteTo(dstFile)
	srcFile2.WriteTo(dstFile2)
}

// sftpRouter : Функция, создающая sftp соединение и экспортирующая backup\rsc.
func sftpRouter(client *ssh.Client, bkp bool, path string) {
	sftp, err := sftp.NewClient(client)
	if err != nil {
		fmt.Printf("Failed to create new sftp-client: %s", err)
	}
	defer sftp.Close()

	srcPath := "/"
	dstPath := path
	filename := "BackUp.backup"
	config := "config.rsc"

	if bkp == true {

		// Open the source file
		srcFile, err := sftp.Open(srcPath + filename)
		if err != nil {
			fmt.Printf("Failed to open backup file on router: %s", err)
		}
		defer srcFile.Close()

		// Create the destination file
		dstFile, err := os.Create(dstPath + filename)
		if err != nil {
			fmt.Printf("Failed to create destination file: %s", err)
		}
		defer dstFile.Close()

		// Copy the file
		srcFile.WriteTo(dstFile)
	} else {

		// Open the source file
		srcFile2, err := sftp.Open(srcPath + config)
		if err != nil {
			fmt.Printf("Failed to open config file on router: %s", err)
		}
		defer srcFile2.Close()

		dstFile2, err := os.Create(dstPath + config)
		if err != nil {
			fmt.Printf("Failed to create destination file: %s", err)
		}
		defer dstFile2.Close()
		// Copy the file
		srcFile2.WriteTo(dstFile2)
	}

}

//*************************************************************
//-------------------------SQL---------------------------------
// sqlDB : Функция, создающая БД в PostgreSQL -- Изменить!
func sqlDB(pmd5bkp string, psha1bkp string, pmd5cfg string, psha1cfg string) {

	// Подключение к БД
	db, err := sql.Open("postgres", "user=backuper password=backup dbname=backup sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	bkp, err := os.Open("C:/Go/Projects/Test/BackUp/BackUp.backup")
	if err != nil {
		fmt.Println("Can't open BackUp.backup")
	}
	defer bkp.Close()
	bkpInfo, _ := bkp.Stat()
	bkpsize := bkpInfo.Size()
	bkpbytes := make([]byte, bkpsize)

	cfg, err := os.Open("C:/Go/Projects/Test/BackUp/config.rsc")
	if err != nil {
		fmt.Println("Can't open config.rsc")
	}
	defer cfg.Close()
	cfgInfo, _ := cfg.Stat()
	cfgsize := cfgInfo.Size()
	cfgbytes := make([]byte, cfgsize)

	var lastInsertID int
	err = db.QueryRow("INSERT INTO backupinfo(md5_bkp, sha1_bkp, md5_cfg, sha1_cfg, bkp_file, cfg_file) VALUES($1,$2,$3,$4,$5,$6) returning backup_id;", "md5_BACKUP", "sha1_BACKUP", "md5_CONFIG", "sha1_CONFIG", "bkp_ByteA", "cfg_ByteA").Scan(&lastInsertID)
	if err != nil {
		fmt.Println("QueryRow err # 211 string")
	}
	fmt.Println("last inserted id =", lastInsertID)

	fmt.Println("# Updating")
	stmt, err := db.Prepare("update backupinfo set md5_bkp=$1, sha1_bkp=$2, md5_cfg=$3, sha1_cfg=$4, bkp_file=$5, cfg_file=$6 where backup_id=$7")
	if err != nil {
		fmt.Println("Prepare error #218 string")
	}
	res, err := stmt.Exec(pmd5bkp, psha1bkp, pmd5cfg, psha1cfg, bkpbytes, cfgbytes, lastInsertID)
	if err != nil {
		fmt.Println("Exec error #222 string")
	}

	affect, err := res.RowsAffected()
	if err != nil {
		fmt.Println("RowsAffected error #227 string")
	}

	fmt.Println(affect, "rows changed")

	fmt.Println("# Querying")

	rows, err := db.Query("SELECT * FROM backupinfo")
	if err != nil {
		fmt.Println("Query error #237 string")
	}

	for rows.Next() {
		var backupid int
		md5bkp := pmd5bkp
		sha1bkp := psha1bkp
		md5cfg := pmd5cfg
		sha1cfg := psha1cfg
		bkpfile := bkpbytes
		cfgfile := cfgbytes
		err = rows.Scan(&backupid, &md5bkp, &sha1bkp, &md5cfg, &sha1cfg, &bkpfile, &cfgfile)
		if err != nil {
			panic(err) //fmt.Println("Scan error #249 string")
		}
		fmt.Println("backup_id | md5_bkp | sha1_bkp | md5_cfg | sha1_cfg | bkp_file | cfg_file")
		fmt.Printf("%3v | %8v | %6v | %8v | %6v | %2v | %1v\n", backupid, md5bkp, sha1bkp, md5cfg, sha1cfg, bkpfile, cfgfile)
		//**********Back to FILE from ByteArray
		permissions := os.FileMode(0644)
		bkpbyteArray := []byte("to be written to a file\n")
		bkperr := ioutil.WriteFile("C:/Go/Projects/Test/BackUp/reversedBackUp.backup", bkpbyteArray, permissions)
		if bkperr != nil {
			fmt.Println("Не удалось конвертировать обратно файл")
		}

		cfgbyteArray := []byte("to be written to a file\n")
		cfgerr := ioutil.WriteFile("C:/Go/Projects/Test/BackUp/reversed_config.rsc", cfgbyteArray, permissions)
		if cfgerr != nil {
			fmt.Println("Не удалось конвертировать обратно в файл")
		}

		//*************************************
	}

	//DELETING
	fmt.Println("# Deleting")
	stmt, err = db.Prepare("delete from backupinfo where backup_id=$1")
	if err != nil {
		fmt.Println("Prepare error #260 string")
	}

	res, err = stmt.Exec(lastInsertID)
	if err != nil {
		fmt.Println("Exec error #265 string")
	}

	affect, err = res.RowsAffected()
	if err != nil {
		fmt.Println("RowsAffected error #270 string")
	}

	fmt.Println(affect, "rows changed")

}

// connectDB : Функция, создающая подключение к PostgreSql по настройкам из файла data.json
func connectDB(settings string) *sql.DB {

	// Подключение к БД
	db, err := sql.Open("postgres", settings)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// addNewRouter : Функция, добавляющая данные нового подключения в PostgreSql
func addNewRouter(r Router, settings string) {
	// db - sql.DB
	//db := connectDB(settings)
	db, err := sql.Open("postgres", settings)

	var lastInsertID int
	err = db.QueryRow("INSERT INTO test(name, ip, port, login, pass) VALUES($1,$2,$3,$4,$5) returning test_id;", r.name, r.ip, r.port, r.login, r.pass).Scan(&lastInsertID)
	if err != nil {
		fmt.Println("QueryRow err #370 string")
	}
	fmt.Println("last inserted id =", lastInsertID)

	defer db.Close()
}

// addNewHash : Функция, добавляющая хэши нового бэкапа
func addNewHash(md5 string, sha1 string, name string, set string, bkp bool) {

	current := time.Now()
	time := current.Format("15:04")
	date := current.Format("02.01.2006")

	db, err := sql.Open("postgres", set)

	if bkp == true {
		var lastInsertID int
		err = db.QueryRow("INSERT INTO hashs_test(date, name, md5bkp, sha1bkp, md5cfg, sha1cfg, time) VALUES($1,$2,$3,$4,$5,$6,$7) returning hash_id;", date, name, md5, sha1, "", "", time).Scan(&lastInsertID)
		if err != nil {
			panic(err) // fmt.Println("QueryRow err #471 string")
		}
		// fmt.Println("last inserted id =", lastInsertID)
	} else {
		var lastInsertID int
		err = db.QueryRow("INSERT INTO hashs_test(date, name, md5bkp, sha1bkp, md5cfg, sha1cfg, time) VALUES($1,$2,$3,$4,$5,$6,$7) returning hash_id;", date, name, "", "", md5, sha1, time).Scan(&lastInsertID)
		if err != nil {
			fmt.Println("QueryRow err #478 string")
		}
		// fmt.Println("last inserted id =", lastInsertID)
	}
	defer db.Close()

}

// addNewFile : Функция, добавляющая файл нового бэкапа
func addNewFile(file []byte, set string, bkp bool) {
	current := time.Now()
	time := current.Format("15:04")
	date := current.Format("02.01.2006")

	db, err := sql.Open("postgres", set)

	if bkp == true {
		var lastInsertID int
		err = db.QueryRow("INSERT INTO backups_test(date, time, bkp, cfg) VALUES($1,$2,$3,$4) returning backup_id;", date, time, file, "").Scan(&lastInsertID)
		if err != nil {
			panic(err) // fmt.Println("QueryRow err #498 string")
		}
		// fmt.Println("last inserted id =", lastInsertID)
	} else {
		var lastInsertID int
		err = db.QueryRow("INSERT INTO backups_test(date, time, bkp, cfg) VALUES($1,$2,$3,$4) returning backup_id;", date, time, "", file).Scan(&lastInsertID)
		if err != nil {
			fmt.Println("QueryRow err #505 string")
		}
		// fmt.Println("last inserted id =", lastInsertID)
	}
	defer db.Close()
}

// test : TEST
func test(r Router, settings string) {
	// Подключение к БД
	db, err := sql.Open("postgres", settings)
	if err != nil {
		log.Fatal(err)
	}

	var lastInsertID int
	err = db.QueryRow("INSERT INTO test(name, ip, port, login, pass) VALUES($1,$2,$3,$4,$5) returning test_id;", r.name, r.ip, r.port, r.login, r.pass).Scan(&lastInsertID)
	if err != nil {
		panic(err) //fmt.Println("QueryRow err # 427 string")
	}
	fmt.Println("last inserted id =", lastInsertID)

	fmt.Println("# Updating")
	stmt, err := db.Prepare("update test set name=$1, ip=$2, port=$3, login=$4, pass=$5 where test_id=$6")
	if err != nil {
		fmt.Println("Prepare error #434 string")
	}
	res, err := stmt.Exec(r.name, r.ip, r.port, r.login, r.pass, lastInsertID)
	if err != nil {
		fmt.Println("Exec error #438 string")
	}

	affect, err := res.RowsAffected()
	if err != nil {
		fmt.Println("RowsAffected error #443 string")
	}

	fmt.Println(affect, "rows changed")

	fmt.Println("# Querying")

	rows, err := db.Query("SELECT * FROM test")
	if err != nil {
		fmt.Println("Query error #452 string")
	}

	for rows.Next() {
		var testid int
		name := r.name
		ip := r.ip
		port := r.port
		login := r.login
		pass := r.pass
		err = rows.Scan(&testid, &name, &ip, &port, &login, &pass)
		if err != nil {
			panic(err) //fmt.Println("Scan error #464 string")
		}
		fmt.Println("test_id | name | ip | port | login | pass")
		fmt.Printf("%3v | %8v | %6v | %8v | %6v | %6v\n", testid, name, ip, port, login, pass)

		//*************************************
	}

	//DELETING
	fmt.Println("# Deleting")
	stmt, err = db.Prepare("delete from test where test_id=$1")
	if err != nil {
		fmt.Println("Prepare error #476 string")
	}

	res, err = stmt.Exec(lastInsertID)
	if err != nil {
		fmt.Println("Exec error #481 string")
	}

	affect, err = res.RowsAffected()
	if err != nil {
		fmt.Println("RowsAffected error #486 string")
	}

	fmt.Println(affect, "rows changed")
}

// printAllConnected : Функция, выводящая все подключенные роутеры
func printAllConnected(db *sql.DB) {
	rows, err := db.Query("SELECT * FROM test")
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
		err = rows.Scan(&testid, &name, &ip, &port, &login, &pass)
		if err != nil {
			panic(err) //fmt.Println("Scan error #464 string")
		}
		fmt.Println("test_id | name | ip | port | login | pass")
		fmt.Printf("%3v | %8v | %6v | %8v | %6v | %6v\n", testid, name, ip, port, login, pass)

		//*************************************
	}
	defer db.Close()
}

// deleteRow : Delete row
func deleteRow(index int, settings string) {
	fmt.Println("# Deleting")

	db := connectDB(settings)
	stmt, _ := db.Prepare("delete from test where test_id=$1")

	res, _ := stmt.Exec(index)

	affect, _ := res.RowsAffected()

	fmt.Println(affect, "rows changed")
}

// routerData : Функция, создающая .backup\config всех подключенных роутеров.
func routerData(db *sql.DB, bkp bool, params [2]string) {
	var r Router
	rows, err := db.Query("SELECT * FROM test")
	if err != nil {
		fmt.Println("Query error #452 string")
	}

	for rows.Next() {
		var testid int
		path := params[1]
		set := params[0]

		name := ""
		ip := ""
		port := 0
		login := ""
		pass := ""
		err = rows.Scan(&testid, &name, &ip, &port, &login, &pass)
		if err != nil {
			fmt.Println("Scan error #464 string")
		}
		fmt.Println("test_id | name | ip | port | login | pass")
		fmt.Printf("%3v | %8v | %6v | %8v | %6v | %6v\n", testid, name, ip, port, login, pass)

		r.num = testid
		r.name = name
		r.ip = ip
		r.port = port
		r.login = login
		r.pass = pass

		//Получение backup/rsc с роутера по полученным данным
		sftpRouter(sshRouter(login, pass, ip, port), bkp, path)
		//Занесение в БД хэшей и backup/config
		sqlRouter(bkp, path, r, set)

	}
	defer db.Close()
}

// sqlRouter : Функция добавляющая данные в БД(хэши и файлы)
func sqlRouter(bkp bool, path string, r Router, set string) {
	if bkp == true {
		//Добавление .backup
		file, err := os.Open(path + "BackUp.backup")
		if err != nil {
			fmt.Println("Не удалось открыть файл # 605 string")
		}
		defer file.Close()

		md5 := hashMD5Bkp(path)
		sha1 := hashSHA1Bkp(path)

		fileInfo, _ := file.Stat()
		fileSize := fileInfo.Size()
		//file :: bytea
		filebytes := make([]byte, fileSize)

		//Sql: hash
		addNewHash(md5, sha1, r.name, set, bkp)

		//Sql: file
		addNewFile(filebytes, set, bkp)

	} else {
		//Добавление .rsc

		file, err := os.Open(path + "config.rsc")
		if err != nil {
			fmt.Println("Не удалось открыть файл # 614 string")
		}
		defer file.Close()

		md5 := hashMD5Cfg(path)
		sha1 := hashSHA1Cfg(path)

		fileInfo, _ := file.Stat()
		fileSize := fileInfo.Size()
		//file :: bytea
		filebytes := make([]byte, fileSize)

		//Sql: hash
		addNewHash(md5, sha1, r.name, set, bkp)

		//Sql: file
		addNewFile(filebytes, set, bkp)
	}
}

//**************************************************************
//----------------------MD5 u SHA-1-----------------------------

// hashMD5Bkp : Функция, создающая hash MD5 BackUp.backup
func hashMD5Bkp(_path string) string {
	path := _path
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

	//fmt.Printf("BackUp_MD5:%x", backupHash.Sum(nil))
	return fmt.Sprintf("%x", backupHash.Sum(nil))
}

// hashMD5Cfg : Функция, создающая hash MD5 config.rsc
func hashMD5Cfg(_path string) string {
	path := _path
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

	//fmt.Printf("Config_MD5:%x", configHash.Sum(nil))
	return fmt.Sprintf("%x", configHash.Sum(nil))
}

// hashSHA1Bkp : Функция, создающая hash SHA1 BackUp.backup
func hashSHA1Bkp(_path string) string {

	path := _path
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

	//fmt.Printf("BackUp_SHA1:% x", backupHash.Sum(nil))
	return fmt.Sprintf("%x", backupHash.Sum(nil))
}

// hashSHA1Cfg : Функция, создающая hash SHA1 config.rsc
func hashSHA1Cfg(_path string) string {

	path := _path
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

	//fmt.Printf("Config_SHA1:% x", configHash.Sum(nil))
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
	//fmt.Println(str[0])
	//fmt.Println(str[1])
	return str
}

//**************************************************************
//-------------------NEW CONNECTION(ROUTER)---------------------
// newConnection : Функция, создающая новое подключение к роутеру -- Изменить! Добавить!
func newConnection(r Router, name string, hostname string, ip string, login string, port int, pass string, params [2]string) {
	r.name = name
	//r.host = hostname
	r.ip = ip
	r.login = login
	r.port = port
	r.pass = pass
	fmt.Println("----------------------------------------------------------------------------------------------")
	fmt.Println("\nРоутер\t", r.name, "\t", r.ip, "\t", r.login, "\t", r.pass, "\t", r.port, "\t\tбыл добавлен")
	fmt.Println("\n----------------------------------------------------------------------------------------------")
	//SQL-Adding
	//test(r, params[0])

	addNewRouter(r, params[0])
	// printAllConnected(connectDB(params[0]))
	// deleteRow(12, params[0])

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

	//dbconfig := params[0]
	//savepath := params[1]

	routerData(connectDB(params[0]), bkp, params)
}

// makeBackUp : -make -bkp <names>
func makeBackUp(params [2]string, ip string, port int, login string, pass string, bkp bool) {
	//Снятие бэкапов с перечисленных роутеров

	//dbconfig := params[0]
	//savepath := params[1]

	//sftpRouter(sshRouter(login, pass, ip, port), bkp, params)
}

// makeAllConfig : -make -all
func makeAllConfig(params [2]string, ip string, port int, login string, pass string, bkp bool, names []string) {
	//Снятие конфигов со всех роутеров

	//dbconfig := params[0]
	//savepath := params[1]

	routerData(connectDB(params[0]), bkp, params)
}

// makeConfig : -make <names>
func makeConfig(params [2]string, ip string, port int, login string, pass string, bkp bool) {
	//Снятие конфигов с перечисленных роутеров

	//dbconfig := params[0]
	//savepath := params[1]

	//sftpRouter(sshRouter(login, pass, ip, port), bkp, params)
}

//***************************************************************
