#include "set_value.h"

extern "C" int set_value(MyDevice myDevice, const char* key, const char* value, char** err) {
	int ret = SUCCESS;
	std::string e = "";
	try {
		*err = (char*)malloc(e.length() + 1);
		if (NULL == *err) {
			std::cerr << "malloc() memory allocation failed!" << std::endl;
			return MALLAC_ERROR;
		}
		strcpy(*err, e.c_str());
		std::shared_ptr<GenApi::CNodeMapRef> nodemap = myDevice.dev->getRemoteNodeMap();
		ret = rcg::setString(nodemap, key, value, true);
	}
	catch (const std::invalid_argument& ex) {
		ret = INVALID_ARGUMENT;
		//std::cout << ex.what() << std::endl;
		std::string e = ex.what();
		*err = (char*)malloc(e.length() + 1);
		if (NULL == *err) {
			std::cerr << "malloc() memory allocation failed!" << std::endl;
			return MALLAC_ERROR;
		}
		strcpy(*err, ex.what());
	}
	catch (const std::exception& ex) {
		ret = EXCEPTION;
		//std::cout << ex.what() << std::endl;
		std::string e = ex.what();
		*err = (char*)malloc(e.length() + 1);
		if (NULL == *err) {
			std::cerr << "malloc() memory allocation failed!" << std::endl;
			return MALLAC_ERROR;
		}
		strcpy(*err, ex.what());
	}
	return ret;
}