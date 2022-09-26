#include "close_device.h"

extern "C" int close_device(MyDevice myDevice) {
	int ret = SUCCESS;
	if (myDevice.dev != nullptr) {
		myDevice.dev->close();
	}
	else {
		ret = 1;
	}
	return ret;
}