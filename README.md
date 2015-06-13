# README #

### What is this repository for? ###

* A commandline Wallpaper changer for windows.

### How do I get set up? ###

* Download and go build

### Usage ###

    (in Explorer) D&D wallpaper file.

	#wpchanger wallpaper.jpg
		=> change the wallpaper to wallpaper.jpg
	#type wallppaper.jpg | wpchanger
	#cat wallppaper.jpg | wpchanger
		=> change the wallpaper to wallpaper.jpg
	#mojimg test1 test2  |  wpchanger
		=> change the wallpaper with [mojimg](https://bitbucket.org/shu/mojimg)

### Dependency ###

* github.com/andrew-d/go-termutil
	* isatty
