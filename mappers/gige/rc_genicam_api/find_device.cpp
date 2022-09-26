#include "find_device.h"

extern "C" void find_device() {
	std::cout << "Available GigE Vision devices:" << std::endl;
	std::vector<std::shared_ptr<rcg::System> > system = rcg::System::getSystems();

	for (size_t i = 0; i < system.size(); i++) {
		system[i]->open();
		std::vector<std::shared_ptr<rcg::Interface> > interf = system[i]->getInterfaces();
		for (size_t k = 0; k < interf.size(); k++) {
			interf[k]->open();
			std::vector<std::shared_ptr<rcg::Device> > device = interf[k]->getDevices();
			for (size_t j = 0; j < device.size(); j++) {
				if (device[j]->getTLType() == "GEV") {
					std::cout << "  " << interf[k]->getID() << ":" << device[j]->getSerialNumber() << " (";
					std::string uname = device[j]->getDisplayName();
					if (uname.size() > 0) {
						std::cout << uname << ", ";
					}
					std::cout << device[j]->getID() << ")" << std::endl;
				}
			}
			interf[k]->close();
		}
		system[i]->close();
	}
	rcg::System::clearSystems();
}