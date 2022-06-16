#include "open_device.h"

extern "C" int open_device(MyDevice * myDevice, const char* device_serial_number, char** err) {
	int ret = SUCCESS;
	std::string e = "";
	*myDevice = MyDevice();
	(*myDevice).dev = rcg::getDevice(device_serial_number);
	if ((*myDevice).dev != 0) {
		try {
			*err = (char*)malloc(e.length() + 1);
			strcpy(*err, e.c_str());
			if (NULL == *err) {
				std::cerr << "malloc() memory allocation failed!" << std::endl;
				return MALLAC_ERROR;
			}
			(*myDevice).dev->open(rcg::Device::CONTROL);
		}
		catch (const std::exception& ex) {
			ret = EXCEPTION;
			std::string e = ex.what();
			*err = (char*)malloc(e.length() + 1);
			if (NULL == *err) {
				std::cerr << "malloc() memory allocation failed!" << std::endl;
				return MALLAC_ERROR;
			}
			strcpy(*err, ex.what());
			//std::cout << ex.what() << std::endl;
		}
	}
	else {
		ret = NO_DEVICE;
		//std::cout << "Cannot find device: " << device_serial_number << std::endl;
		std::string e = "Cannot find device: " + (std::string)device_serial_number;
		*err = (char*)malloc(e.length() + 1);
		if (NULL == *err) {
			std::cerr << "malloc() memory allocation failed!" << std::endl;
			return MALLAC_ERROR;
		}
		strcpy(*err, e.c_str());
	}
	return ret;
}
