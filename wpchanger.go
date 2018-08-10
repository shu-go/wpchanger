package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"unsafe"

	"image"
	"image/jpeg"
	_ "image/png"

	"bitbucket.org/shu_go/gli"
	"bitbucket.org/shu_go/rog"
	"github.com/andrew-d/go-termutil"
)

var (
	dllUser32                = syscall.NewLazyDLL("user32.dll")
	procSystemParametersInfo = dllUser32.NewProc("SystemParametersInfoW")
)

const (
	// SystemParametersInfo
	SPI_SETDESKWALLPAPER  = 0x0014
	SPI_GETDESKWALLPAPER  = 0x0073
	SPIF_UPDATEINIFILE    = 0x0001
	SPIF_SENDWININICHANGE = 0x0002
)

type globalCmd struct {
	File  string `cli:"f, file=FILE_NAME" help:"the name of an image file. (defaults to stdin/stdout)"`
	Get   getCmd
	Debug bool
}

type getCmd struct{}

func (cmd globalCmd) Run() error {
	if cmd.Debug {
		rog.EnableDebug()
	}

	input := cmd.File

	// save stdin raw data to a file
	if input == "" {
		if !termutil.Isatty(os.Stdin.Fd()) {
			input = filepath.Join(homeDirPath(), "_wallpaper_by_wpchanger_.jpg")

			img, _, err := image.Decode(os.Stdin)
			if err != nil {
				return fmt.Errorf("decode stdin: %v", err)
			}

			f, err := os.Create(input)
			if err != nil {
				return fmt.Errorf("create input: %v", err)
			}
			// defer f.Close()  no defer!!

			b := bufio.NewWriter(f)
			err = jpeg.Encode(b, img, &jpeg.Options{jpeg.DefaultQuality})
			if err != nil {
				return fmt.Errorf("encode: %v", err)
			}
			err = b.Flush()
			if err != nil {
				return fmt.Errorf("flush: %v", err)
			}

			f.Close()
		} else {
			return fmt.Errorf("no file specified")
		}
	}

	rog.Debug("input", input)
	err := SetWallpaper(input)
	if err != nil {
		return fmt.Errorf("set wallpaper: %v", err)
	}
	return nil
}

func (cmd getCmd) Run(global globalCmd) error {
	wallname, err := GetWallpaper()
	if err != nil {
		return fmt.Errorf("get wallpaper: %v", err)
	}

	wallfile, err := os.Open(wallname)
	if err != nil {
		return fmt.Errorf("open wallpaper: %v", err)
	}
	defer wallfile.Close()

	if global.File == "" {
		if !termutil.Isatty(os.Stdout.Fd()) {
			_, err := io.Copy(os.Stdout, wallfile)
			if err != nil {
				return fmt.Errorf("copy: %v", err)
			}
		} else {
			return fmt.Errorf("no file specified")
		}
	} else {
		output, err := os.Create(global.File)
		if err != nil {
			return fmt.Errorf("create: %v", err)
		}
		defer output.Close()

		_, err = io.Copy(output, wallfile)
		if err != nil {
			return fmt.Errorf("copy: %v", err)
		}
	}

	return nil
}

func main() {
	app := gli.NewWith(&globalCmd{})
	app.Version = "0.2.0"
	app.Copyright = "(C) 2018 Shuhei Kubota"
	app.Usage = `wpchanger -f wallpaper.jpg
wpchanger -f wallpaper.jpg get`
	err := app.Run(os.Args)
	if err != nil {
		os.Exit(1)
	}
}

func SetWallpaper(filename string) error {
	_, _, err := procSystemParametersInfo.Call(
		SPI_SETDESKWALLPAPER,
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(filename))),
		SPIF_SENDWININICHANGE|SPIF_UPDATEINIFILE)

	return err
}

func GetWallpaper() (string, error) {
	buf := make([]uint16, 260)

	_, _, _ /*err*/ = procSystemParametersInfo.Call(
		SPI_GETDESKWALLPAPER,
		260,
		uintptr(unsafe.Pointer(&buf[0])),
		0)

	return syscall.UTF16ToString(buf), nil
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
