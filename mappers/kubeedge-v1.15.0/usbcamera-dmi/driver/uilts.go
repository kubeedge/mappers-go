package driver

import (
	"encoding/base64"
	"fmt"
	"github.com/blackjack/webcam"
)

func SetControl(controls map[webcam.ControlID]webcam.Control, featureName string, data interface{}) error {
	for controlID, control := range controls {
		if control.Name == featureName {
			err := Cam.SetControl(controlID, int32(data.(int64)))
			if err != nil {
				return err
			}
			return nil
		}
	}
	err := fmt.Errorf("set control:%s not found the parament", featureName)
	return err
}

func GetControl(controls map[webcam.ControlID]webcam.Control, featureName string) (int32, error) {
	for controlID, control := range controls {
		if control.Name == featureName {
			value, err := Cam.GetControl(controlID)
			if err != nil {
				return -1, err
			}
			return value, nil
		}
	}
	err := fmt.Errorf("get control:%s not found the parament", featureName)
	return -1, err
}

var format webcam.PixelFormat = 0

func GetImage(width, height, pixelFormat int) (string, error) {
	// 读取摄像头的参数
	if format == 0 {
		format_ := webcam.PixelFormat(pixelFormat)
		if result, _, _, err := Cam.SetImageFormat(format_, uint32(width), uint32(height)); err != nil {
			return "", fmt.Errorf("error setting format:%v", err)
		} else {
			format = result
		}
	}
	if err := Cam.StartStreaming(); err != nil {
		if err.Error() != "Already streaming" {
			return "", fmt.Errorf("error starting streaming:%v", err)
		}
	}
	frame, err := Cam.ReadFrame()
	if err != nil {
		return "", fmt.Errorf("error reading frame:%v", err)
	}
	result := base64.StdEncoding.EncodeToString(frame)
	return result, nil
}
