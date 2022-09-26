#pragma once
#include <iostream>
#include <rc_genicam_api/device.h>
#include <rc_genicam_api/config.h>
#include "const.h"

/**
  Returns 0 if the device turns on normally.

  @param mydevice  The MyDevice struct encapsulates a Genicam device
  @param device_SN The serial number of a Genicam device
  @param err       The error infomation
  @return          The status of the function execution(0 is correct,
				   otherwise it is wrong)
*/
extern "C" int open_device(MyDevice * myDevice, const char* device_SN, char** err);
