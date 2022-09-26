#include <iostream>
#include <time.h>
#include "find_device.h"
#include "get_value.h"
#include "set_value.h"
#include "get_image.h"
#include "close_device.h"
#include "open_device.h"
#include "const.h"

using namespace std;

int main() {
	time_t a, b;
	find_device();
	std::string key_value = "";
	const char* device_id = "";
	char* err = NULL;
	int ret;
	MyDevice myDevice;
	cout << "device_id:";
	cin >> key_value;
	device_id = key_value.c_str();
	ret = open_device(&myDevice, device_id, &err);
	std::cout << ret << " " << err << endl;
	free(err);
	while (1) {
		unsigned char* image_buffer = NULL;
		int size = 0;
		cout << "key_value: ";
		cin >> key_value;
		try {
			if (key_value == "exit") {
				break;
			}
			else if (key_value == "png" || key_value == "pnm" || key_value == "jpeg") {
				a = clock();
				const char* imgfmt = key_value.c_str();
				ret = get_image(myDevice, imgfmt, &image_buffer, &size, &err);
				b = clock();
				std::cout << "get image time:" << b - a << "  " << (double)((b - a) / CLOCKS_PER_SEC) << std::endl;
				cout << "get_image errCode:" << ret << endl;
				cout << "get_image err:" << err << endl;
				free(err);
				err = NULL;
				cout << "The size of image:" << size << endl;
				unsigned char* image_free = image_buffer;
				std::string name = "test." + key_value;
				std::ofstream out(name, std::ios::binary);
				std::streambuf* sb = out.rdbuf();
				for (int i = 0; i < size && out.good(); i++, image_buffer++) {
					sb->sputc(*image_buffer);
				}
				out.close();
				free_image(&image_free);
			}
			else {
				size_t k = key_value.find('=');
				char* value = NULL;
				ret = get_value(myDevice, key_value.substr(0, k).c_str(), &value, &err);
				if (value != nullptr) {
					std::cout << key_value.substr(0, k).c_str() << ": " << value << endl;
				}
				cout << "get_value errCode:" << ret << endl;
				cout << "get_value err:" << err << endl;
				free(err);
				free(value);
				err = NULL;
				value = NULL;
				ret = set_value(myDevice, key_value.substr(0, k).c_str(), key_value.substr(k + 1).c_str(), &err);
				cout << "set_value errCode:" << ret << endl;
				cout << "set_value err:" << err << endl;
				free(err);
				err = NULL;
				ret = get_value(myDevice, key_value.substr(0, k).c_str(), &value, &err);
				if (value != nullptr) {
					std::cout << key_value.substr(0, k).c_str() << ": " << value << endl;
				}
				cout << "get_value errCode:" << ret << endl;
				cout << "get_value err:" << err << endl;
				free(err);
				free(value);
				err = NULL;
				value = NULL;
			}
		}
		catch (const std::exception& ex) {
			std::cout << "Exception: " << ex.what() << std::endl;
		}
		catch (const GENICAM_NAMESPACE::GenericException& ex) {
			std::cout << "Exception: " << ex.what() << std::endl;
		}
		catch (...) {
			std::cout << "Unknown exception!" << std::endl;
		}

	}
	close_device(myDevice);
	rcg::System::clearSystems();

	return 0;
}