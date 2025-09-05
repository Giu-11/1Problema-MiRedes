package estilo

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func Clear() {
    var cmd *exec.Cmd
    if runtime.GOOS == "windows" {
        cmd = exec.Command("cmd", "/c", "cls")
    } else {
        cmd = exec.Command("clear")
    }
    cmd.Stdout = os.Stdout
    cmd.Run()
}

func PrintVerm(texto string) {
	fmt.Printf("\033[31m%s\033[0m", texto)
}

func PrintVerd(texto string) {
	fmt.Printf("\033[32m%s\033[0m", texto)
}

func PrintMag(texto string) {
	fmt.Printf("\033[35m%s\033[0m", texto)
}

func PrintCian(texto string) {
	fmt.Printf("\033[36m%s\033[0m", texto)
}

func PrintAma(texto string) {
	fmt.Printf("\033[33m%s\033[0m", texto)
}