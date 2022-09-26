#include "get_value.h"

extern "C" int get_value(MyDevice myDevice, const char* key, char** value, char** err) {
	int ret = SUCCESS;
	std::string e = "";
	std::string v = "";
	try {
		*err = (char*)malloc(e.length() + 1);
		strcpy(*err, e.c_str());
		if (NULL == *err) {
			std::cerr << "malloc() memory allocation failed!" << std::endl;
			return MALLAC_ERROR;
		}
		std::shared_ptr<GenApi::CNodeMapRef> nodemap = myDevice.dev->getRemoteNodeMap();
		v = rcg::getString(nodemap, key, true, true);
		if (v.size() != 0) {
			*value = (char*)malloc(v.length() + 1);
			if (NULL == *value) {
				std::cerr << "malloc() memory allocation failed!" << std::endl;
				return MALLAC_ERROR;
			}
			strcpy(*value, v.c_str());
		}
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

