/*
Copyright 2021 The KubeEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package driver

import (
	"fmt"
	"os"
	"time"
	"unsafe"

	"k8s.io/klog/v2"

	"github.com/sailorvii/goav/avcodec"
	"github.com/sailorvii/goav/avformat"
	"github.com/sailorvii/goav/avutil"
	"github.com/sailorvii/goav/swscale"
)

var IfSaveFrame bool

// generate file name with current time. Formate f<year><month><day><hour><minute><second><millisecond>.<format>
func GenFileName(dir string, format string) string {
	return fmt.Sprintf("%s/f%s%s", dir, time.Now().Format("060102150405000"), format)
}

func save(frame *avutil.Frame, width int, height int, dir string, format string) {
	fileName := GenFileName(dir, format)
	file, err := os.Create(fileName)
	if err != nil {
		klog.Error("Error create file: ", fileName)
		return
	}
	defer file.Close()

	// Write header
	header := fmt.Sprintf("P6\n%d %d\n255\n", width, height)
	_, err = file.Write([]byte(header))
	if err != nil {
		klog.Error("Error write file: ", fileName)
		return
	}

	// Write pixel data
	for y := 0; y < height; y++ {
		data0 := avutil.Data(frame)[0]
		buf := make([]byte, width*3)
		startPos := uintptr(unsafe.Pointer(data0)) + uintptr(y)*uintptr(avutil.Linesize(frame)[0])
		for i := 0; i < width*3; i++ {
			element := *(*uint8)(unsafe.Pointer(startPos + uintptr(i)))
			buf[i] = element
		}
		_, err = file.Write(buf)
		if err != nil {
			klog.Error("Error write")
			return
		}
	}
}

func SaveFrame(input string, outDir string, format string) error {
	// Open video file
	pFormatContext := avformat.AvformatAllocContext()
	if avformat.AvformatOpenInput(&pFormatContext, input, nil, nil) != 0 {
		return fmt.Errorf("Unable to open file %s", input)
	}

	// Retrieve stream information
	if pFormatContext.AvformatFindStreamInfo(nil) < 0 {
		return fmt.Errorf("Couldn't find stream information")
	}

	// Dump information about file onto standard error
	pFormatContext.AvDumpFormat(0, input, 0)

	// Find the first video stream
	var i int
	for i = 0; i < int(pFormatContext.NbStreams()); i++ {
		if pFormatContext.Streams()[i].CodecParameters().AvCodecGetType() == avformat.AVMEDIA_TYPE_VIDEO {
			break
		}
	}
	if i == int(pFormatContext.NbStreams()) {
		return fmt.Errorf("couldn't find video stream")
	}

	// Get a pointer to the codec context for the video stream
	pCodecCtxOrig := pFormatContext.Streams()[i].Codec()
	// Find the decoder for the video stream
	pCodec := avcodec.AvcodecFindDecoder(avcodec.CodecId(pCodecCtxOrig.GetCodecId()))
	if pCodec == nil {
		return fmt.Errorf("unsupported codec")
	}
	// Copy context
	pCodecCtx := pCodec.AvcodecAllocContext3()
	if pCodecCtx.AvcodecCopyContext((*avcodec.Context)(unsafe.Pointer(pCodecCtxOrig))) != 0 {
		return fmt.Errorf("couldn't copy codec context")
	}

	// Open codec
	if pCodecCtx.AvcodecOpen2(pCodec, nil) < 0 {
		return fmt.Errorf("could not open codec")
	}

	// Allocate video frame
	pFrame := avutil.AvFrameAlloc()

	// Allocate an AVFrame structure
	pFrameRGB := avutil.AvFrameAlloc()
	if pFrameRGB == nil {
		return fmt.Errorf("unable to allocate RGB Frame")
	}

	// Determine required buffer size and allocate buffer
	numBytes := uintptr(avcodec.AvpictureGetSize(avcodec.AV_PIX_FMT_RGB24, pCodecCtx.Width(),
		pCodecCtx.Height()))
	buffer := avutil.AvMalloc(numBytes)

	// Assign appropriate parts of buffer to image planes in pFrameRGB
	// Note that pFrameRGB is an AVFrame, but AVFrame is a superset
	// of AVPicture
	avp := (*avcodec.Picture)(unsafe.Pointer(pFrameRGB))
	avp.AvpictureFill((*uint8)(buffer), avcodec.AV_PIX_FMT_RGB24, pCodecCtx.Width(), pCodecCtx.Height())

	// initialize SWS context for software scaling
	swsCtx := swscale.SwsGetcontext(
		pCodecCtx.Width(),
		pCodecCtx.Height(),
		(swscale.PixelFormat)(pCodecCtx.PixFmt()),
		pCodecCtx.Width(),
		pCodecCtx.Height(),
		avcodec.AV_PIX_FMT_RGB24,
		avcodec.SWS_BILINEAR,
		nil,
		nil,
		nil,
	)

	packet := avcodec.AvPacketAlloc()
	for {
		if !IfSaveFrame {
			time.Sleep(time.Second)
			continue
		}

		if pFormatContext.AvReadFrame(packet) <= 0 {
			klog.Error("Read frame failed")
			continue
		}

		// Is this a packet from the video stream?
		if packet.StreamIndex() == i {
			// Decode video frame
			response := pCodecCtx.AvcodecSendPacket(packet)
			if response < 0 {
				klog.Errorf("Error while sending a packet to the decoder: %s", avutil.ErrorFromCode(response))
			}
			for response >= 0 {
				response = pCodecCtx.AvcodecReceiveFrame((*avcodec.Frame)(unsafe.Pointer(pFrame)))
				if response == avutil.AvErrorEAGAIN || response == avutil.AvErrorEOF {
					break
				} else if response < 0 {
					klog.Errorf("Error while receiving a frame from the decoder: %s", avutil.ErrorFromCode(response))
				}

				// Convert the image from its native format to RGB
				swscale.SwsScale2(swsCtx, avutil.Data(pFrame),
					avutil.Linesize(pFrame), 0, pCodecCtx.Height(),
					avutil.Data(pFrameRGB), avutil.Linesize(pFrameRGB))

				// Save the frame to disk
				save(pFrameRGB, pCodecCtx.Width(), pCodecCtx.Height(), outDir, format)
			}
		}
	}
	/*
		// Free the packet that was allocated by av_read_frame
		packet.AvFreePacket()

		// Free the RGB image
		avutil.AvFree(buffer)
		avutil.AvFrameFree(pFrameRGB)

		// Free the YUV frame
		avutil.AvFrameFree(pFrame)

		// Close the codecs
		pCodecCtx.AvcodecClose()
		(*avcodec.Context)(unsafe.Pointer(pCodecCtxOrig)).AvcodecClose()

		// Close the video file
		pFormatContext.AvformatCloseInput()

		return nil
	*/
}
