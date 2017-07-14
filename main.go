package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

//import "github.com/pkg/sftp"

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

//type Arr [100]Router

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
	}
}

// importFile : Функция, создающая ssh-клиент и sftp-соединение и передающая BackUp конфигурации
func importFile(loginArg string, passArg string, ipArg string, portArg int) {

	//Создадим ssh-клиент

	//Создадим sftp-соединение и передадим файл конфигурации
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

	b, err := session.CombinedOutput("/system backup save name=BackUp dont-encrypt=yes") // /system backup save name=BackUp dont-encrypt=yes
	if err != nil {
		fmt.Printf("Failed to send output command: %s", err)
	}
	fmt.Print(string(b))
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
	fmt.Println("\nlist: \t\t\t\t\t\t\t\t\t\t\t\t\t\t\t//Список обслуживаемых роутеров")
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

// newConnection() : Функция, создающая новое подключение к роутеру
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
}
