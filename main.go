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
	helpArg := flag.Bool("help", false, "a boolean")

	flag.Parse()

	if *helpArg == true {
		fmt.Print("----------------------------------------------------------------------------------------------")
		fmt.Println("\nAllowed commands:")
		fmt.Println("\nstart:\t\t\t\t\t\t//Запускает программу в режиме сервиса")
		fmt.Println("\nstop:\t\t\t\t\t\t//Останавливает работу программы")
		fmt.Println("\ntime: -time=<hh.mm>\t\t\t\t//Установить время создания BackUp-файла")
		fmt.Println("\nnew: -login=<Username>  <Router's IP>\t\t//Создать новое подключение")
		fmt.Println("\nbkp: [-login=<Username>]  [<Router's IP>]\t//Сделать принудительный BackUp")
		fmt.Println("\n----------------------------------------------------------------------------------------------")

	}
}
