#pragma once
#include <rc_genicam_api/device.h>
#include "const.h"

/**
  Returns 0 if the device turns off normally.

  @param mydevice  The MyDevice struct encapsulates a Genicam device
  @return          The status of the function execution(0 is correct,
				   otherwise it is wrong)
*/
extern "C" int close_device(MyDevice myDevice);
