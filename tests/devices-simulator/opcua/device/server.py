"""
Copyright 2021 The KubeEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
"""

from opcua import ua, uamethod, Server
import sys
import random
import time
from threading import Thread
import logging
import os

sys.path.insert(0, "..")


# 需要有的类型int32 bool,int32 int32, float double string

class Temperature(Thread):

    def __init__(self, switch, temperature, humidity, temperature_threshold, humidity_threshold, device_name):
        Thread.__init__(self)
        self._stop = False
        self.switch = switch
        self.temperature = temperature
        self.humidity = humidity
        self.temperature_threshold = temperature_threshold
        self.humidity_threshold = humidity_threshold
        self.device_name = device_name

    def stop(self):
        self._stop = True

    def run(self):
        while not self._stop:
            print("device is running...")
            value = random.uniform(-99, 99)
            self.temperature.set_value(value)
            print("temperature value is ", self.temperature.get_value())
            value = random.uniform(0, 100)
            self.humidity.set_value(value)
            print("humidity value is ", self.humidity.get_value())
            print("temperature threshold is ", self.temperature_threshold.get_value())
            print("humidity threshold is ", self.humidity_threshold.get_value())
            print("device_name is ", self.device_name.get_value())

            time.sleep(60)
        print("device has been stopped")


if __name__ == "__main__":

    logger = logging.getLogger()
    logger.setLevel(logging.WARNING)
    logging.basicConfig(stream=sys.stdout)

    server = Server()
    server.set_endpoint("opc.tcp://0.0.0.0:4840/freeopcua/server/")
    server.set_server_name("FreeOpcUa Example Server")

    # set all possible endpoint policies for clients to connect through
    server.set_security_policy([
        ua.SecurityPolicyType.Basic256Sha256_Sign,
        ua.SecurityPolicyType.Basic256Sha256_SignAndEncrypt,
    ])

    server.load_certificate("cert.pem")
    server.load_private_key("key.pem")

    server.user_manager.user_manager = lambda session, username,password: username == "testuser" and password == "testpass2"

    # setup our own namespace
    uri = "http://examples.freeopcua.github.io"
    idx = server.register_namespace(uri)

    # direct directly some objects and variables
    device = server.nodes.objects.add_object(idx, "device")

    device_switch = device.add_variable(idx, "switch", True)
    device_switch.set_writable()
    device_temperature = device.add_variable(idx, "temperature", 0)
    device_humidity = device.add_variable(idx, "humidity", 0)
    temperature_threshold = device.add_variable(idx, "temperature_threshold", 40)
    temperature_threshold.set_writable()
    humidity_threshold = device.add_variable(idx, "humidity_threshold", 60)
    humidity_threshold.set_writable()
    device_name = device.add_variable(idx, "device_name", "Huawei opcua simulator")

    server.start()
    print("Start opcua server...")

    device_thread = Temperature(device_switch, device_temperature, device_humidity, temperature_threshold,
                                humidity_threshold, device_name)
    device_thread.start()

    try:
        while True:
            time.sleep(1)
    finally:
        print("Exit opcua server...")
        device_thread.stop()
        server.stop()
