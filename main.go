package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

var volume int8

func runCommand(name string, args ...string) (string, error) {
	log.Println(name, args)
	cmd := exec.Command(name, args...)
	var buffer bytes.Buffer
	cmd.Stdout = &buffer
	errorRunCommand := cmd.Run()
	return buffer.String(), errorRunCommand
}

func volumeOffset(flag int8) int8 {
	volumeChange := volume + flag
	if volumeChange < 0 || volumeChange > 100 {
		return 0
	}
	return flag
}

func volumeControl(context *gin.Context) {
	updateCurrentVolume()

	var stdOut string
	var errorMessage error

	control := context.Param("control")
	switch control {
	case "up":
		volume += volumeOffset(5)
		break
	case "down":
		volume += volumeOffset(-5)
		break
	default:
		volumeFromParam, errorConvertVolumeFromParam := strconv.Atoi(control)
		if errorConvertVolumeFromParam != nil {
			context.JSON(http.StatusInternalServerError, gin.H{
				"stdout": stdOut,
				"error":  errorMessage,
			})
			return
		}
		volume = int8(volumeFromParam)
	}

	switch runtime.GOOS {
	case "darwin":
		stdOut, errorMessage = runCommand("osascript", "-e", fmt.Sprintf("set volume output volume %d", volume))
		break
	case "linux":
		stdOut, errorMessage = runCommand("sh", "script/linux/set_volume.sh", fmt.Sprintf("%d%%", volume))
		log.Println(stdOut, errorMessage)
		break
	}

	context.JSON(http.StatusOK, gin.H{
		"stdout": stdOut,
		"error":  errorMessage,
	})
}

func updateCurrentVolume() {
	var stdOut string
	var errorGetCurrentVolume error

	switch runtime.GOOS {
	case "darwin":
		stdOut, errorGetCurrentVolume = runCommand("osascript", "-e set ovol to output volume of (get volume settings)")
		break
	case "linux":
		stdOut, errorGetCurrentVolume = runCommand("sh", "script/linux/get_current_volume.sh")
		break
	}

	if errorGetCurrentVolume != nil {
		log.Fatalln(errorGetCurrentVolume)
	}
	currentVolumeTrim := strings.TrimSpace(stdOut)
	currentVolume, errorConvertCurrentVolumeToInt := strconv.Atoi(currentVolumeTrim)
	if errorConvertCurrentVolumeToInt != nil {
		log.Fatalln(errorConvertCurrentVolumeToInt)
	}
	volume = int8(currentVolume)
}

func main() {
	updateCurrentVolume()
	route := gin.Default()
	route.GET("volume/:control", volumeControl)
	route.LoadHTMLGlob("*.tmpl")
	route.GET("volume", func(context *gin.Context) {
		updateCurrentVolume()
		context.HTML(http.StatusOK, "index.tmpl", gin.H{
			"current_volume": volume,
		})
	})
	route.Run(":44110")
}
