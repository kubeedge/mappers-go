#pragma once
#include <rc_genicam_api/device.h>

#define MALLAC_ERROR -1
#define SUCCESS 0
#define INVALID_ARGUMENT 1
#define NO_IMAGES 1
#define NO_DEVICE 2
#define NO_STREAMS 2
#define IOEXCEPTION 3
#define GEN_EXCEPTION 4
#define EXCEPTION 5

/**
  The MyDevice struct encapsulates a Genicam device.
*/
struct MyDevice
{
	std::shared_ptr<rcg::Device> dev;
};
