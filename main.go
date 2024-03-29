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

	"github.com/andrew-d/go-termutil"
	"github.com/shu-go/gli"
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
	Set setCmd `cli:"set" help:"set wallpaper"`
	Get getCmd `cli:"get" help:"get wallpaper"`
}

type setCmd struct{}

type getCmd struct{}

func (cmd setCmd) Run(args []string) error {
	var input string
	if len(args) > 0 {
		input = args[0]

		var err error
		input, err = filepath.Abs(input)
		if err != nil {
			return err
		}

		s, err := os.Stat(input)
		if err != nil {
			return err
		} else if s.IsDir() {
			return fmt.Errorf("%s is a directory", input)
		}
	}

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
			return fmt.Errorf("no input")
		}
	}

	err := SetWallpaper(input)
	if err != nil {
		return fmt.Errorf("set wallpaper: %v", err)
	}

	return nil
}

func (cmd getCmd) Run(args []string) error {
	var output string
	if len(args) > 0 {
		output = args[0]

		var err error
		output, err = filepath.Abs(output)
		if err != nil {
			return err
		}
	}

	wallname, err := GetWallpaper()
	if err != nil {
		return fmt.Errorf("get wallpaper: %v", err)
	}

	wallfile, err := os.Open(wallname)
	if err != nil {
		return fmt.Errorf("open wallpaper: %v", err)
	}
	defer wallfile.Close()

	if output == "" {
		if !termutil.Isatty(os.Stdout.Fd()) {
			_, err := io.Copy(os.Stdout, wallfile)
			if err != nil {
				return fmt.Errorf("copy: %v", err)
			}
		} else {
			return fmt.Errorf("no output")
		}
	} else {
		f, err := os.Create(output)
		if err != nil {
			return fmt.Errorf("create: %v", err)
		}
		defer f.Close()

		_, err = io.Copy(f, wallfile)
		if err != nil {
			return fmt.Errorf("copy: %v", err)
		}
	}

	return nil
}

func (cmd globalCmd) Run(args []string, app *gli.App) error {
	if len(args) == 0 {
		app.Help(os.Stderr)
		return nil
	}

	return app.Run(append([]string{"set"}, args...))
}

func SetWallpaper(filename string) error {
	r1, _, err := procSystemParametersInfo.Call(
		SPI_SETDESKWALLPAPER,
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(filename))),
		SPIF_SENDWININICHANGE|SPIF_UPDATEINIFILE)

	if r1 == 1 {
		return nil
	}
	return err
}

func GetWallpaper() (string, error) {
	buf := make([]uint16, 260)

	r1, _, err := procSystemParametersInfo.Call(
		SPI_GETDESKWALLPAPER,
		260,
		uintptr(unsafe.Pointer(&buf[0])),
		0)

	if r1 == 1 {
		return syscall.UTF16ToString(buf), nil
	}
	return "", err
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

// Version is app version
var Version string

func main() {
	app := gli.NewWith(&globalCmd{})
	app.Name = "wpchanger"
	app.Desc = "A commandline Wallpaper changer for windows"
	app.Version = Version
	app.Usage = `# change
wpchanger wallpaper.png
cat(or type) wallpaper.png | wpchanger

# get
wpchanger get original.png
`
	app.Copyright = "(C) 2018 Shuhei Kubota"
	app.SuppressErrorOutput = true
	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
