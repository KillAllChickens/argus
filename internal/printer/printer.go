package printer

import (
	"github.com/KillAllChickens/argus/internal/colors"
	"github.com/KillAllChickens/argus/internal/shared"
	"fmt"
	"math/rand"
	"strings"

	cowsay "github.com/Code-Hex/Neo-cowsay/v2"
)

func Info(format string, a ...any) {
	fmt.Print(colors.FgCyan + "[*] " + colors.Reset)
	fmt.Printf(format+"\n", a...)
}

func Error(format string, a ...any) {
	fmt.Print(colors.FgRed + "[!] " + colors.Reset)
	fmt.Printf(format+"\n", a...)
}

func Warning(format string, a ...any) {
	fmt.Print(colors.FgYellow + "[!] " + colors.Reset)
	fmt.Printf(format+"\n", a...)
}

func Success(format string, a ...any) {
	fmt.Print(colors.FgGreen + "[✔] " + colors.Reset)
	fmt.Printf(format+"\n", a...)
}

func AsciiArtwork() {
	if rand.Intn(10) == 0 { // 1/10 chance of cowsay
		msg := fmt.Sprintf("Argus, made with %s<3%s by the KAC crew!", colors.FgRed, colors.Reset)
		cowText, err := cowsay.Say(
			msg,
			// cowsay.Random(),
			cowsay.Type("eyes"),
		)
		if err != nil {
			fmt.Println("Error generating cowsay:", err)
			return
		}
		fmt.Printf("%s\n\n", cowText)
	} else {
		works := strings.Split(shared.ArtworkFile, "{S}")
		randWork := works[rand.Intn(len(works))]
		workMsg := fmt.Sprintf("Argus, made with %s<3%s by the KAC crew!\n", colors.FgRed, colors.Reset)
		fmt.Printf("%s%s%s\n", colors.FgRed, randWork, colors.Reset)
		fmt.Printf("\t%s\n", workMsg)
	}

}
