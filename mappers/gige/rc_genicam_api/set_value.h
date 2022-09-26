#pragma once
#include <iostream>
#include <rc_genicam_api/device.h>
#include <rc_genicam_api/config.h>
#include "const.h"

/**
  Returns 0 if the device properties are set normally.

  Set the key attribute of the device to value

  @param mydevice The MyDevice struct encapsulates a Genicam device
  @param key      The name of the device attribute
  @param value    The value of the device attribute
  @param err      The error infomation
  @return         The status of the function execution(0 is correct,
				  otherwise it is wrong)
*/
extern "C" int set_value(MyDevice myDevice, const char* key, const char* value, char** err);