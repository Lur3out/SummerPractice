package main

import (
	"flag"
	"fmt"
)

func main() {
	/*ipPtr := flag.String("ip", "10.0.0.1", "a string")
	logPtr := flag.String("log", "admin", "a string")

	numbPtr := flag.Int("numb", 42, "an int")
	newPtr := flag.Bool("new", true, "a bool")

	var svar string
	flag.StringVar(&svar, "svar", "bar", "a string var")

	flag.Parse()

	fmt.Println("ip:", *ipPtr)
	fmt.Println("log:", *logPtr)
	fmt.Println("numb:", *numbPtr)
	fmt.Println("new:", *newPtr)
	fmt.Println("svar:", svar)
	fmt.Println("tail:", flag.Args())

	//defer recover()*/
	helpArg := flag.Bool("help", false, "a boolean")   //Отображает доступные команды
	startArg := flag.Bool("start", false, "a boolean") // Запускает программу в режим службы
	stopArg := flag.Bool("stop", false, "a boolean")   // Оставанливает программу

	timeArg := flag.String("time", "00.00", "a string") //Задает новое время снятия BackUp
	ipArg := flag.String("ip", "", "a string")          // Задает Ip роутера
	changeArg := flag.Bool("chng", false, "a boolean")  // Включает функцию изменения времени

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
}

// helpPrint : Функция-принтер
func helpPrint() {
	fmt.Print("----------------------------------------------------------------------------------------------")
	fmt.Println("\nAllowed commands:")
	fmt.Println("\nstart:\t\t\t\t\t\t//Запускает программу в режиме сервиса")
	fmt.Println("\nstop:\t\t\t\t\t\t//Останавливает работу программы")
	fmt.Println("\ntime: -time <hh.mm> [<Router's IP>]\t\t//Установить время создания BackUp-файла")
	fmt.Println("\nnew: -login=<Username>  <Router's IP>\t\t//Создать новое подключение")
	fmt.Println("\nbkp: [-login=<Username>]  [<Router's IP>]\t//Сделать принудительный BackUp")
	fmt.Println("\nlist: \t\t\t\t\t\t//Список обслуживаемых роутеров")
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
