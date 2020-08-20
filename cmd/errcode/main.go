package main

import (
	"fmt"
	"math/rand"
	"os/exec"
	"runtime"
	"time"
)

// How to name custom error codes https://stackoverflow.com/a/22600394
// • Start with a capital letter but not F (predefined config file errors), H (fdw), P (PL/pgSQL) or X (internal).
// • Do not use 0 (zero) or P in the 3rd column. Predefined error codes use these commonly.
// • Use a capital letter in the 4th position. No predefined error codes have this.
// 'As an example, start with a character for your app: "T". Then a two-char error class: "3G". Then a sequential code "A0"-"A9", "B0"-"B9", etc. Yields T3GA0, T3GA1, etc.'

func main() {
	rand.Seed(time.Now().UnixNano())
	alphabets := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
	numbers := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}
	alphanums := append(alphabets, numbers...)
	errcode := "O" +
		alphabets[rand.Intn(len(alphabets))] +
		alphanums[rand.Intn(len(alphanums))] +
		alphanums[rand.Intn(len(alphanums))] +
		alphanums[rand.Intn(len(alphanums))]
	fmt.Println(errcode)

	// Attempt to copy the errcode into the user's clipboard. Depends on the
	// user's operating system
	var err error
	switch runtime.GOOS {
	case "windows":
	case "darwin":
		cmd := exec.Command("bash", "-c", fmt.Sprintf("printf '%s' | pbcopy", errcode))
		err = cmd.Run()
	case "linux":
	}
	if err != nil {
		fmt.Println(err.Error())
	}
}
