#pragma once
#include <rc_genicam_api/system.h>
#include <rc_genicam_api/interface.h>
#include <rc_genicam_api/device.h>
#include <rc_genicam_api/stream.h>
#include <rc_genicam_api/buffer.h>
#include <rc_genicam_api/image.h>
#include <rc_genicam_api/image_store.h>
#include <rc_genicam_api/config.h>
#include <rc_genicam_api/pixel_formats.h>
#include <Base/GCException.h>
#include <signal.h>
#include <iostream>
#include <fstream>
#include <iomanip>
#include <rc_genicam_api/image_store.h>
#include <algorithm>
#include <atomic>
#include <thread>
#include <chrono>
#include <exception>
#include <iostream>
#include <sstream>
#include <iomanip>
#include <limits>

#include <rc_genicam_api/image.h>
#include <png.h>
#include "const.h"
#include <jpeglib.h>

/**
  Returns 0 if the device propertie gets normally.

  Grab a picture from a Genicam device

  @param mydevice     The MyDevice struct encapsulates a Genicam device
  @param imgfmt       The name of the device attribute
  @param image_buffer The value of the device attribute
  @param size         The length of the image_buffer
  @param err          The error infomation
  @return             The status of the function execution(0 is correct,
					  otherwise it is wrong)
*/
extern "C" int get_image(MyDevice myDevice, const char* imgfmt, unsigned char** image_buffer, int* size, char** err);
int storeBufferPNM(const rcg::Image& image, unsigned char** image_buffer, size_t yoffset, size_t height);
int storeBufferPNG(const rcg::Image& image, unsigned char** image_buffer, size_t yoffset, size_t height);
int storeBufferJPEG(const rcg::Image& image, unsigned char** image_buffer, size_t yoffset, size_t height);
extern "C" void free_image(unsigned char** image_buffer);
