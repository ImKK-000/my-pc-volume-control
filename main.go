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
	cmd := exec.Command(name, args...)
	var buffer bytes.Buffer
	cmd.Stdout = &buffer
	errorRunCommand := cmd.Run()
	return buffer.String(), errorRunCommand
}

func volumeOffset(flag int8) int8 {
	if volume+flag < 0 || volume+flag > 100 {
		return 0
	}
	return flag
}

func volumeControl(context *gin.Context) {
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
	}

	switch runtime.GOOS {
	case "darwin":
		stdOut, errorMessage = runCommand("osascript", "-e", fmt.Sprintf("set volume output volume %d", volume))
		break
	case "linux":
		stdOut, errorMessage = runCommand("pactl", "--", "set-sink-volume", "0", fmt.Sprintf("%d%%", volume))
	}

	context.JSON(http.StatusOK, gin.H{
		"stdout": stdOut,
		"error":  errorMessage,
	})
}

func main() {
	stdOut, errorGetCurrentVolume := runCommand("osascript", "-e", "set ovol to output volume of (get volume settings)")
	if errorGetCurrentVolume != nil {
		log.Fatalln(errorGetCurrentVolume)
	}
	currentVolumeTrim := strings.TrimSpace(stdOut)
	currentVolume, errorConvertCurrentVolumeToInt := strconv.Atoi(currentVolumeTrim)
	if errorConvertCurrentVolumeToInt != nil {
		log.Fatalln(errorConvertCurrentVolumeToInt)
	}
	volume = int8(currentVolume)

	server := gin.Default()
	server.GET("volume/:control", volumeControl)
	server.Run(":44110")
}
