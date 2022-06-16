#pragma once
#include <iostream>
#include <rc_genicam_api/device.h>
#include <rc_genicam_api/config.h>
#include "const.h"

/**
  Returns 0 if the device propertie gets normally.

  Get the key attribute value of the device to value

  @param mydevice The MyDevice struct encapsulates a Genicam device
  @param key      The name of the device attribute
  @param value    The value of the device attribute
  @param err      The error infomation
  @return         The status of the function execution(0 is correct,
                  otherwise it is wrong)
*/
extern "C" int get_value(MyDevice myDevice, const char* key, char** value, char** err);
