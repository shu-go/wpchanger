package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime"
	"syscall"
	"unsafe"

	"image"
	"image/jpeg"
	_ "image/png"

	"github.com/andrew-d/go-termutil"
)

var (
	dllUser32                = syscall.NewLazyDLL("user32.dll")
	procSystemParametersInfo = dllUser32.NewProc("SystemParametersInfoW")
)

const (
	// SystemParametersInfo
	SPI_SETDESKWALLPAPER  = 0x0014
	SPIF_UPDATEINIFILE    = 0x0001
	SPIF_SENDWININICHANGE = 0x0002
)

func main() {
	var input string

	if len(os.Args) > 2 {
		log.Println("too many arguments.")
		os.Exit(1)
	}

	if len(os.Args) == 2 {
		input = os.Args[1]
	}

	// save stdin raw data to a file
	if input == "" && !termutil.Isatty(os.Stdin.Fd()) {
		input = fmt.Sprintf("%s/%s", homeDirPath(), "_wallpaper_by_wpchanger_.jpg")

		img, _, err := image.Decode(os.Stdin)
		if err != nil {
			log.Println("failed to load base image from stdin")
			log.Fatal(err)
		}

		f, err := os.Create(input)
		if err != nil {
			log.Fatal(err)
		}
		// defer f.Close()  no defer!!

		b := bufio.NewWriter(f)
		err = jpeg.Encode(b, img, &jpeg.Options{jpeg.DefaultQuality})
		if err != nil {
			log.Fatal(err)
		}
		err = b.Flush()
		if err != nil {
			log.Fatal(err)
		}

		f.Close()
	}

	ret := SetWallpaper(input)

	if ret == 0 {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func SetWallpaper(filename string) uintptr {
	ret, _, _ := procSystemParametersInfo.Call(
		SPI_SETDESKWALLPAPER,
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(filename))),
		SPIF_SENDWININICHANGE|SPIF_UPDATEINIFILE)

	return ret
}

func homeDirPath() string {
	var path string

	if runtime.GOOS == "windows" {
		path = os.Getenv("APPDATA")
	} else {
		path = os.Getenv("HOME")
	}

	return path
}
