package driver

/*
#include <dlfcn.h>
#include <stdlib.h>
int open_device(unsigned int** device, char* deviceSN, char** error)
{
    void* handle;
    int result = -1;
    typedef int (*FPTR)(unsigned int**, char*, char**);
    handle = dlopen("../bin/librcapi.so", 1);
    if(handle == NULL){
        *error = (char *)dlerror();
        return result;
    }
    FPTR fptr = (FPTR)dlsym(handle, "open_device");
    if(fptr == NULL){
        *error = (char *)dlerror();
        result = -2;
    } else {
        result = (*fptr)(device, deviceSN, error);
    }
    dlclose(handle);
    return result;
}
int set_value (unsigned int* device, char* feature, char* value, char** error)
{
    void* handle;
    int result = -1;
    typedef int (*FPTR)(unsigned int*, char*, char*, char**);
    handle = dlopen("../bin/librcapi.so", 1);
    if(handle == NULL){
        *error = (char *)dlerror();
        return result;
    }
    FPTR fptr = (FPTR)dlsym(handle, "set_value");
    if(fptr == NULL){
        *error = (char *)dlerror();
        result = -2;
    } else {
        result = (*fptr)(device, feature, value, error);
    }
    dlclose(handle);
    return result;
}
int get_value (unsigned int* device, char* feature, char** value, char** error)
{
    void* handle;
    int result = -1;
    typedef int (*FPTR)(unsigned int*, char*, char**, char**);
	handle = dlopen("../bin/librcapi.so", 1);
    if(handle == NULL){
        *error = (char *)dlerror();
        return result;
    }
    FPTR fptr = (FPTR)dlsym(handle, "get_value");
    if(fptr == NULL){
        *error = (char *)dlerror();
        result = -2;
    } else {
        result = (*fptr)(device, feature, value, error);
    }
    dlclose(handle);
    return result;
}
int get_image (unsigned int* device, char* type, char** bufferPointer, int* size, char** error)
{
    void* handle;
    int result = -1;
    typedef int (*FPTR)(unsigned int*, char*, char**, int*, char**);
    handle = dlopen("../bin/librcapi.so", 1);
    if(handle == NULL){
        *error = (char *)dlerror();
        return result;
    }
    FPTR fptr = (FPTR)dlsym(handle, "get_image");
    if(fptr == NULL){
        *error = (char *)dlerror();
        result = -2;
    } else {
        result = (*fptr)(device, type, bufferPointer, size, error);
    }
    dlclose(handle);
    return result;
}
int close_device (unsigned int* device)
{
    void* handle;
    typedef void (*FPTR)(unsigned int*);
    handle = dlopen("../bin/librcapi.so", 1);
    if(handle == NULL){
        return -1;
    }
    FPTR fptr = (FPTR)dlsym(handle, "close_device");
    if(fptr == NULL){
	dlclose(handle);
        return -2;
    } else {
        (*fptr)(device);
    }
    dlclose(handle);
    return 0;
}
int free_image (char** bufferPointer)
{
    void* handle;
    typedef void (*FPTR)(char** bufferPointer);
    handle = dlopen("../bin/librcapi.so", 1);
    if(handle == NULL){
        return -1;
    }
    FPTR fptr = (FPTR)dlsym(handle, "free_image");
    if(fptr == NULL){
	dlclose(handle);
        return -2;
    } else {
        (*fptr)(bufferPointer);
    }
    dlclose(handle);
    return 0;
}
#cgo LDFLAGS: -ldl
*/
import "C"
import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"k8s.io/klog/v2"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

const (
	imageTriggerSingle   = "single"
	imageTriggerContinus = "continuous"
	imageTriggerStop     = "stop"
)

// Set is used to set device properitys
func (gigEClient *GigEVisionDevice) Set(DeviceSN string, value interface{}) (err error) {
	convertValue, err := gigEClient.convert(value)
	if err != nil {
		return err
	}
	switch gigEClient.deviceMeta[DeviceSN].FeatureName {
	case "ImageTrigger":
		switch convertValue {
		case imageTriggerSingle:
			//gigEClient.deviceMeta[DeviceSN].ImageTrigger = "single"  // move to the funcion postImage
		case imageTriggerContinus:
			//gigEClient.deviceMeta[DeviceSN].ImageTrigger = "continuous" // move to the funcion postImage
		case imageTriggerStop:
			gigEClient.deviceMeta[DeviceSN].ImageTrigger = imageTriggerStop
			gigEClient.deviceMeta[DeviceSN].ImagePostingFlag = false

		default:
			err = fmt.Errorf("set %s's ImageTrigger to %s failed, it only support  single, continuous, stop",
				DeviceSN, convertValue)
			return err
		}

		if convertValue != imageTriggerStop {
			if gigEClient.deviceMeta[DeviceSN].imageURL != "" {
				gigEClient.PostImage(DeviceSN, convertValue)
			} else {
				klog.V(4).Infof("Device %v's imageURL is null, so do not post the image", DeviceSN)
			}
		}
		return nil

	case "ImageFormat":
		switch convertValue {
		case "png":
			gigEClient.deviceMeta[DeviceSN].imageFormat = "png"
		case "pnm":
			gigEClient.deviceMeta[DeviceSN].imageFormat = "pnm"
		case "jpeg":
			gigEClient.deviceMeta[DeviceSN].imageFormat = "jpeg"
		default:
			err = fmt.Errorf("set %s's imageformat to %s failed, it only support format jpeg, png or pnm",
				DeviceSN, convertValue)
			return err
		}
	case "ImageURL":
		convertValue = strings.TrimSpace(convertValue)
		_, err := url.Parse(convertValue)
		if err != nil {
			err = fmt.Errorf("set imageURL failed because of incorrect format, message: %s", err)
			return err
		}
		gigEClient.deviceMeta[DeviceSN].imageURL = convertValue
		//gigEClient.PostImage(DeviceSN)
	default:
		var msg *C.char
		defer C.free(unsafe.Pointer(msg))
		signal := C.set_value(gigEClient.deviceMeta[DeviceSN].dev, C.CString(gigEClient.deviceMeta[DeviceSN].FeatureName), C.CString(convertValue), &msg)
		if signal != 0 {
			err = fmt.Errorf("set %s from device %s failed: %s", gigEClient.deviceMeta[DeviceSN].FeatureName, DeviceSN, (string)(C.GoString(msg)))
			if signal > 1 {
				gigEClient.deviceMeta[DeviceSN].deviceStatus = false
				go gigEClient.ReconnectDevice(DeviceSN)
			}
			return err
		}
	}
	return nil
}

// Get is used to get device properitys*/
func (gigEClient *GigEVisionDevice) Get(DeviceSN string) (results string, err error) {
	switch gigEClient.deviceMeta[DeviceSN].FeatureName {
	case "ImageTrigger":
		if gigEClient.deviceMeta[DeviceSN].ImageTrigger != "" {
			results = gigEClient.deviceMeta[DeviceSN].ImageTrigger
		} else {
			gigEClient.deviceMeta[DeviceSN].ImageTrigger = imageTriggerStop
			err = fmt.Errorf("maybe init %s's ImageTrigger failed, current value is  null", DeviceSN)
			return "", err
		}

	case "ImageFormat":
		if gigEClient.deviceMeta[DeviceSN].imageFormat != "" {
			results = gigEClient.deviceMeta[DeviceSN].imageFormat
		} else {
			err = fmt.Errorf("maybe init %s's image format failed, it only support format png or pnm", DeviceSN)
			return "", err
		}
	case "ImageURL":
		if gigEClient.deviceMeta[DeviceSN].imageURL != "" {
			results = gigEClient.deviceMeta[DeviceSN].imageURL
		} else {
			results = ""
		}

		return results, nil
	default:
		var msg *C.char
		var value *C.char
		defer C.free(unsafe.Pointer(msg))
		defer C.free(unsafe.Pointer(value))
		signal := C.get_value(gigEClient.deviceMeta[DeviceSN].dev, C.CString(gigEClient.deviceMeta[DeviceSN].FeatureName), &value, &msg)
		if signal != 0 {
			err = fmt.Errorf("get %s from device %s's failed: %s", gigEClient.deviceMeta[DeviceSN].FeatureName, DeviceSN, (string)(C.GoString(msg)))
			if signal > 1 {
				gigEClient.deviceMeta[DeviceSN].deviceStatus = false
				go gigEClient.ReconnectDevice(DeviceSN)
			}
			return "", err
		}
		results = C.GoString(value)
	}
	return results, err
}

// NewClient connect new device
func (gigEClient *GigEVisionDevice) NewClient(DeviceSN string) (err error) {
	var msg *C.char
	var dev *C.uint
	if gigEClient.deviceMeta == nil {
		gigEClient.deviceMeta = make(map[string]*DeviceMeta)
	}
	if DeviceSN == "" {
		err = fmt.Errorf("deviceSN can not be empty")
		return err
	}
	_, ok := gigEClient.deviceMeta[DeviceSN]
	if !ok {
		signal := C.open_device(&dev, C.CString(DeviceSN), &msg)
		if signal != 0 {
			klog.Errorf("Failed to open device %s: %s.", DeviceSN, (string)(C.GoString(msg)))
			gigEClient.deviceMeta[DeviceSN] = &DeviceMeta{
				dev:              nil,
				deviceStatus:     false,
				imageFormat:      "jpeg",
				imageURL:         "",
				ImageTrigger:     "stop",
				ImagePostingFlag: false,
				FeatureName:      "",
				maxRetryTimes:    100,
			}
			go gigEClient.ReconnectDevice(DeviceSN)
			return nil
		}
		gigEClient.deviceMeta[DeviceSN] = &DeviceMeta{
			dev:           dev,
			deviceStatus:  true,
			imageFormat:   "jpeg",
			imageURL:      "",
			FeatureName:   "",
			maxRetryTimes: 100,
		}
	}
	return nil
}

// PostImage is used to post images to dest url
func (gigEClient *GigEVisionDevice) PostImage(DeviceSN string, convertValue string) {
	var imageBuffer *byte
	var size int
	var p = &imageBuffer
	var msg *C.char
	defer C.free(unsafe.Pointer(msg))

	if gigEClient.deviceMeta[DeviceSN].ImagePostingFlag {
		klog.Errorf("image post is processing,do nothing,deviceSN %v", DeviceSN)
		return
	}

	signal := C.get_image(gigEClient.deviceMeta[DeviceSN].dev, C.CString(gigEClient.deviceMeta[DeviceSN].imageFormat), (**C.char)(unsafe.Pointer(p)), (*C.int)(unsafe.Pointer(&size)), &msg)
	if signal != 0 {
		klog.Errorf("Failed to get %s's images: %s.", DeviceSN, (string)(C.GoString(msg)))
		if signal > 1 {
			gigEClient.deviceMeta[DeviceSN].deviceStatus = false
			go gigEClient.ReconnectDevice(DeviceSN)
		}
		return
	}
	gigEClient.deviceMeta[DeviceSN].ImagePostingFlag = true

	go func() {
		var buffer []byte
		var bufferHdr = (*reflect.SliceHeader)(unsafe.Pointer(&buffer))
		defer C.free_image((**C.char)(unsafe.Pointer(&imageBuffer)))
		defer func() {
			gigEClient.deviceMeta[DeviceSN].ImagePostingFlag = false
		}()
		bufferHdr.Data = uintptr(unsafe.Pointer(imageBuffer))
		bufferHdr.Len = size
		bufferHdr.Cap = size
		postStr := base64.StdEncoding.EncodeToString(buffer)
		bufferSize := len(buffer)
		body := strings.NewReader(`{"size":` + strconv.Itoa(bufferSize) + `,"value":"` + postStr + `"}`)
		req, _ := http.NewRequest(http.MethodPost, gigEClient.deviceMeta[DeviceSN].imageURL, body)
		if req == nil {
			klog.Errorf("Failed to post %s's images: URL can't POST.", DeviceSN)
			return
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		client := &http.Client{}
		resp, _ := client.Do(req)
		if resp == nil {
			klog.Errorf("Failed to post %s's images: URL no reaction.", DeviceSN)
			return
		}

		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {

			}
		}(resp.Body)

		data, _ := ioutil.ReadAll(resp.Body)
		if resp.StatusCode == 200 {
			gigEClient.deviceMeta[DeviceSN].ImageTrigger = convertValue
		}
		fmt.Println("response Status:", resp.Status)
		fmt.Println("response Headers:", resp.Header)
		fmt.Println("response Body:", string(data))
	}()
}

func (gigEClient *GigEVisionDevice) convert(value interface{}) (convertValue string, err error) {
	switch value := value.(type) {
	case float64:
		convertValue = strconv.FormatFloat(value, 'f', -1, 64)
	case float32:
		convertValue = strconv.FormatFloat(float64(value), 'f', -1, 64)
	case int:
		convertValue = strconv.Itoa(value)
	case uint:
		convertValue = strconv.Itoa(int(value))
	case int8:
		convertValue = strconv.Itoa(int(value))
	case uint8:
		convertValue = strconv.Itoa(int(value))
	case int16:
		convertValue = strconv.Itoa(int(value))
	case uint16:
		convertValue = strconv.Itoa(int(value))
	case int32:
		convertValue = strconv.Itoa(int(value))
	case uint32:
		convertValue = strconv.Itoa(int(value))
	case int64:
		convertValue = strconv.FormatInt(value, 10)
	case uint64:
		convertValue = strconv.FormatUint(value, 10)
	case string:
		convertValue = value
	case bool:
		convertValue = strconv.FormatBool(value)
	case []byte:
		convertValue = string(value)
	default:
		err = fmt.Errorf("assertion is not supported for %v type", value)
		return "", err
	}
	return convertValue, nil
}
