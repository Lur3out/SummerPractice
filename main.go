package main

import (
<<<<<<< HEAD
	"crypto/md5"
	"crypto/sha1"
	"database/sql"
	"flag"
=======
>>>>>>> master
	"fmt"
	"io"
	"log"
	"net"
<<<<<<< HEAD
	"os"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"io/ioutil"

	_ "github.com/lib/pq"
)

// Router :: class for Router type
type Router struct {
	num   int
	name  string
	host  string
	login string
	pass  string
	ip    string
	port  int
}

// BackUp :: class for BackUp type
type BackUp struct {
	backupHash string
	configHash string
}

// Arr :: collection for added routers
var Arr [100]Router
var index = 0 //Global counter

// dbconnect :: Параметры подключения к БД
const dbconnect = "host=localhost port=5432 user=postgres password=N0vember1 dbname=backup sslmode=disable"

// Print : Выводит поля экземпляра структуры
func rPrint(r Router) {
	fmt.Println("Num: ", r.num, " name: ", r.name, " hostname: ", r.host, " login: ", r.login, " pass: ", r.pass, " ip: ", r.ip)
}

func main() {
	helpArg := flag.Bool("help", false, "a boolean")      // Отображает доступные команды
	startArg := flag.Bool("start", false, "a boolean")    // Запускает программу в режим службы
	stopArg := flag.Bool("stop", false, "a boolean")      // Оставанливает программу
	timeArg := flag.String("time", "00.00", "a string")   // Задает новое время снятия BackUp
	ipArg := flag.String("ip", "0.0.0.0", "a string")     // Задает Ip роутера
	changeArg := flag.Bool("chng", false, "a boolean")    // Включает функцию изменения времени
	newArg := flag.Bool("new", false, "a boolean")        // Создает новое подключение к роутеру
	loginArg := flag.String("login", "admin", "a string") // Задает логин для подключения к роутеру
	passArg := flag.String("pass", "", "a string")        // Задает пароль для подключения к роутеру
	nameArg := flag.String("name", "Unknown", "a string") // Задает псеводним роутера
	hostArg := flag.String("host", "Default", "a string") // Задает хостнейм роутера
	portArg := flag.Int("port", 22, "an int")             // Задает порт SSH соединения
	bkpArg := flag.Bool("bkp", false, "a boolean")        // Принудительно запускает процедуру снятия BackUp`ов

	flag.Parse()

	if *helpArg == true {
		helpPrint()
	}

	if *startArg == true {
		startProgram()
	}

	if *stopArg == true {
		stopProgram()
	}

	if *changeArg != false {
		timeSetup(*timeArg, *ipArg)
	}

	if *newArg != false {
		var r Router
		newConnection(r, *nameArg, *hostArg, *ipArg, *loginArg, *portArg, *passArg)
		importFile(*loginArg, *passArg, *ipArg, *portArg)
		fmt.Println("\nMD5_Backup: ", hashMD5Bkp())
		fmt.Println("SHA1_Backup: ", hashSHA1Bkp())
		fmt.Println("\n\nMD5_Config: ", hashMD5Cfg())
		fmt.Println("SHA1_Config: ", hashSHA1Cfg())
		fmt.Println("\n\nStarting work w/ PGsql")
		fmt.Println()
		sqlDB(hashMD5Bkp(), hashMD5Cfg(), hashSHA1Bkp(), hashSHA1Cfg())
	}

	if *bkpArg != false {
		//importFile(*loginArg, *passArg, *ipArg, *portArg)
	}

}

// importFile : Функция, создающая ssh-клиент и sftp-соединение и передающая BackUp конфигурации
func importFile(loginArg string, passArg string, ipArg string, portArg int) {

	// Создадим ssh и sftp для передачи файла
	sftpConnection(sshClient(loginArg, passArg, ipArg, portArg))
}

// sshClient : Функция, создающая ssh клиент
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

// sftpConnection : Функция, создающая sftp соединение и импортирующая BackUp файл
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

// sqlDB : Функция, создающая БД в PostgreSQL НЕ РАБОТАЕТ!
func sqlDB(pmd5bkp string, psha1bkp string, pmd5cfg string, psha1cfg string) {

	// Создание БД
	db, err := sql.Open("postgres", "user=backuper password=backup dbname=backup sslmode=disable") //try user:localhost
	if err != nil {
		log.Fatal(err)
	}

	bkp, err := os.Open("C:/Go/Projects/Test/BackUp/BackUp.backup")
	if err != nil {
		fmt.Println("Can't open BackUp.backup")
	}
	defer bkp.Close()
	bkpInfo, err := bkp.Stat()
	if err != nil {
		fmt.Println("Error bkp.Stat()")
	}
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

// hashMD5Bkp : Функция, создающая hash MD5 BackUp.backup
func hashMD5Bkp() string {
	path := "C:/Go/Projects/Test/BackUp/"
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
func hashMD5Cfg() string {
	path := "C:/Go/Projects/Test/BackUp/"
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
func hashSHA1Bkp() string {

	path := "C:/Go/Projects/Test/BackUp/"
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
func hashSHA1Cfg() string {

	path := "C:/Go/Projects/Test/BackUp/"
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

// helpPrint : Функция-принтер
func helpPrint() {
	fmt.Print("----------------------------------------------------------------------------------------------")
	fmt.Println("\nAllowed commands:")
	fmt.Println("\nstart:\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t//Запускает программу в режиме сервиса")
	fmt.Println("\nstop:\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t//Останавливает работу программы")
	fmt.Println("\ntime: -time <hh.mm> [-ip <Router's IP>]\t\t\t\t\t\t\t\t\t\t\t//Установить время создания BackUp-файла")
	fmt.Println("\nnew: -name <Router's name> -host <Hostname> -login <Username> -pass <Password> -ip <Router's IP> -port <Port>\t\t//Создать новое подключение")
	fmt.Println("\nbkp: [-login <Username>]  [-ip <Router's IP>]\t\t\t\t\t\t\t\t\t\t//Сделать принудительный BackUp")
	fmt.Println("\n----------------------------------------------------------------------------------------------")
}

// startProgram : Функция, которая запускает программу в режиме службы
func startProgram() {
	fmt.Println("Программа успешно запущена!")
}

// stopProgram : Функция, останавливающая работу программы.
func stopProgram() {
	fmt.Println("Программа успешно остановлена!")
}

// timeSetup : Функция, задающая таймер создания BackUp-файла
func timeSetup(time string, ip string) {
	fmt.Println("----------------------------------------------------------------------------------------------")
	fmt.Println("Время успешно изменено!")
	fmt.Print("Новое время: ", time)
	if ip != "" {
		fmt.Print(" для ", ip)
	}
	fmt.Println("\n----------------------------------------------------------------------------------------------")
}

// newConnection : Функция, создающая новое подключение к роутеру
func newConnection(r Router, name string, hostname string, ip string, login string, port int, pass string) {
	r.name = name
	r.host = hostname
	r.ip = ip
	r.login = login
	r.port = port
	r.pass = pass
	fmt.Println("----------------------------------------------------------------------------------------------")
	fmt.Println("\nРоутер\t", name, "\t", hostname, "\t", ip, "\t", login, "\t", pass, "\t", port, "\t\tбыл добавлен")
	fmt.Println("\n----------------------------------------------------------------------------------------------")
	Arr[index] = r
	index++
}

// routerPrint : Функция-принтер для Router
func routerPrint(r Router, i int) {
	fmt.Println("#", i, "\tName: ", r.name, "\tHost: ", r.host, "\tLogin: ", r.login, "\tPassword: ", r.pass, "\tIp: ", r.ip, "\tPort: ", r.port)
}

// printArr : Функция-принтер для массива
func printArr() {
	for i := 0; i < index || i < len(Arr); i++ {
		routerPrint(Arr[i], i)
	}
}
=======
)

func main() {
	
}

>>>>>>> master
